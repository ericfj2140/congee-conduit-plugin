package embed

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
)

const (
	SourceExplicitFake = "explicit_fake"
	SourceONNX         = "onnx"
	SourceHTTP         = "http"
	SourceUnavailable  = "unavailable"
)

// Selection is the embedder chosen at process start plus UI/status metadata.
type Selection struct {
	Embedder Embedder
	ModelID  string
	Source   string
	Error    string
	Warning  string
	Dim      int
}

// VectorRanking reports whether this selection may rank with embeddings.
func (s Selection) VectorRanking() bool {
	return s.Embedder != nil
}

// SelectOpts chooses among fake (env), a verified HTTP provider, and on-device ONNX.
type SelectOpts struct {
	ModelPath        string
	Provider         string
	HTTPURL          string
	HTTPModel        string
	HTTPKey          string
	SavedFingerprint string
	Dim              int
}

// Select chooses an embedder. Fake is used only when CONDUIT_EMBEDDER=fake.
// Otherwise ONNX is required; failure leaves Embedder nil and VectorRanking false.
func Select(modelPath string) Selection {
	return SelectWith(SelectOpts{ModelPath: modelPath, Provider: "on_device"})
}

// SelectWith prefers a verified HTTP provider and otherwise loads ONNX.
// When HTTP is active, ONNX is not loaded.
func SelectWith(o SelectOpts) Selection {
	dim := normalizeDim(o.Dim)
	env := strings.TrimSpace(os.Getenv("CONDUIT_EMBEDDER"))
	if env == "fake" {
		return ExplicitFakeDim(dim)
	}
	wantHTTP := strings.EqualFold(strings.TrimSpace(o.Provider), "http")
	if wantHTTP {
		url := strings.TrimSpace(o.HTTPURL)
		if url != "" && o.SavedFingerprint != "" && o.SavedFingerprint == HTTPFingerprint(url, o.HTTPModel, o.HTTPKey, dim) {
			h := NewHTTP(url, o.HTTPModel, o.HTTPKey, dim)
			return Selection{
				Embedder: h,
				ModelID:  h.ModelID(),
				Source:   SourceHTTP,
				Dim:      h.Dim(),
				Warning:  "vector rank uses the external embedding provider; on-device MiniLM is not loaded",
			}
		}
	}
	if !wantHTTP && dim != DefaultDim {
		return Selection{
			Source:  SourceUnavailable,
			Warning: "on-device MiniLM is 384-d; set embed_dim to 384 or use an external provider whose vectors match embed_dim",
			Dim:     dim,
		}
	}
	e, err := NewONNX(o.ModelPath)
	if err != nil {
		warn := "vector rank is disabled; onnx did not load. set CONDUIT_EMBEDDER=fake to opt in to the test bag-of-words embedder"
		if wantHTTP {
			warn = "external provider is not verified (run Test); on-device MiniLM also did not load"
		}
		return Selection{
			Source:  SourceUnavailable,
			Error:   err.Error(),
			Warning: warn,
		}
	}
	s := Selection{
		Embedder: e,
		ModelID:  e.ModelID(),
		Source:   SourceONNX,
		Dim:      e.Dim(),
	}
	if wantHTTP {
		s.Warning = "external provider is not verified (run Test); using on-device MiniLM"
	}
	return s
}

// ExplicitFake is the test embedder used when CONDUIT_EMBEDDER=fake.
func ExplicitFake() Selection {
	return ExplicitFakeDim(DefaultDim)
}

// ExplicitFakeDim is CONDUIT_EMBEDDER=fake at a chosen width.
func ExplicitFakeDim(dim int) Selection {
	f := Fake{Width: normalizeDim(dim)}
	return Selection{
		Embedder: f,
		ModelID:  f.ModelID(),
		Source:   SourceExplicitFake,
		Dim:      f.Dim(),
		Warning:  "vector rank uses the test bag-of-words embedder because CONDUIT_EMBEDDER=fake",
	}
}

// HTTPFingerprint binds a successful probe to URL, model name, API key, and dim.
func HTTPFingerprint(url, model, key string, dim int) string {
	sum := sha256.Sum256([]byte(normalizeEmbeddingsURL(url) + "\n" + strings.TrimSpace(model) + "\n" + key + "\n" + strconv.Itoa(normalizeDim(dim))))
	return hex.EncodeToString(sum[:])
}

// New returns an embedder for tests that already opted into Fake, or ONNX when it loads.
// Production startup should call Select instead so a missing ONNX is not silently faked.
func New(modelPath string) Embedder {
	return Select(modelPath).Embedder
}
