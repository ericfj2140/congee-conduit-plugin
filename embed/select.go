package embed

import (
	"os"
	"strings"
)

const (
	SourceExplicitFake = "explicit_fake"
	SourceONNX         = "onnx"
	SourceUnavailable  = "unavailable"
)

// Selection is the embedder chosen at process start plus UI/status metadata.
type Selection struct {
	Embedder Embedder
	ModelID  string
	Source   string
	Error    string
	Warning  string
}

// VectorRanking reports whether this selection may rank with embeddings.
func (s Selection) VectorRanking() bool {
	return s.Embedder != nil
}

// Select chooses an embedder. Fake is used only when CONDUIT_EMBEDDER=fake.
// Otherwise ONNX is required; failure leaves Embedder nil and VectorRanking false.
func Select(modelPath string) Selection {
	env := strings.TrimSpace(os.Getenv("CONDUIT_EMBEDDER"))
	if env == "fake" {
		return ExplicitFake()
	}
	e, err := NewONNX(modelPath)
	if err != nil {
		return Selection{
			Source:  SourceUnavailable,
			Error:   err.Error(),
			Warning: "vector rank is disabled; onnx did not load. set CONDUIT_EMBEDDER=fake to opt in to the test bag-of-words embedder",
		}
	}
	return Selection{
		Embedder: e,
		ModelID:  e.ModelID(),
		Source:   SourceONNX,
	}
}

// ExplicitFake is the test embedder used when CONDUIT_EMBEDDER=fake.
func ExplicitFake() Selection {
	f := Fake{}
	return Selection{
		Embedder: f,
		ModelID:  f.ModelID(),
		Source:   SourceExplicitFake,
		Warning:  "vector rank uses the test bag-of-words embedder because CONDUIT_EMBEDDER=fake",
	}
}

// New returns an embedder for tests that already opted into Fake, or ONNX when it loads.
// Production startup should call Select instead so a missing ONNX is not silently faked.
func New(modelPath string) Embedder {
	if strings.TrimSpace(os.Getenv("CONDUIT_EMBEDDER")) == "fake" {
		return Fake{}
	}
	if e, err := NewONNX(modelPath); err == nil {
		return e
	}
	return nil
}
