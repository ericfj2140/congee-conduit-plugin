package embed

import "testing"

func TestSelectExplicitFake(t *testing.T) {
	t.Setenv("CONDUIT_EMBEDDER", "fake")
	s := Select("missing.onnx")
	if s.Source != SourceExplicitFake || s.Embedder == nil || s.ModelID != (Fake{}).ModelID() {
		t.Fatalf("%+v", s)
	}
	if !s.VectorRanking() {
		t.Fatal("explicit fake should allow vector ranking")
	}
}

func TestSelectUnavailableDoesNotFake(t *testing.T) {
	t.Setenv("CONDUIT_EMBEDDER", "")
	s := Select("definitely-missing.onnx")
	if s.Embedder != nil || s.VectorRanking() {
		t.Fatalf("missing onnx must not fall back to Fake: %+v", s)
	}
	if s.Source != SourceUnavailable || s.Error == "" || s.Warning == "" {
		t.Fatalf("%+v", s)
	}
}
