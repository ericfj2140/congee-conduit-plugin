package index

import (
	"sort"
	"sync"

	"github.com/michmich112/conduit-plugin/embed"
)

type annItem struct {
	Coord     string
	EventID   string
	CreatedAt int64
	Vec       []float32
}

// ANN is an in-memory active-vector cache. Brute force ≤50k; same path above that in v1.
type ANN struct {
	mu    sync.RWMutex
	items map[string]annItem
}

func NewANN() *ANN { return &ANN{items: map[string]annItem{}} }

func (a *ANN) Upsert(coord, eventID string, created int64, vec []float32) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.items[coord] = annItem{Coord: coord, EventID: eventID, CreatedAt: created, Vec: vec}
}

func (a *ANN) Delete(coord string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.items, coord)
}

func (a *ANN) Replace(items map[string]annItem) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if items == nil {
		a.items = map[string]annItem{}
		return
	}
	a.items = items
}

func (a *ANN) Snapshot(coords map[string]annItem) []annItem {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if coords == nil {
		out := make([]annItem, 0, len(a.items))
		for _, it := range a.items {
			out = append(out, it)
		}
		return out
	}
	out := make([]annItem, 0, len(coords))
	for c := range coords {
		if it, ok := a.items[c]; ok {
			out = append(out, it)
		}
	}
	return out
}

func rankByCosine(query []float32, items []annItem) {
	sort.SliceStable(items, func(i, j int) bool {
		si := embed.Cosine(query, items[i].Vec)
		sj := embed.Cosine(query, items[j].Vec)
		if si == sj {
			if items[i].CreatedAt == items[j].CreatedAt {
				return items[i].EventID > items[j].EventID
			}
			return items[i].CreatedAt > items[j].CreatedAt
		}
		return si > sj
	})
}
