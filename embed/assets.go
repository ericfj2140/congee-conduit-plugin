package embed

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	DefaultModelURL = "https://huggingface.co/sentence-transformers/all-MiniLM-L6-v2/resolve/main/onnx/model.onnx"
	defaultORTVer   = "1.19.2"
)

// AssetOpts locates MiniLM ONNX and the onnxruntime shared library.
type AssetOpts struct {
	DataDir    string
	ModelURL   string
	RuntimeURL string
	Force      bool
}

// AssetStatus is written to data/assets-status.json and returned to the UI.
type AssetStatus struct {
	ModelOK     bool   `json:"model_ok"`
	RuntimeOK   bool   `json:"runtime_ok"`
	ModelPath   string `json:"model_path,omitempty"`
	RuntimePath string `json:"runtime_path,omitempty"`
	ModelURL    string `json:"model_url,omitempty"`
	RuntimeURL  string `json:"runtime_url,omitempty"`
	Error       string `json:"error,omitempty"`
	Skipped     bool   `json:"skipped,omitempty"`
}

var assetMu sync.Mutex

func skipAssetDownload() bool {
	v := strings.TrimSpace(os.Getenv("CONDUIT_EMBEDDER"))
	if v == "fake" {
		return true
	}
	return os.Getenv("CONDUIT_SKIP_EMBED_DOWNLOAD") == "1"
}

// DefaultRuntimeURL is the Microsoft onnxruntime tarball for this GOOS/GOARCH.
func DefaultRuntimeURL() string {
	osName, arch := ortArchiveParts()
	if osName == "" {
		return ""
	}
	name := "onnxruntime-" + osName + "-" + arch + "-" + defaultORTVer
	return "https://github.com/microsoft/onnxruntime/releases/download/v" + defaultORTVer + "/" + name + ".tgz"
}

func ortArchiveParts() (osName, arch string) {
	switch runtime.GOOS {
	case "darwin":
		osName = "osx"
	case "linux":
		osName = "linux"
	default:
		return "", ""
	}
	switch runtime.GOARCH {
	case "arm64":
		if runtime.GOOS == "linux" {
			arch = "aarch64"
		} else {
			arch = "arm64"
		}
	case "amd64":
		if runtime.GOOS == "darwin" {
			arch = "x86_64"
		} else {
			arch = "x64"
		}
	default:
		return "", ""
	}
	return osName, arch
}

func resolveAssetURLs(o AssetOpts) (modelURL, runtimeURL string) {
	modelURL = strings.TrimSpace(o.ModelURL)
	if modelURL == "" {
		modelURL = DefaultModelURL
	}
	runtimeURL = strings.TrimSpace(o.RuntimeURL)
	if runtimeURL == "" {
		runtimeURL = DefaultRuntimeURL()
	}
	return modelURL, runtimeURL
}

func assetsMetaPath(dataDir string) string {
	return filepath.Join(dataDir, "assets-status.json")
}

func readAssetsMeta(dataDir string) AssetStatus {
	b, err := os.ReadFile(assetsMetaPath(dataDir))
	if err != nil {
		return AssetStatus{}
	}
	var s AssetStatus
	_ = json.Unmarshal(b, &s)
	return s
}

func writeAssetsMeta(dataDir string, s AssetStatus) {
	if dataDir == "" {
		return
	}
	_ = os.MkdirAll(dataDir, 0o700)
	b, err := json.Marshal(s)
	if err != nil {
		return
	}
	_ = os.WriteFile(assetsMetaPath(dataDir), b, 0o600)
}

func InspectAssets(dataDir string) AssetStatus {
	s := readAssetsMeta(dataDir)
	model := DefaultModelPath(dataDir)
	if st, err := os.Stat(model); err == nil && !st.IsDir() {
		s.ModelOK = true
		s.ModelPath = model
	}
	if lib, err := findRuntimeLib(dataDir); err == nil {
		s.RuntimeOK = true
		s.RuntimePath = lib
	}
	return s
}

