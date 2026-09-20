package embed

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPEmbedOpenAIShape(t *testing.T) {
	t.Setenv("CONDUIT_EMBEDDER", "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/embeddings" {
			t.Fatalf("path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer sk-test" {
			t.Fatalf("auth %q", r.Header.Get("Authorization"))
		}
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatal(err)
		}
		if req["dimensions"].(float64) != float64(DefaultDim) {
			t.Fatalf("dimensions %+v", req["dimensions"])
		}
		vec := make([]float64, DefaultDim)
		vec[0] = 1
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []any{map[string]any{"embedding": vec}},
		})
	}))
	defer srv.Close()

	e := NewHTTP(srv.URL, "minilm-compat", "sk-test", 0)
	v, err := e.Embed(t.Context(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if len(v) != DefaultDim {
		t.Fatalf("dim %d", len(v))
	}
}

func TestHTTPEmbedRejectsWrongDim(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []any{map[string]any{"embedding": []float64{0.1, 0.2, 0.3}}},
		})
	}))
	defer srv.Close()
	e := NewHTTP(srv.URL+"/v1/embeddings", "big-model", "", 0)
	_, err := e.Embed(t.Context(), "hello")
	if err == nil || !strings.Contains(err.Error(), "dim 3") {
		t.Fatalf("got %v", err)
	}
}

func TestSelectHTTPSkipsONNXWhenVerified(t *testing.T) {
	t.Setenv("CONDUIT_EMBEDDER", "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vec := make([]float64, DefaultDim)
		vec[1] = 1
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []any{map[string]any{"embedding": vec}},
		})
	}))
	defer srv.Close()
	url := srv.URL
	model := "all-minilm"
	key := "k"
	s := SelectWith(SelectOpts{
		ModelPath:        "missing.onnx",
		Provider:         "http",
		HTTPURL:          url,
		HTTPModel:        model,
		HTTPKey:          key,
		SavedFingerprint: HTTPFingerprint(url, model, key, DefaultDim),
	})
	if s.Source != SourceHTTP || s.Embedder == nil {
		t.Fatalf("%+v", s)
	}
	if err := Probe(t.Context(), s.Embedder); err != nil {
		t.Fatal(err)
	}
}

func TestSelectHTTPUnverifiedDoesNotUseHTTP(t *testing.T) {
	t.Setenv("CONDUIT_EMBEDDER", "")
	s := SelectWith(SelectOpts{
		ModelPath:        "missing.onnx",
		Provider:         "http",
		HTTPURL:          "http://127.0.0.1:9",
		HTTPModel:        "x",
		HTTPKey:          "k",
		SavedFingerprint: "not-a-match",
	})
	if s.Source != SourceUnavailable || s.Embedder != nil {
		t.Fatalf("unverified http must not skip missing onnx: %+v", s)
	}
}

func TestNormalizeEmbeddingsURL(t *testing.T) {
	if got := normalizeEmbeddingsURL("https://api.openai.com/v1"); got != "https://api.openai.com/v1/embeddings" {
		t.Fatal(got)
	}
	if got := normalizeEmbeddingsURL("https://ollama.local/v1/embeddings/"); got != "https://ollama.local/v1/embeddings" {
		t.Fatal(got)
	}
}
