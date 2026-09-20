package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const httpMaxBody = 16 << 20

// HTTP is an OpenAI-compatible embeddings client. The response vector must match Dim().
type HTTP struct {
	url    string
	model  string
	key    string
	dim    int
	client *http.Client
}

// NewHTTP builds a client. url may be a full embeddings path or an API root
// (…/v1 or a host); /v1/embeddings is appended when missing. dim<=0 uses DefaultDim.
func NewHTTP(url, model, key string, dim int) *HTTP {
	return &HTTP{
		url:   normalizeEmbeddingsURL(url),
		model: strings.TrimSpace(model),
		key:   key,
		dim:   normalizeDim(dim),
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (h *HTTP) ModelID() string {
	m := h.model
	if m == "" {
		m = "default"
	}
	return "http:" + m
}

func (h *HTTP) Dim() int { return h.dim }

func (h *HTTP) Embed(ctx context.Context, text string) ([]float32, error) {
	if h.url == "" {
		return nil, fmt.Errorf("http embed: empty url")
	}
	want := h.Dim()
	reqBody := map[string]any{
		"input":      text,
		"dimensions": want,
	}
	if h.model != "" {
		reqBody["model"] = h.model
	}
	raw, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.url, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(h.key) != "" {
		req.Header.Set("Authorization", "Bearer "+h.key)
	}
	res, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http embed: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, httpMaxBody))
	if err != nil {
		return nil, fmt.Errorf("http embed: read: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("http embed: status %d: %s", res.StatusCode, truncateErrBody(body))
	}
	vec, err := parseEmbeddingResponse(body)
	if err != nil {
		return nil, err
	}
	if len(vec) != want {
		return nil, fmt.Errorf("http embed: dim %d, want %d", len(vec), want)
	}
	L2Normalize(vec)
	return vec, nil
}

func normalizeEmbeddingsURL(raw string) string {
	u := strings.TrimRight(strings.TrimSpace(raw), "/")
	if u == "" {
		return ""
	}
	lower := strings.ToLower(u)
	if strings.HasSuffix(lower, "/embeddings") {
		return u
	}
	if strings.HasSuffix(lower, "/v1") {
		return u + "/embeddings"
	}
	return u + "/v1/embeddings"
}

func parseEmbeddingResponse(body []byte) ([]float32, error) {
	var env struct {
		Error json.RawMessage `json:"error"`
		Data  []struct {
			Embedding json.RawMessage `json:"embedding"`
		} `json:"data"`
		Embedding json.RawMessage `json:"embedding"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("http embed: json: %w", err)
	}
	if len(env.Error) > 0 && string(env.Error) != "null" {
		return nil, fmt.Errorf("http embed: %s", truncateErrBody(env.Error))
	}
	raw := env.Embedding
	if len(env.Data) > 0 && len(env.Data[0].Embedding) > 0 {
		raw = env.Data[0].Embedding
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("http embed: missing embedding")
	}
	var floats []float64
	if err := json.Unmarshal(raw, &floats); err != nil {
		return nil, fmt.Errorf("http embed: embedding must be a float array (dim %d): %w", DefaultDim, err)
	}
	out := make([]float32, len(floats))
	for i, x := range floats {
		out[i] = float32(x)
	}
	return out, nil
}

func truncateErrBody(b []byte) string {
	s := strings.TrimSpace(string(b))
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > 300 {
		return s[:300] + "…"
	}
	return s
}

// Probe embeds a short string and requires e.Dim().
func Probe(ctx context.Context, e Embedder) error {
	if e == nil {
		return fmt.Errorf("nil embedder")
	}
	v, err := e.Embed(ctx, "conduit embed probe")
	if err != nil {
		return err
	}
	if len(v) != e.Dim() {
		return fmt.Errorf("embedding dim %d, want %d", len(v), e.Dim())
	}
	return nil
}
