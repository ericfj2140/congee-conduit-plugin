package index

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/michmich112/conduit-plugin/embed"
	"github.com/michmich112/conduit-plugin/listing"
)

const pk = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestTursoSearchRankAndInactive(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "conduit-index.db")
	st, err := OpenTurso(ctx, path, embed.Fake{})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	bike, ok := listing.FromEvent(listing.Event{
		ID: "id-bike", PubKey: pk, CreatedAt: 20, Kind: listing.KindClassified,
		Content: "a red bicycle for trails",
		Tags:    [][]string{{"d", "bike"}, {"title", "Red Bicycle"}, {"g", "9q8yy"}},
	}, false)
	if !ok {
		t.Fatal("bike")
	}
	pizza, ok := listing.FromEvent(listing.Event{
		ID: "id-pizza", PubKey: pk, CreatedAt: 30, Kind: listing.KindClassified,
		Content: "wood fired pizza slice",
		Tags:    [][]string{{"d", "food"}, {"title", "Pizza"}},
	}, false)
	if !ok {
		t.Fatal("pizza")
	}
	if err := st.Upsert(ctx, bike); err != nil {
		t.Fatal(err)
	}
	if err := st.Upsert(ctx, pizza); err != nil {
		t.Fatal(err)
	}
	ids, err := st.Search(ctx, Query{
		Search: "bicycle", Kinds: []int{listing.KindClassified}, Limit: 10,
		ActiveOnly: true, VectorEnabled: true, SearchCandidateCap: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) == 0 || ids[0] != "id-bike" {
		t.Fatalf("rank %v", ids)
	}

	sold, ok := listing.FromEvent(listing.Event{
		ID: "id-bike-2", PubKey: pk, CreatedAt: 40, Kind: listing.KindClassified,
		Content: "a red bicycle for trails",
		Tags:    [][]string{{"d", "bike"}, {"title", "Red Bicycle"}, {"status", "sold"}},
	}, false)
	if !ok {
		t.Fatal("sold")
	}
	if err := st.Upsert(ctx, sold); err != nil {
		t.Fatal(err)
	}
	ids, err = st.Search(ctx, Query{
		Search: "bicycle", Kinds: []int{listing.KindClassified}, Limit: 10,
		ActiveOnly: true, VectorEnabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range ids {
		if id == "id-bike" || id == "id-bike-2" {
			t.Fatalf("inactive leaked %v", ids)
		}
	}

	geo, err := st.Search(ctx, Query{
		Kinds: []int{listing.KindClassified}, GeoPrefixes: []string{"9q8"}, Limit: 10,
		ActiveOnly: true, GeoEnabled: true, GeoMinPrefixLen: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = geo

	page, err := st.ListListings(ctx, ListQuery{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 2 || len(page.Items) != 2 {
		t.Fatalf("listings %+v", page)
	}
	emb, err := st.ListEmbeddings(ctx, ListQuery{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatal(err)
	}
	if emb.Total < 1 {
		t.Fatalf("embeddings %+v", emb)
	}
}

func TestTursoSearchGeoProximity(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "conduit-geo.db")
	st, err := OpenTurso(ctx, path, embed.Fake{})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	cities := []struct {
		id        string
		name      string
		lat, lon  float64
		createdAt int64
	}{
		{"sf", "San Francisco", 37.7749, -122.4194, 2010},
		{"oak", "Oakland", 37.8044, -122.2712, 2020},
		{"sj", "San Jose", 37.3382, -121.8863, 2030},
		{"sac", "Sacramento", 38.5816, -121.4944, 2040},
		{"la", "Los Angeles", 34.0522, -118.2437, 2050},
		{"lb", "Long Beach", 33.7701, -118.1937, 2060},
		{"fre", "Fresno", 36.7378, -119.7871, 2070},
		{"sb", "Santa Barbara", 34.4208, -119.6982, 2080},
		{"lv", "Las Vegas", 36.1699, -115.1398, 2090},
		{"bak", "Bakersfield", 35.3733, -119.0187, 2095},
		{"nyc", "New York", 40.7128, -74.0060, 2100},
	}
	hashes := map[string]string{}
	for _, c := range cities {
		gh := listing.EncodeGeohash(c.lat, c.lon, 5)
		hashes[c.id] = gh
		if c.id != "nyc" && listing.GeoMatchPrefix(gh, 2) != "9q" {
			t.Fatalf("%s hash %q is not in 9q", c.name, gh)
		}
		l, ok := listing.FromEvent(listing.Event{
			ID: "id-" + c.id, PubKey: pk, CreatedAt: c.createdAt, Kind: listing.KindClassified,
			Content: c.name + " classified listing",
			Tags:    [][]string{{"d", c.id}, {"title", c.name + " listing"}, {"g", gh}},
		}, false)
		if !ok {
			t.Fatalf("parse %s", c.id)
		}
		if err := st.Upsert(ctx, l); err != nil {
			t.Fatal(err)
		}
	}

	requesters := []string{"sf", "la", "sac", "lb", "lv"}
	for _, rid := range requesters {
		ids, err := st.Search(ctx, Query{
			Kinds: []int{listing.KindClassified}, GeoPrefixes: []string{hashes[rid]}, Limit: 20,
			ActiveOnly: true, GeoEnabled: true, GeoMinPrefixLen: 2,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(ids) < 5 {
			t.Fatalf("%s: want >=5 hits, got %d %v", rid, len(ids), ids)
		}
		for _, id := range ids {
			if id == "id-nyc" {
				t.Fatalf("%s: NYC leaked into 9q filter %v", rid, ids)
			}
		}
		if ids[0] != "id-"+rid {
			t.Fatalf("%s: closest should be self, got %v", rid, ids)
		}
		oLat, oLon, ok := listing.DecodeGeohash(hashes[rid])
		if !ok {
			t.Fatal("origin")
		}
		prev := -1.0
		for _, id := range ids {
			var gh string
			for cid, h := range hashes {
				if "id-"+cid == id {
					gh = h
					break
				}
			}
			lat, lon, ok := listing.DecodeGeohash(gh)
			if !ok {
				t.Fatalf("decode %s", id)
			}
			d := listing.HaversineKm(oLat, oLon, lat, lon)
			if prev >= 0 && d+0.01 < prev {
				t.Fatalf("%s: not sorted by distance: %v at %f after %f", rid, ids, d, prev)
			}
			prev = d
		}
	}
}
