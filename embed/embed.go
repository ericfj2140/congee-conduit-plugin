package embed

import (
	"context"
	"math"
)

// DefaultDim is all-MiniLM-L6-v2's width. Operators may set embed_dim for other models.
const DefaultDim = 384

const (
	MinDim = 8
	MaxDim = 4096
)

func normalizeDim(d int) int {
	if d <= 0 {
		return DefaultDim
	}
	return d
}

func validDim(d int) bool {
	return d >= MinDim && d <= MaxDim
}

// Embedder produces L2-normalized vectors.
type Embedder interface {
	ModelID() string
	Dim() int
	Embed(ctx context.Context, text string) ([]float32, error)
}

// L2Normalize scales v to unit length in place.
func L2Normalize(v []float32) {
	var sum float64
	for _, x := range v {
		sum += float64(x) * float64(x)
	}
	if sum == 0 {
		return
	}
	inv := float32(1 / math.Sqrt(sum))
	for i := range v {
		v[i] *= inv
	}
}

// Cosine returns the dot product of two L2-normalized vectors.
func Cosine(a, b []float32) float32 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var sum float32
	for i := 0; i < n; i++ {
		sum += a[i] * b[i]
	}
	return sum
}
