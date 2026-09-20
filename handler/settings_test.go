package handler

import (
	"encoding/json"
	"testing"
)

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
