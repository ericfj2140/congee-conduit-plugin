package embed

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const packagedModelName = "minilm.onnx"
const packagedTokenizerName = "tokenizer.json"

func runtimeLibNames() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{"libonnxruntime.dylib"}
	case "windows":
		return []string{"onnxruntime.dll"}
	default:
		return []string{"libonnxruntime.so"}
	}
}

func packageRootFromExe() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	dir := filepath.Dir(exe)
	base := filepath.Base(dir)
	if base == "bin" {
		return filepath.Clean(filepath.Join(dir, ".."))
	}
	return dir
}

func existingFile(paths ...string) string {
	for _, p := range paths {
		if p == "" {
			continue
		}
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}

// DefaultModelPath prefers models/minilm.onnx in the plugin package (beside bin/), then the data dir.
func DefaultModelPath(dataDir string) string {
	root := packageRootFromExe()
	candidates := []string{}
	if root != "" {
		candidates = append(candidates, filepath.Join(root, "models", packagedModelName))
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(dir, "models", packagedModelName),
			filepath.Join(dir, "..", "models", packagedModelName),
		)
	}
	if dataDir != "" {
		candidates = append(candidates, filepath.Join(dataDir, "models", packagedModelName))
	}
	if p := existingFile(candidates...); p != "" {
		return p
	}
	if dataDir != "" {
		return filepath.Join(dataDir, "models", packagedModelName)
	}
	return filepath.Join("models", packagedModelName)
}

// DefaultTokenizerPath is tokenizer.json beside the MiniLM weights.
func DefaultTokenizerPath(dataDir string) string {
	model := DefaultModelPath(dataDir)
	return filepath.Join(filepath.Dir(model), packagedTokenizerName)
}

// DefaultRuntimeLibPath is the packaged onnxruntime shared library, if present.
func DefaultRuntimeLibPath() string {
	p, _ := findRuntimeLib("")
	return p
}

func findRuntimeLib(dataDir string) (string, error) {
	root := packageRootFromExe()
	plat := runtime.GOOS + "_" + runtime.GOARCH
	var dirs []string
	if root != "" {
		dirs = append(dirs,
			filepath.Join(root, "lib", plat),
			filepath.Join(root, "lib"),
		)
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		dirs = append(dirs,
			filepath.Join(dir, "lib", plat),
			filepath.Join(dir, "lib"),
			filepath.Join(dir, "..", "lib", plat),
			filepath.Join(dir, "..", "lib"),
		)
	}
	if dataDir != "" {
		dirs = append(dirs,
			filepath.Join(dataDir, "lib", plat),
			filepath.Join(dataDir, "lib"),
		)
	}
	names := runtimeLibNames()
	for _, dir := range dirs {
		for _, name := range names {
			p := filepath.Join(dir, name)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return p, nil
			}
		}
	}
	return "", fmt.Errorf("not found (looked in package lib/%s)", plat)
}

// PrepareRuntimeLibrary prepends the packaged onnxruntime directory to the
// dynamic linker path so a linked build can dlopen the library next to the plugin.
func PrepareRuntimeLibrary(dataDir string) {
	lib, err := findRuntimeLib(dataDir)
	if err != nil {
		return
	}
	dir := filepath.Dir(lib)
	key := "LD_LIBRARY_PATH"
	if runtime.GOOS == "darwin" {
		key = "DYLD_LIBRARY_PATH"
	}
	if runtime.GOOS == "windows" {
		key = "PATH"
	}
	cur := os.Getenv(key)
	if cur == "" {
		_ = os.Setenv(key, dir)
		return
	}
	_ = os.Setenv(key, dir+string(os.PathListSeparator)+cur)
}
