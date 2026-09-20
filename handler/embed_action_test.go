package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/michmich112/conduit-plugin/embed"
)

func TestTestEmbedPersistsFingerprint(t *testing.T) {
	t.Setenv("CONDUIT_EMBEDDER", "")
	dir := t.TempDir()
	h := New(dir, embed.Selection{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = body
		vec := make([]float64, embed.DefaultDim)
		vec[0] = 1
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []any{map[string]any{"embedding": vec}},
		})
	}))
	defer srv.Close()
	raw, err := json.Marshal(map[string]any{
		"url": srv.URL, "model": "minilm", "api_key": "secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := h.AdminAction(t.Context(), "test_embed", raw)
	if err != nil {
		t.Fatal(err)
	}
	var res map[string]any
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatal(err)
	}
	if res["ok"] != true {
		t.Fatalf("%s", out)
	}
	sec, err := loadSecretsFile(dir)
	if err != nil {
		t.Fatal(err)
	}
	if sec.EmbedHTTPAPIKey != "secret" {
		t.Fatalf("key %q", sec.EmbedHTTPAPIKey)
	}
	if sec.EmbedHTTPFingerprint != embed.HTTPFingerprint(srv.URL, "minilm", "secret", embed.DefaultDim) {
		t.Fatalf("fp %q", sec.EmbedHTTPFingerprint)
	}
	if _, err := os.Stat(filepath.Join(dir, "secrets.json")); err != nil {
		t.Fatal(err)
	}
}
