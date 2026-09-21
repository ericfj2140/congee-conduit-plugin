package handler

import (
	"encoding/json"
	"testing"
)

func TestSubscriptionsForInterceptsProductKindsOnly(t *testing.T) {
	s := defaultSettings()
	subs := subscriptionsFor(s)
	if len(subs) != 2 {
		t.Fatalf("subs %d", len(subs))
	}
	req := subs[1]
	if !req.InterceptREQ || len(req.MessageTypes) != 1 || req.MessageTypes[0] != "REQ" {
		t.Fatalf("%+v", req)
	}
	if req.ReqHasSearch || len(req.ReqTagNames) != 0 {
		t.Fatal("search and #g must not subscribe without product kinds")
	}
	if len(req.Kinds) == 0 {
		t.Fatal("expected product kinds")
	}
	want := map[int]bool{}
	for _, k := range s.ProductKinds {
		want[k] = true
	}
	for _, k := range req.Kinds {
		if !want[k] {
			t.Fatalf("intercept kind %d is not a product kind", k)
		}
	}
	if !subs[0].OnStoredEvent {
		t.Fatal("expected OnStoredEvent subscription")
	}
	for _, k := range s.StallKinds {
		found := false
		for _, got := range subs[0].Kinds {
			if got == k {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("store still indexes stall kind %d", k)
		}
	}
	s.IndexDrafts = true
	req = subscriptionsFor(s)[1]
	for _, k := range s.DraftKinds {
		for _, got := range req.Kinds {
			if got == k {
				t.Fatalf("draft kind %d must not intercept REQ", k)
			}
		}
	}
}

func TestParseSettingsDropsCommunityFromStallKinds(t *testing.T) {
	raw, _ := json.Marshal(map[string]any{
		"stall_kinds":   []int{30017, 34550},
		"product_kinds": []int{30018, 34560, 30402},
	})
	s, err := parseSettings(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range s.StallKinds {
		if k == 34550 {
			t.Fatal("34550 must not remain a stall kind")
		}
	}
	for _, k := range s.ProductKinds {
		if k == 34560 || k == 34550 {
			t.Fatalf("unofficial/community kind %d in product_kinds", k)
		}
	}
	if len(s.StallKinds) != 1 || s.StallKinds[0] != 30017 {
		t.Fatalf("stall kinds %v", s.StallKinds)
	}
}

func TestParseSettingsEmbedProvider(t *testing.T) {
	raw, _ := json.Marshal(map[string]any{
		"embed_provider":   "http",
		"embed_http_url":   " https://embed.example/v1 ",
		"embed_http_model": "all-minilm",
	})
	s, err := parseSettings(raw)
	if err != nil {
		t.Fatal(err)
	}
	if s.EmbedProvider != "http" || s.EmbedHTTPURL != "https://embed.example/v1" || s.EmbedHTTPModel != "all-minilm" {
		t.Fatalf("%+v", s)
	}
	if s.redacted().EmbedHTTPAPIKey != "" {
		t.Fatal("api key must redact")
	}
	_, err = parseSettings([]byte(`{"embed_provider":"other"}`))
	if err == nil {
		t.Fatal("expected invalid provider")
	}
	_, err = parseSettings([]byte(`{"embed_dim": 3}`))
	if err == nil {
		t.Fatal("expected invalid dim")
	}
}

func TestSaveSecretsPreservesEmbedKey(t *testing.T) {
	dir := t.TempDir()
	if err := saveSecretsFile(dir, secretsFile{EmbedHTTPAPIKey: "k", EmbedHTTPFingerprint: "fp"}); err != nil {
		t.Fatal(err)
	}
	if err := saveSecrets(dir, "pw"); err != nil {
		t.Fatal(err)
	}
	s, err := loadSecretsFile(dir)
	if err != nil {
		t.Fatal(err)
	}
	if s.PostgresPassword != "pw" || s.EmbedHTTPAPIKey != "k" || s.EmbedHTTPFingerprint != "fp" {
		t.Fatalf("%+v", s)
	}
}
