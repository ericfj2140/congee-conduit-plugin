package embed

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	ort "github.com/yalue/onnxruntime_go"
)

// ONNX runs all-MiniLM-L6-v2 via onnxruntime (384-d mean-pooled, L2-normalized).
type ONNX struct {
	mu       sync.Mutex
	session  *ort.DynamicAdvancedSession
	tok      *Tokenizer
	dim      int
	inNames  []string
	outNames []string
}

func (o *ONNX) ModelID() string { return "all-MiniLM-L6-v2" }
func (o *ONNX) Dim() int        { return o.dim }

// NewONNX loads MiniLM ONNX from path and tokenizer.json beside it. The onnxruntime
// shared library must already be findable (PrepareRuntimeLibrary / data/lib).
func NewONNX(path string) (Embedder, error) {
	if path == "" {
		return nil, fmt.Errorf("onnx: empty path")
	}
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("onnx: model not found at %s (%w); ship models/minilm.onnx in the plugin data dir", path, err)
	}
	dataDir := inferDataDir(path)
	lib, err := findRuntimeLib(dataDir)
	if err != nil {
		return nil, fmt.Errorf("onnx: model present at %s but onnxruntime library missing (%v); expected data/lib/%s_%s/libonnxruntime.*", path, err, runtime.GOOS, runtime.GOARCH)
	}
	tokPath := filepath.Join(filepath.Dir(path), "tokenizer.json")
	tok, err := LoadTokenizer(tokPath)
	if err != nil {
		return nil, fmt.Errorf("onnx: tokenizer: %w", err)
	}
	if err := initORT(lib); err != nil {
		return nil, err
	}
	inputs, outputs, err := ort.GetInputOutputInfo(path)
	if err != nil {
		return nil, fmt.Errorf("onnx: inspect model: %w", err)
	}
	inNames := pickInputNames(inputs)
	outNames := pickOutputNames(outputs)
	if len(inNames) == 0 || len(outNames) == 0 {
		return nil, fmt.Errorf("onnx: unexpected io names inputs=%v outputs=%v", namesOf(inputs), namesOf(outputs))
	}
	sess, err := ort.NewDynamicAdvancedSession(path, inNames, outNames, nil)
	if err != nil {
		return nil, fmt.Errorf("onnx: session: %w", err)
	}
	return &ONNX{session: sess, tok: tok, dim: DefaultDim, inNames: inNames, outNames: outNames}, nil
}

func (o *ONNX) Embed(ctx context.Context, text string) ([]float32, error) {
	_ = ctx
	o.mu.Lock()
	defer o.mu.Unlock()
	ids, mask := o.tok.Encode(text)
	seq := int64(len(ids))
	idT, err := ort.NewTensor(ort.NewShape(1, seq), ids)
	if err != nil {
		return nil, err
	}
	defer idT.Destroy()
	maskT, err := ort.NewTensor(ort.NewShape(1, seq), mask)
	if err != nil {
		return nil, err
	}
	defer maskT.Destroy()
	inputs := make([]ort.Value, len(o.inNames))
	for i, n := range o.inNames {
		switch {
		case strings.Contains(strings.ToLower(n), "mask"):
			inputs[i] = maskT
		case strings.Contains(strings.ToLower(n), "type"):
			zeros := make([]int64, len(ids))
			tt, err := ort.NewTensor(ort.NewShape(1, seq), zeros)
			if err != nil {
				return nil, err
			}
			defer tt.Destroy()
			inputs[i] = tt
		default:
			inputs[i] = idT
		}
	}
	outputs := make([]ort.Value, len(o.outNames))
	if err := o.session.Run(inputs, outputs); err != nil {
		return nil, fmt.Errorf("onnx: run: %w", err)
	}
	for i := range outputs {
		if outputs[i] != nil {
			defer outputs[i].Destroy()
		}
	}
	hidden, ok := outputs[0].(*ort.Tensor[float32])
	if !ok {
		return nil, fmt.Errorf("onnx: output is not float32 tensor")
	}
	shape := hidden.GetShape()
	if len(shape) == 2 {
		v := append([]float32(nil), hidden.GetData()...)
		if int(shape[1]) > 0 {
			L2Normalize(v)
		}
		return v, nil
	}
	return meanPool(hidden.GetData(), shape, mask)
}

func meanPool(data []float32, shape ort.Shape, mask []int64) ([]float32, error) {
	if len(shape) < 2 {
		return nil, fmt.Errorf("onnx: unexpected output shape %v", shape)
	}
	seq := int(shape[len(shape)-2])
	dim := int(shape[len(shape)-1])
	if dim <= 0 || seq <= 0 || len(data) < seq*dim {
		return nil, fmt.Errorf("onnx: bad output shape %v len=%d", shape, len(data))
	}
	out := make([]float32, dim)
	var count float32
	for t := 0; t < seq && t < len(mask); t++ {
		if mask[t] == 0 {
			continue
		}
		off := t * dim
		for d := 0; d < dim; d++ {
			out[d] += data[off+d]
		}
		count++
	}
	if count == 0 {
		return out, nil
	}
	inv := 1 / count
	for i := range out {
		out[i] *= inv
	}
	L2Normalize(out)
	return out, nil
}

func initORT(lib string) error {
	if !ort.IsInitialized() {
		ort.SetSharedLibraryPath(lib)
		if err := ort.InitializeEnvironment(); err != nil {
			return fmt.Errorf("onnx: init runtime %s: %w", lib, err)
		}
	}
	return nil
}

func inferDataDir(modelPath string) string {
	dir := filepath.Dir(modelPath)
	if filepath.Base(dir) == "models" {
		return filepath.Dir(dir)
	}
	return dir
}

func pickInputNames(in []ort.InputOutputInfo) []string {
	var ids, mask, types string
	for _, n := range in {
		l := strings.ToLower(n.Name)
		switch {
		case strings.Contains(l, "mask"):
			mask = n.Name
		case strings.Contains(l, "type"):
			types = n.Name
		case strings.Contains(l, "input_ids") || l == "input_ids" || strings.Contains(l, "ids"):
			if ids == "" {
				ids = n.Name
			}
		}
	}
	if ids == "" && len(in) > 0 {
		ids = in[0].Name
	}
	out := []string{}
	if ids != "" {
		out = append(out, ids)
	}
	if mask != "" {
		out = append(out, mask)
	}
	if types != "" {
		out = append(out, types)
	}
	return out
}

func pickOutputNames(out []ort.InputOutputInfo) []string {
	for _, n := range out {
		l := strings.ToLower(n.Name)
		if strings.Contains(l, "sentence") || strings.Contains(l, "pool") {
			return []string{n.Name}
		}
	}
	for _, n := range out {
		l := strings.ToLower(n.Name)
		if strings.Contains(l, "last_hidden") || strings.Contains(l, "token_embed") {
			return []string{n.Name}
		}
	}
	if len(out) > 0 {
		return []string{out[0].Name}
	}
	return nil
}

func namesOf(in []ort.InputOutputInfo) []string {
	s := make([]string, len(in))
	for i, n := range in {
		s[i] = n.Name
	}
	return s
}
