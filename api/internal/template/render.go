package template

import (
	"fmt"
	"regexp"
	"strings"
)

var varPattern = regexp.MustCompile(`\{\{([^}]+)\}\}`)

// Render substitutes {{variable}} and {{nested.path}} placeholders using payload values.
func Render(body string, payload map[string]any) string {
	if body == "" {
		return ""
	}

	return varPattern.ReplaceAllStringFunc(body, func(match string) string {
		path := strings.TrimSpace(varPattern.FindStringSubmatch(match)[1])
		val := resolvePath(payload, path)
		if val == nil {
			return match
		}
		return fmt.Sprint(val)
	})
}

func resolvePath(data map[string]any, path string) any {
	parts := strings.Split(path, ".")
	var current any = data

	for _, key := range parts {
		key = strings.TrimSpace(key)
		if key == "" {
			return nil
		}

		m, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		v, ok := m[key]
		if !ok {
			return nil
		}
		current = v
	}

	return current
}
