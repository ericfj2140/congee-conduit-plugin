package embed

import (
	"context"
	"hash/fnv"
	"strconv"
	"strings"
)

// Fake is a deterministic bag-of-words embedder for tests and e2e (CONDUIT_EMBEDDER=fake).
type Fake struct {
	Width int
}

func (f Fake) dim() int {
	if f.Width > 0 {
		return f.Width
	}
	return DefaultDim
}

func (f Fake) ModelID() string { return "fake-bow-" + strconv.Itoa(f.dim()) }
func (f Fake) Dim() int        { return f.dim() }

func (f Fake) Embed(ctx context.Context, text string) ([]float32, error) {
	_ = ctx
	n := f.dim()
	v := make([]float32, n)
	for _, tok := range strings.Fields(strings.ToLower(text)) {
		h := fnv.New32a()
		_, _ = h.Write([]byte(tok))
		idx := int(h.Sum32() % uint32(n))
		v[idx] += 1
		v[(idx+17)%n] += 0.5
	}
	L2Normalize(v)
	return v, nil
}
