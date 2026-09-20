package handler

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/michmich112/conduit-plugin/embed"
	"github.com/michmich112/conduit-plugin/index"
	"github.com/michmich112/conduit-plugin/listing"
	sdk "github.com/michmich112/congee/sdk/plugin"
)

// Handler implements sdk.Handler for Conduit.
type Handler struct {
	dataDir  string
	embedder embed.Embedder

	mu              sync.RWMutex
	settings        Settings
	store           index.Store
	ready           bool
	storeOpen       bool
	embedWarm       bool
	backfillStarted bool
	backfillState   string
	lastErr         string
	host            sdk.Host

	interceptN   atomic.Int64
	passthroughN atomic.Int64
	respondN     atomic.Int64
	reshapeN     atomic.Int64
}

func New(dataDir string, e embed.Embedder) *Handler {
	return &Handler{
		dataDir:       dataDir,
		embedder:      e,
		settings:      defaultSettings(),
		backfillState: "idle",
	}
}

func (h *Handler) SetHost(host sdk.Host) { h.host = host }

func (h *Handler) Handshake(ctx context.Context, settings json.RawMessage) (*sdk.HandshakeResult, error) {
	if err := h.apply(ctx, settings, true); err != nil {
		return nil, err
	}
	h.startBackfill(context.WithoutCancel(ctx))
	h.mu.RLock()
	st := h.settings
	h.mu.RUnlock()
	return &sdk.HandshakeResult{
		PluginID:            "conduit",
		Name:                "Conduit",
		Version:             "0.1.0",
		Capabilities:        []string{sdk.CapIntercept, sdk.CapEventsRead, sdk.CapIndexOwn, sdk.CapAdminUI},
		Subscriptions:       subscriptionsFor(st),
		InterceptDeadlineMs: 200,
	}, nil
}

func (h *Handler) Health(ctx context.Context) (bool, string, error) {
	_ = ctx
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.ready {
		return true, "ready", nil
	}
	return false, "not ready: " + h.lastErr + " backfill=" + h.backfillState, nil
}

func (h *Handler) Observe(ctx context.Context, msg sdk.ObserveMessage) error {
	_ = ctx
	_ = msg
	return nil
}

func (h *Handler) OnStoredEvent(ctx context.Context, ev sdk.Event, stored bool) error {
	if !stored {
		return nil
	}
	h.mu.RLock()
	st := h.settings
	store := h.store
	h.mu.RUnlock()
	if store == nil {
		return nil
	}
	l, ok := listing.FromEvent(listing.Event{
		ID: ev.ID, PubKey: ev.PubKey, CreatedAt: ev.CreatedAt, Kind: ev.Kind, Tags: ev.Tags, Content: ev.Content,
	}, st.IndexDrafts)
	if !ok {
		return nil
	}
	return store.Upsert(ctx, l)
}

func (h *Handler) InterceptREQ(ctx context.Context, req sdk.Req) (*sdk.InterceptResult, error) {
	h.interceptN.Add(1)
	h.mu.RLock()
	st := h.settings
	ready := h.ready
	store := h.store
	h.mu.RUnlock()
	d := decide(req, st, ready)
	if d.kind == decPassthrough {
		h.passthroughN.Add(1)
		return &sdk.InterceptResult{Action: sdk.InterceptPassthrough}, nil
	}
	if d.kind == decReshape {
		h.reshapeN.Add(1)
		return &sdk.InterceptResult{Action: sdk.InterceptReshapeREQ, ReshapeFilters: d.filters}, nil
	}
	if store == nil {
		h.passthroughN.Add(1)
		return &sdk.InterceptResult{Action: sdk.InterceptPassthrough}, nil
	}
	f := sdk.Filter{}
	if len(req.Filters) > 0 {
		f = req.Filters[0]
	}
	q := index.Query{
		Search:             f.Search,
		Kinds:              kindsForSearch(f, st),
		Authors:            f.Authors,
		Since:              f.Since,
		Until:              f.Until,
		GeoPrefixes:        geoPrefixes(f),
		Limit:              mergeLimit(f, st.MaxResults),
		GeoMinPrefixLen:    st.GeoMinPrefixLen,
		SearchCandidateCap: st.SearchCandidateCap,
		ActiveOnly:         st.ActiveFilter,
		VectorEnabled:      st.VectorEnabled,
		GeoEnabled:         st.GeoEnabled,
	}
	if d.kind == decRespondGeo {
		q.Search = ""
		q.VectorEnabled = false
	}
	ids, err := store.Search(ctx, q)
	if err != nil {
		h.passthroughN.Add(1)
		h.log(ctx, "warn", "intercept search failed", map[string]string{"error": err.Error()})
		return &sdk.InterceptResult{Action: sdk.InterceptPassthrough}, nil
	}
	h.respondN.Add(1)
	return &sdk.InterceptResult{
		Action:              sdk.InterceptRespond,
		EventIDs:            ids,
		SubscriptionFilters: stripSearch(req.Filters),
	}, nil
}

