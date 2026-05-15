package workflow

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&WorkflowEntity{}, &WorkflowVersionEntity{}))
	return db
}

func testNodesEdges() ([]byte, []byte) {
	nodes, _ := json.Marshal([]map[string]string{{"id": "n1", "type": "trigger"}})
	edges, _ := json.Marshal([]map[string]string{{"from": "n1", "to": "n2"}})
	return nodes, edges
}

func TestSaveCreatesVersion(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	nodes, edges := testNodesEdges()
	saved, err := repo.Save(ctx, SaveInput{
		Name:    "Test",
		Nodes:   nodes,
		Edges:   edges,
		Filters: map[string]string{},
	})
	require.NoError(t, err)

	versions, err := repo.ListVersions(ctx, saved.ID)
	require.NoError(t, err)
	require.Len(t, versions, 1)
	require.Equal(t, 1, versions[0].VersionNumber)
	require.Equal(t, VersionSourceSave, versions[0].Source)
}

func TestSaveSkipsDuplicateVersion(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	nodes, edges := testNodesEdges()
	saved, err := repo.Save(ctx, SaveInput{
		ID:      "",
		Name:    "Test",
		Nodes:   nodes,
		Edges:   edges,
		Filters: map[string]string{},
	})
	require.NoError(t, err)

	_, err = repo.Save(ctx, SaveInput{
		ID:      saved.ID,
		Name:    "Test",
		Nodes:   nodes,
		Edges:   edges,
		Filters: map[string]string{},
	})
	require.NoError(t, err)

	versions, err := repo.ListVersions(ctx, saved.ID)
	require.NoError(t, err)
	require.Len(t, versions, 1)
}

func TestSaveCreatesVersionOnContentChange(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	nodes, edges := testNodesEdges()
	saved, err := repo.Save(ctx, SaveInput{
		Name:    "Test",
		Nodes:   nodes,
		Edges:   edges,
		Filters: map[string]string{},
	})
	require.NoError(t, err)

	nodes2, _ := json.Marshal([]map[string]string{{"id": "n2", "type": "action"}})
	_, err = repo.Save(ctx, SaveInput{
		ID:      saved.ID,
		Name:    "Test Updated",
		Nodes:   nodes2,
		Edges:   edges,
		Filters: map[string]string{},
	})
	require.NoError(t, err)

	versions, err := repo.ListVersions(ctx, saved.ID)
	require.NoError(t, err)
	require.Len(t, versions, 2)
}

func TestRestoreVersion(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	nodes, edges := testNodesEdges()
	saved, err := repo.Save(ctx, SaveInput{
		Name:    "V1",
		Nodes:   nodes,
		Edges:   edges,
		Filters: map[string]string{},
	})
	require.NoError(t, err)

	versions, err := repo.ListVersions(ctx, saved.ID)
	require.NoError(t, err)
	firstVersionID := versions[0].ID

	nodes2, _ := json.Marshal([]map[string]string{{"id": "n2"}})
	_, err = repo.Save(ctx, SaveInput{
		ID:      saved.ID,
		Name:    "V2",
		Nodes:   nodes2,
		Edges:   edges,
		Filters: map[string]string{},
	})
	require.NoError(t, err)

	restored, err := repo.RestoreVersion(ctx, saved.ID, firstVersionID)
	require.NoError(t, err)
	require.Equal(t, "V1", restored.Name)

	versions, err = repo.ListVersions(ctx, saved.ID)
	require.NoError(t, err)
	require.Len(t, versions, 3)
	require.Equal(t, VersionSourceRestore, versions[0].Source)
}

func TestDeleteWorkflowRemovesVersions(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	nodes, edges := testNodesEdges()
	saved, err := repo.Save(ctx, SaveInput{
		Name:  "To Delete",
		Nodes: nodes,
		Edges: edges,
	})
	require.NoError(t, err)

	require.NoError(t, repo.Delete(ctx, saved.ID))

	var count int64
	require.NoError(t, db.Model(&WorkflowVersionEntity{}).Where("workflow_id = ?", saved.ID).Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestComputeContentHashStable(t *testing.T) {
	nodes := []byte(`[{"id":"a"}]`)
	edges := []byte(`[]`)
	filters := map[string]string{"b": "2", "a": "1"}

	h1 := ComputeContentHash(nodes, edges, filters)
	h2 := ComputeContentHash(nodes, edges, map[string]string{"a": "1", "b": "2"})
	require.Equal(t, h1, h2)
}
