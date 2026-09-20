package embed

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureAssetsDownloads(t *testing.T) {
	t.Setenv("CONDUIT_EMBEDDER", "")
	t.Setenv("CONDUIT_SKIP_EMBED_DOWNLOAD", "")
	model := []byte("fake-onnx")
	libName := runtimeLibNames()[0]
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	body := []byte("ortlib")
	if err := tw.WriteHeader(&tar.Header{Name: "lib/" + libName, Mode: 0o755, Size: int64(len(body))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	ort := buf.Bytes()
	mux := http.NewServeMux()
	mux.HandleFunc("/model.onnx", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(model) })
	mux.HandleFunc("/tokenizer.json", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"model":{"vocab":{"[CLS]":1}}}`))
	})
	mux.HandleFunc("/ort.tgz", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(ort) })
	srv := httptest.NewServer(mux)
	defer srv.Close()
	dir := t.TempDir()
	st, err := EnsureAssets(context.Background(), AssetOpts{
		DataDir:      dir,
		ModelURL:     srv.URL + "/model.onnx",
		TokenizerURL: srv.URL + "/tokenizer.json",
		RuntimeURL:   srv.URL + "/ort.tgz",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !st.ModelOK || !st.RuntimeOK {
		t.Fatalf("%+v", st)
	}
	got, err := os.ReadFile(filepath.Join(dir, "models", "minilm.onnx"))
	if err != nil || string(got) != "fake-onnx" {
		t.Fatalf("model %q %v", got, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "models", "tokenizer.json")); err != nil {
		t.Fatal("tokenizer missing")
	}
	if err := RemoveDownloadedAssets(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "models")); !os.IsNotExist(err) {
		t.Fatal("models should be gone after uninstall cleanup")
	}
}

func TestEnsureAssetsSkipFake(t *testing.T) {
	t.Setenv("CONDUIT_EMBEDDER", "fake")
	st, err := EnsureAssets(context.Background(), AssetOpts{DataDir: t.TempDir(), ModelURL: "http://127.0.0.1:1/nope"})
	if err != nil {
		t.Fatal(err)
	}
	if !st.Skipped {
		t.Fatalf("%+v", st)
	}
}
