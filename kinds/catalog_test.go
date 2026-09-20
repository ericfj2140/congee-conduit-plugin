package kinds

import (
	"os"
	"testing"
)

func TestCatalogCommunityIsNotStall(t *testing.T) {
	if err := LoadError(); err != nil {
		t.Fatal(err)
	}
	if !HasRole(34550, RoleCommunity) {
		t.Fatal("34550 should be community")
	}
	if HasRole(34550, RoleStall) || HasRole(34550, RoleProduct) {
		t.Fatal("34550 must not be stall or product")
	}
	if HasRole(34560, RoleProduct) || HasRole(34560, RoleStall) {
		t.Fatal("34560 must not be treated as NIP-15 product or stall")
	}
	stalls := KindsWithRole(RoleStall)
	if len(stalls) != 1 || stalls[0] != 30017 {
		t.Fatalf("stall kinds %v", stalls)
	}
	products := KindsWithRole(RoleProduct)
	if len(products) != 1 || products[0] != 30018 {
		t.Fatalf("product kinds %v", products)
	}
	if !HasRole(30402, RoleListing) || !HasRole(30403, RoleListingDraft) || !HasRole(5, RoleDeletion) {
		t.Fatal("expected NIP-99 and NIP-09 roles")
	}
}

func TestMarketplaceIndexKindsOmitsCommunity(t *testing.T) {
	for _, k := range MarketplaceIndexKinds() {
		if k == 34550 || k == 34560 {
			t.Fatalf("marketplace index includes unofficial/community kind %d", k)
		}
	}
}

func TestRootKindsJSONMatchesEmbed(t *testing.T) {
	root, err := os.ReadFile("../kinds.json")
	if err != nil {
		t.Fatal(err)
	}
	if string(root) != string(embeddedJSON) {
		t.Fatal("kinds.json at plugin root must match kinds/kinds.json (go:embed)")
	}
}
