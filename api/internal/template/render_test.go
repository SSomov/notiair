package template

import "testing"

func TestRenderSimple(t *testing.T) {
	body := "Hello {{name}}!"
	payload := map[string]any{"name": "World"}
	got := Render(body, payload)
	want := "Hello World!"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRenderNested(t *testing.T) {
	body := "{{user.email}}"
	payload := map[string]any{
		"user": map[string]any{"email": "a@b.c"},
	}
	got := Render(body, payload)
	if got != "a@b.c" {
		t.Fatalf("got %q", got)
	}
}

func TestRenderMissing(t *testing.T) {
	body := "x {{missing}} y"
	got := Render(body, map[string]any{})
	if got != "x {{missing}} y" {
		t.Fatalf("got %q", got)
	}
}