// EnsureAssets downloads MiniLM ONNX and onnxruntime into the plugin data dir.
// It never panics; failures are returned and stored on AssetStatus.Error.
func EnsureAssets(ctx context.Context, o AssetOpts) (st AssetStatus, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("assets: panic: %v", rec)
			st.Error = err.Error()
		}
		if o.DataDir != "" {
			writeAssetsMeta(o.DataDir, st)
		}
	}()
	assetMu.Lock()
	defer assetMu.Unlock()

	modelURL, runtimeURL := resolveAssetURLs(o)
	st.ModelURL = modelURL
	st.RuntimeURL = runtimeURL

	if skipAssetDownload() {
		st = InspectAssets(o.DataDir)
		st.Skipped = true
		st.ModelURL = modelURL
		st.RuntimeURL = runtimeURL
		return st, nil
	}

	prev := readAssetsMeta(o.DataDir)
	modelPath := filepath.Join(o.DataDir, "models", packagedModelName)
	needModel := o.Force || prev.ModelURL != modelURL
	if _, err := os.Stat(modelPath); err != nil {
		needModel = true
	}
	if needModel {
		if err := downloadFile(ctx, modelURL, modelPath); err != nil {
			st.Error = err.Error()
			st.ModelOK = false
			cur := InspectAssets(o.DataDir)
			st.RuntimeOK = cur.RuntimeOK
			st.RuntimePath = cur.RuntimePath
			return st, err
		}
	}
	st.ModelOK = true
	st.ModelPath = modelPath

	libDir := filepath.Join(o.DataDir, "lib", runtime.GOOS+"_"+runtime.GOARCH)
	needRT := o.Force || prev.RuntimeURL != runtimeURL
	if _, err := findRuntimeLib(o.DataDir); err != nil {
		needRT = true
	}
	if needRT && runtimeURL != "" {
		if err := downloadRuntime(ctx, runtimeURL, libDir); err != nil {
			st.Error = err.Error()
			cur := InspectAssets(o.DataDir)
			st.RuntimeOK = cur.RuntimeOK
			st.RuntimePath = cur.RuntimePath
			return st, err
		}
	}
	if lib, err := findRuntimeLib(o.DataDir); err == nil {
		st.RuntimeOK = true
		st.RuntimePath = lib
	} else if runtimeURL == "" {
		st.Error = "no onnxruntime archive for this platform"
		return st, fmt.Errorf("%s", st.Error)
	}
	return st, nil
}

func downloadFile(ctx context.Context, url, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	tmp := dest + ".tmp"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	client := &http.Client{Timeout: 10 * time.Minute}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: status %d", url, res.StatusCode)
	}
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(f, res.Body)
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("download %s: %w", url, copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return closeErr
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func downloadRuntime(ctx context.Context, url, libDir string) error {
	tmp, err := os.MkdirTemp("", "conduit-ort-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	archive := filepath.Join(tmp, "ort.tgz")
	if err := downloadFile(ctx, url, archive); err != nil {
		return err
	}
	if err := extractORTLibs(archive, tmp); err != nil {
		return err
	}
	if err := os.MkdirAll(libDir, 0o755); err != nil {
		return err
	}
	return copyRuntimeLibs(tmp, libDir)
}

func extractORTLibs(archive, dest string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("onnxruntime archive: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		name := hdr.Name
		base := filepath.Base(name)
		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeGNUSparse {
			continue
		}
		if !isRuntimeLibName(base) {
			continue
		}
		out := filepath.Join(dest, base)
		w, err := os.Create(out)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(w, tr)
		closeErr := w.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		_ = os.Chmod(out, 0o755)
	}
}

func isRuntimeLibName(base string) bool {
	n := strings.ToLower(base)
	return strings.Contains(n, "onnxruntime") &&
		(strings.HasSuffix(n, ".dylib") || strings.HasSuffix(n, ".so") || strings.HasSuffix(n, ".dll") ||
			strings.Contains(n, ".so."))
}

func copyRuntimeLibs(srcDir, libDir string) error {
	copied := 0
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if !isRuntimeLibName(info.Name()) {
			return nil
		}
		dest := filepath.Join(libDir, info.Name())
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(dest)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(out, in)
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		_ = os.Chmod(dest, 0o755)
		copied++
		return nil
	})
	if err != nil {
		return err
	}
	if copied == 0 {
		return fmt.Errorf("onnxruntime archive had no shared library")
	}
	return nil
}
