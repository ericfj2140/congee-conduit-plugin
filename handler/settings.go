package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/michmich112/conduit-plugin/kinds"
	"github.com/michmich112/conduit-plugin/listing"
	sdk "github.com/michmich112/congee/sdk/plugin"
)

// Settings is Conduit plugin configuration (JSON in host item.settings).
type Settings struct {
	IndexBackend               string `json:"index_backend"`
	PostgresURL                string `json:"postgres_url"`
	PostgresUser               string `json:"postgres_user"`
	PostgresPassword           string `json:"postgres_password,omitempty"`
	ProductKinds               []int  `json:"product_kinds"`
	StallKinds                 []int  `json:"stall_kinds"`
	DraftKinds                 []int  `json:"draft_kinds"`
	DeletionKinds              []int  `json:"deletion_kinds"`
	IndexDrafts                bool   `json:"index_drafts"`
	GeoEnabled                 bool   `json:"geo_enabled"`
	VectorEnabled              bool   `json:"vector_enabled"`
	ActiveFilter               bool   `json:"active_filter"`
	RankAllProductReqs         bool   `json:"rank_all_product_reqs"`
	InjectProductKindsOnSearch bool   `json:"inject_product_kinds_on_search"`
	MaxResults                 int    `json:"max_results"`
	GeoMinPrefixLen            int    `json:"geo_min_prefix_len"`
	SearchCandidateCap         int    `json:"search_candidate_cap"`
}

func defaultSettings() Settings {
	return Settings{
		IndexBackend:               "turso",
		ProductKinds:               listing.DefaultProductKinds(),
		StallKinds:                 listing.DefaultStallKinds(),
		DraftKinds:                 listing.DefaultDraftKinds(),
		DeletionKinds:              listing.DefaultDeletionKinds(),
		GeoEnabled:                 true,
		VectorEnabled:              true,
		ActiveFilter:               true,
		MaxResults:                 0,
		GeoMinPrefixLen:            2,
		SearchCandidateCap:         2000,
		InjectProductKindsOnSearch: false,
	}
}

func parseSettings(raw json.RawMessage) (Settings, error) {
	s := defaultSettings()
	if len(raw) == 0 || string(raw) == "null" {
		return s, nil
	}
	if err := json.Unmarshal(raw, &s); err != nil {
		return s, fmt.Errorf("settings: %w", err)
	}
	s.IndexBackend = strings.ToLower(strings.TrimSpace(s.IndexBackend))
	if s.IndexBackend == "" {
		s.IndexBackend = "turso"
	}
	if s.IndexBackend != "turso" && s.IndexBackend != "postgres" {
		return s, fmt.Errorf("settings: index_backend must be turso or postgres")
	}
	if s.MaxResults < 0 {
		s.MaxResults = 0
	}
	if s.GeoMinPrefixLen <= 0 {
		s.GeoMinPrefixLen = 2
	}
	if s.SearchCandidateCap <= 0 {
		s.SearchCandidateCap = 2000
	}
	if len(s.ProductKinds) == 0 {
		s.ProductKinds = listing.DefaultProductKinds()
	} else {
		s.ProductKinds = keepKindsWithAnyRole(s.ProductKinds, kinds.RoleProduct, kinds.RoleListing)
		if len(s.ProductKinds) == 0 {
			s.ProductKinds = listing.DefaultProductKinds()
		}
	}
	if len(s.StallKinds) == 0 {
		s.StallKinds = listing.DefaultStallKinds()
	} else {
		s.StallKinds = keepKindsWithAnyRole(s.StallKinds, kinds.RoleStall)
		if len(s.StallKinds) == 0 {
			s.StallKinds = listing.DefaultStallKinds()
		}
	}
	if len(s.DraftKinds) == 0 {
		s.DraftKinds = listing.DefaultDraftKinds()
	} else {
		s.DraftKinds = keepKindsWithAnyRole(s.DraftKinds, kinds.RoleListingDraft)
		if len(s.DraftKinds) == 0 {
			s.DraftKinds = listing.DefaultDraftKinds()
		}
	}
	if len(s.DeletionKinds) == 0 {
		s.DeletionKinds = listing.DefaultDeletionKinds()
	} else {
		s.DeletionKinds = keepKindsWithAnyRole(s.DeletionKinds, kinds.RoleDeletion)
		if len(s.DeletionKinds) == 0 {
			s.DeletionKinds = listing.DefaultDeletionKinds()
		}
	}
	s.InjectProductKindsOnSearch = false
	return s, nil
}

func (s Settings) redacted() Settings {
	c := s
	c.PostgresPassword = ""
	return c
}

func (s Settings) allIndexKinds() []int {
	out := append([]int{}, s.ProductKinds...)
	out = append(out, s.StallKinds...)
	if s.IndexDrafts {
		out = append(out, s.DraftKinds...)
	}
	out = append(out, s.DeletionKinds...)
	return uniqueInts(out)
}

func (s Settings) interceptKinds() []int {
	out := append([]int{}, s.ProductKinds...)
	out = append(out, s.StallKinds...)
	if s.IndexDrafts {
		out = append(out, s.DraftKinds...)
	}
	return uniqueInts(out)
}

func uniqueInts(in []int) []int {
	seen := map[int]struct{}{}
	var out []int
	for _, n := range in {
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out
}

func keepKindsWithAnyRole(in []int, roles ...string) []int {
	var out []int
	seen := map[int]struct{}{}
	for _, n := range in {
		if _, ok := seen[n]; ok {
			continue
		}
		keep := false
		for _, role := range roles {
			if kinds.HasRole(n, role) {
				keep = true
				break
			}
		}
		if !keep {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out
}

func subscriptionsFor(s Settings) []sdk.TrafficSubscription {
	return []sdk.TrafficSubscription{
		{Kinds: s.allIndexKinds(), OnStoredEvent: true},
		{
			MessageTypes: []string{"REQ"},
			Kinds:        s.interceptKinds(),
			ReqHasSearch: true,
			ReqTagNames:  []string{"g"},
			InterceptREQ: true,
		},
	}
}

func loadSecrets(dataDir string) (password string, err error) {
	p := filepath.Join(dataDir, "secrets.json")
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return "", err
	}
	return m["postgres_password"], nil
}

func saveSecrets(dataDir, password string) error {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return err
	}
	p := filepath.Join(dataDir, "secrets.json")
	b, err := json.Marshal(map[string]string{"postgres_password": password})
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o600)
}

func postgresDSN(s Settings, password string) string {
	if strings.TrimSpace(s.PostgresURL) != "" {
		u := s.PostgresURL
		if password != "" && !strings.Contains(u, "password=") {
			if strings.Contains(u, "?") {
				u += "&password=" + password
			} else if strings.Contains(u, "@") {
				return u
			} else {
				u += "?password=" + password
			}
		}
		return u
	}
	user := s.PostgresUser
	if user == "" {
		user = "postgres"
	}
	return fmt.Sprintf("postgres://%s:%s@127.0.0.1:5432/conduit?sslmode=disable", user, password)
}