func (h *Handler) ApplySettings(ctx context.Context, settings json.RawMessage) ([]sdk.TrafficSubscription, error) {
	if err := h.apply(ctx, settings, false); err != nil {
		return nil, err
	}
	h.mu.RLock()
	st := h.settings
	h.mu.RUnlock()
	return subscriptionsFor(st), nil
}

func (h *Handler) AdminAction(ctx context.Context, name string, payload json.RawMessage) (json.RawMessage, error) {
	switch name {
	case "rebuild":
		h.mu.Lock()
		h.backfillStarted = false
		h.mu.Unlock()
		h.startBackfill(context.WithoutCancel(ctx))
		return json.Marshal(map[string]any{"ok": true})
	case "test_store":
		h.mu.RLock()
		store := h.store
		h.mu.RUnlock()
		if store == nil {
			return json.Marshal(map[string]any{"ok": false, "error": "store not open"})
		}
		if err := store.Ping(ctx); err != nil {
			return json.Marshal(map[string]any{"ok": false, "error": err.Error()})
		}
		return json.Marshal(map[string]any{"ok": true})
	default:
		_ = payload
		return json.Marshal(map[string]any{"ok": false, "error": "unknown action"})
	}
}

func (h *Handler) Status(ctx context.Context) (*sdk.Status, error) {
	h.mu.RLock()
	st := h.settings
	ready := h.ready
	store := h.store
	bf := h.backfillState
	h.mu.RUnlock()
	stats := index.Stats{Backend: st.IndexBackend}
	if store != nil {
		if s, err := store.Stats(ctx); err == nil {
			stats = s
		}
	}
	body, _ := json.Marshal(map[string]any{
		"backend":            stats.Backend,
		"active":             stats.Active,
		"inactive":           stats.Inactive,
		"embeddings":         stats.Embeddings,
		"embedding_mismatch": stats.EmbeddingMismatch,
		"backfill":           bf,
		"intercept_n":        h.interceptN.Load(),
		"passthrough_n":      h.passthroughN.Load(),
		"respond_n":          h.respondN.Load(),
		"reshape_n":          h.reshapeN.Load(),
		"settings":           st.redacted(),
	})
	return &sdk.Status{Ready: ready, JSON: body}, nil
}

func (h *Handler) apply(ctx context.Context, raw json.RawMessage, opening bool) error {
	st, err := parseSettings(raw)
	if err != nil {
		return err
	}
	pw, _ := loadSecrets(h.dataDir)
	if st.PostgresPassword != "" {
		_ = saveSecrets(h.dataDir, st.PostgresPassword)
		pw = st.PostgresPassword
		st.PostgresPassword = ""
	}
	e := h.embedder
	if e == nil || os.Getenv("CONDUIT_EMBEDDER") == "fake" {
		e = embed.Fake{}
		h.embedder = e
	}
	if err := embed.Warm(ctx, e); err != nil {
		h.mu.Lock()
		h.lastErr = err.Error()
		h.embedWarm = false
		h.ready = false
		h.mu.Unlock()
		return err
	}
	store, err := h.openStore(ctx, st, pw, e)
	if err != nil {
		h.mu.Lock()
		h.lastErr = err.Error()
		h.storeOpen = false
		h.ready = false
		h.mu.Unlock()
		return err
	}
	h.mu.Lock()
	if h.store != nil {
		_ = h.store.Close()
	}
	h.store = store
	h.settings = st
	h.storeOpen = true
	h.embedWarm = true
	h.ready = true
	h.lastErr = ""
	h.mu.Unlock()
	_ = opening
	return nil
}

func (h *Handler) openStore(ctx context.Context, st Settings, password string, e embed.Embedder) (index.Store, error) {
	if st.IndexBackend == "postgres" {
		return index.OpenPostgres(ctx, postgresDSN(st, password), e)
	}
	path := filepath.Join(h.dataDir, "conduit-index.db")
	return index.OpenTurso(ctx, path, e)
}

func (h *Handler) log(ctx context.Context, level, msg string, fields map[string]string) {
	if h.host == nil {
		return
	}
	_ = h.host.Log(ctx, level, msg, fields)
}
