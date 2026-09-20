package index

import (
	"context"
	"math"
	"time"

	"github.com/michmich112/conduit-plugin/embed"
	"github.com/michmich112/conduit-plugin/listing"
)

// Query is the search planner input.
type Query struct {
	Search             string
	Kinds              []int
	Authors            []string
	Since              *int64
	Until              *int64
	GeoPrefixes        []string
	Limit              int
	GeoMinPrefixLen    int
	SearchCandidateCap int
	ActiveOnly         bool
	VectorEnabled      bool
	GeoEnabled         bool
}

// Stats is Overview UI counts.
type Stats struct {
	Active            int64  `json:"active"`
	Inactive          int64  `json:"inactive"`
	Embeddings        int64  `json:"embeddings"`
	EmbeddingMismatch int64  `json:"embedding_mismatch"`
	Backend           string `json:"backend"`
}

// Store is the only persistence API.
type Store interface {
	Upsert(ctx context.Context, l listing.Listing) error
	MarkInactive(ctx context.Context, pubkey string, eventIDs, coords []string) error
	Get(ctx context.Context, coord string) (listing.Listing, bool, error)
	Search(ctx context.Context, q Query) ([]string, error)
	Meta(ctx context.Context, key string) (string, error)
	SetMeta(ctx context.Context, key, value string) error
	Stats(ctx context.Context) (Stats, error)
	Ping(ctx context.Context) error
	Close() error
}

func clampLimit(n, max int) int {
	if n <= 0 {
		return 20
	}
	if max > 0 && n > max {
		return max
	}
	return n
}

func nowUnix() int64 { return time.Now().Unix() }

func floatsToBytes(v []float32) []byte {
	b := make([]byte, len(v)*4)
	for i, x := range v {
		u := math.Float32bits(x)
		b[i*4] = byte(u)
		b[i*4+1] = byte(u >> 8)
		b[i*4+2] = byte(u >> 16)
		b[i*4+3] = byte(u >> 24)
	}
	return b
}

func bytesToFloats(b []byte) []float32 {
	n := len(b) / 4
	v := make([]float32, n)
	for i := 0; i < n; i++ {
		u := uint32(b[i*4]) | uint32(b[i*4+1])<<8 | uint32(b[i*4+2])<<16 | uint32(b[i*4+3])<<24
		v[i] = math.Float32frombits(u)
	}
	return v
}

func ensureEmbedder(e embed.Embedder) embed.Embedder {
	if e == nil {
		return embed.Fake{}
	}
	return e
}
