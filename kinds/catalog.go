package kinds

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
)

// Role names used by Conduit defaults and the plugin UI.
const (
	RoleStall        = "stall"
	RoleProduct      = "product"
	RoleListing      = "listing"
	RoleListingDraft = "listing_draft"
	RoleDeletion     = "deletion"
	RoleCommunity    = "community"
)

// Entry is one well-known Nostr kind from kinds.json.
type Entry struct {
	Kind        int      `json:"kind"`
	NIP         string   `json:"nip"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Roles       []string `json:"roles"`
}

type file struct {
	Kinds []Entry `json:"kinds"`
}

var (
	loadOnce sync.Once
	loadErr  error
	byKind   map[int]Entry
	byRole   map[string][]int
)

func load() {
	loadOnce.Do(func() {
		var f file
		if err := json.Unmarshal(embeddedJSON, &f); err != nil {
			loadErr = fmt.Errorf("kinds: parse catalog: %w", err)
			byKind = map[int]Entry{}
			byRole = map[string][]int{}
			return
		}
		byKind = make(map[int]Entry, len(f.Kinds))
		byRole = map[string][]int{}
		for _, e := range f.Kinds {
			byKind[e.Kind] = e
			for _, r := range e.Roles {
				if r == "" {
					continue
				}
				byRole[r] = append(byRole[r], e.Kind)
			}
		}
		for r := range byRole {
			sort.Ints(byRole[r])
		}
	})
}

// LoadError is non-nil when kinds.json could not be parsed.
func LoadError() error {
	load()
	return loadErr
}

// Lookup returns the catalog entry for kind.
func Lookup(kind int) (Entry, bool) {
	load()
	e, ok := byKind[kind]
	return e, ok
}

// HasRole reports whether kind is tagged with role in the catalog.
func HasRole(kind int, role string) bool {
	e, ok := Lookup(kind)
	if !ok {
		return false
	}
	for _, r := range e.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// KindsWithRole returns catalog kinds that include role, sorted.
func KindsWithRole(role string) []int {
	load()
	src := byRole[role]
	out := make([]int, len(src))
	copy(out, src)
	return out
}

// KindsWithAnyRole returns unique sorted kinds that have any of the given roles.
func KindsWithAnyRole(roles ...string) []int {
	seen := map[int]struct{}{}
	var out []int
	for _, role := range roles {
		for _, k := range KindsWithRole(role) {
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			out = append(out, k)
		}
	}
	sort.Ints(out)
	return out
}

// Name returns the catalog name, or empty if unknown.
func Name(kind int) string {
	e, ok := Lookup(kind)
	if !ok {
		return ""
	}
	return e.Name
}

// Description returns the catalog description, or empty if unknown.
func Description(kind int) string {
	e, ok := Lookup(kind)
	if !ok {
		return ""
	}
	return e.Description
}

// MarketplaceIndexKinds are kinds Conduit may parse into its listing index.
func MarketplaceIndexKinds() []int {
	return KindsWithAnyRole(RoleStall, RoleProduct, RoleListing, RoleListingDraft, RoleDeletion)
}
