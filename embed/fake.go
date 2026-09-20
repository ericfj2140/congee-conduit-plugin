package embed

import (
	"context"
	"hash/fnv"
	"strings"
)

const fakeDim = 384

// Fake is a deterministic bag-of-words embedder for tests and e2e (CONDUIT_EMBEDDER=fake).
type Fake struct{}

func (Fake) ModelID() string { return "fake-bow-384" }
func (Fake) Dim() int        { return fakeDim }

func (Fake) Embed(ctx context.Context, text string) ([]float32, error) {
	_ = ctx
	v := make([]float32, fakeDim)
	for _, tok := range strings.Fields(strings.ToLower(text)) {
		h := fnv.New32a()
		_, _ = h.Write([]byte(tok))
		idx := int(h.Sum32() % fakeDim)
		v[idx] += 1
		v[(idx+17)%fakeDim] += 0.5
	}
	L2Normalize(v)
	return v, nil
}
