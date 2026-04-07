package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"gorm.io/gorm"

	"notiair/handlers"
	"notiair/internal/config"
	"notiair/internal/persistence/channel"
	"notiair/internal/persistence/database"
	"notiair/internal/persistence/outbox"
	"notiair/internal/persistence/serviceconfig"
	workflowpersistence "notiair/internal/persistence/workflow"
	"notiair/internal/queue"
	"notiair/internal/routing"
	"notiair/internal/stream"
	"notiair/internal/templates"
	"notiair/internal/workflow"
	"notiair/routes"
	"notiair/services"
)

var (
	appConfig         config.Config
	dbConn            *gorm.DB
	queueClient       queue.Client
	serviceConfigRepo serviceconfig.Repository
	streamConsumer    *stream.Consumer
	streamHub         *stream.Hub
	redisStore        *stream.RedisStore
)

func initConfig() {
	var err error
	appConfig, err = config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
}

func initDatabase() {
	var err error
	dbConn, err = database.Connect(appConfig.DB)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}

	serviceConfigRepo = serviceconfig.NewRepository(dbConn)

	if err := dbConn.AutoMigrate(&outbox.Message{}, &serviceconfig.ServiceConfig{}, &channel.Channel{}, &workflowpersistence.WorkflowEntity{}); err != nil {
		log.Fatalf("migrate db: %v", err)
	}

	if err := seedServiceConfigs(context.Background()); err != nil {
		log.Fatalf("seed service configs: %v", err)
	}

	log.Println("database initialized successfully")
}

func initQueue() {
	queueClient = queue.NewAsynqClient(appConfig.Queue)
}

func seedServiceConfigs(ctx context.Context) error {
	if _, err := serviceConfigRepo.EnsureDefault(ctx, serviceconfig.TypeTelegram); err != nil {
		return err
	}
	if _, err := serviceConfigRepo.EnsureDefault(ctx, serviceconfig.TypeDefault); err != nil {
		return err
	}
	return nil
}

func buildApplication() *fiber.App {
	templateRepo := templates.NewMemoryRepository()
	workflowPersistenceRepo := workflowpersistence.NewRepository(dbConn)
	workflowRepo := workflow.NewDBRepository(workflowPersistenceRepo)
	routerSvc := routing.NewService(workflowRepo)
	outboxRepo := outbox.NewRepository(dbConn)

	notificationService := services.NewNotificationService(routerSvc, queueClient, outboxRepo)
	queueInspector := queue.NewNoopInspector()
	channelRepo := channel.NewRepository(dbConn)
	streamConfig := handlers.StreamConfig{
		Brokers: appConfig.Stream.Brokers,
		Topic:   appConfig.Stream.Topic,
	}
	
	// Создаем Redis store для хранения сообщений
	if redisStore == nil {
		var err error
		redisStore, err = stream.NewRedisStore(appConfig.Redis.URL)
		if err != nil {
			log.Printf("failed to create redis store: %v", err)
			// Продолжаем работу без Redis, но функциональность будет ограничена
		}
	}
	
	// Создаем и запускаем WebSocket hub
	if streamHub == nil {
		streamHub = stream.NewHub(redisStore)
		go streamHub.Run()
	}
	
	apiHandlers := handlers.NewAPI(notificationService, templateRepo, workflowRepo, queueInspector, serviceConfigRepo, channelRepo, streamConfig, streamHub, redisStore)

	app := fiber.New(fiber.Config{
		AppName:      "NotiAir Notification API",
		ServerHeader: "Fiber",
		ReadTimeout:  5 * time.Second,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: false,
	}))
	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(logger.New())

	routes.New(apiHandlers).Register(app.Group("/api/v1"))

	return app
}

func initStreamConsumer() error {
	workflowPersistenceRepo := workflowpersistence.NewRepository(dbConn)
	workflowRepo := workflow.NewDBRepository(workflowPersistenceRepo)
	routerSvc := routing.NewService(workflowRepo)
	outboxRepo := outbox.NewRepository(dbConn)
	notificationService := services.NewNotificationService(routerSvc, queueClient, outboxRepo)

	// Создаем Redis store если еще не создан
	if redisStore == nil {
		var err error
		redisStore, err = stream.NewRedisStore(appConfig.Redis.URL)
		if err != nil {
			log.Printf("failed to create redis store: %v", err)
			// Продолжаем работу без Redis
		}
	}

	// Создаем hub если еще не создан
	if streamHub == nil {
		streamHub = stream.NewHub(redisStore)
		go streamHub.Run()
	}

	var err error
	streamConsumer, err = stream.NewConsumer(
		appConfig.Stream.Brokers,
		appConfig.Stream.Topic,
		appConfig.Stream.GroupID,
		workflowRepo,
		notificationService,
		streamHub,
		redisStore,
	)
	if err != nil {
		return err
	}

	return nil
}

func runServer(app *fiber.App) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Запускаем stream consumer
	if streamConsumer != nil {
		if err := streamConsumer.Start(ctx); err != nil {
			log.Fatalf("failed to start stream consumer: %v", err)
		}
		defer func() {
			if err := streamConsumer.Stop(); err != nil {
				log.Printf("failed to stop stream consumer: %v", err)
			}
		}()
	}

	go func() {
		log.Printf("server starting on %s", appConfig.HTTP.Addr)
		if err := app.Listen(appConfig.HTTP.Addr); err != nil {
			log.Fatalf("fiber listen: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		if err := app.Shutdown(); err != nil {
			log.Printf("fiber shutdown: %v", err)
		}
		close(done)
	}()

	select {
	case <-done:
		log.Println("server shut down gracefully")
	case <-shutdownCtx.Done():
		log.Println("server shutdown timed out")
	}
}

func main() {
	initConfig()
	initDatabase()
	initQueue()
	defer queueClient.Close()

	if err := initStreamConsumer(); err != nil {
		log.Fatalf("failed to init stream consumer: %v", err)
	}

	app := buildApplication()
	runServer(app)
}
