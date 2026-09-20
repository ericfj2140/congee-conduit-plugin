package listing

import "testing"

const pk = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestFromEventNIP99Active(t *testing.T) {
	l, ok := FromEvent(Event{
		ID: "ev1", PubKey: pk, CreatedAt: 10, Kind: KindClassified, Content: "body text",
		Tags: [][]string{{"d", "item-1"}, {"title", "Red Bike"}, {"g", "9q8yy"}, {"t", "bike"}, {"t", "sport"}},
	}, false)
	if !ok {
		t.Fatal("skip")
	}
	if l.Coord != CoordOf(KindClassified, pk, "item-1") {
		t.Fatalf("coord %s", l.Coord)
	}
	if l.Status != StatusActive || l.Title != "Red Bike" || l.Body != "body text" {
		t.Fatalf("%+v", l)
	}
	if !l.HasGeo || l.Geohash != "9q8yy" {
		t.Fatalf("geo %+v", l)
	}
	if len(l.Tags) != 2 {
		t.Fatalf("tags %v", l.Tags)
	}
	if l.TextHash == "" || l.TextHash != l.ComputeTextHash() {
		t.Fatal("hash")
	}
}

func TestFromEventNIP99InactiveStatus(t *testing.T) {
	l, ok := FromEvent(Event{
		ID: "ev2", PubKey: pk, CreatedAt: 10, Kind: KindClassified, Content: "x",
		Tags: [][]string{{"d", "sold"}, {"status", "sold"}},
	}, false)
	if !ok || l.Status != StatusInactive || l.InactiveReason != ReasonNIP99Status {
		t.Fatalf("%+v ok=%v", l, ok)
	}
}

func TestFromEventDraftSkippedUnlessEnabled(t *testing.T) {
	ev := Event{ID: "ev3", PubKey: pk, CreatedAt: 1, Kind: KindClassifiedDraft, Content: "draft",
		Tags: [][]string{{"d", "d1"}, {"title", "Draft"}}}
	l, ok := FromEvent(ev, false)
	if !ok || l.Status != StatusInactive || l.InactiveReason != ReasonDraft {
		t.Fatalf("draft default %+v ok=%v", l, ok)
	}
	l, ok = FromEvent(ev, true)
	if !ok || l.Status != StatusActive {
		t.Fatalf("index drafts %+v ok=%v", l, ok)
	}
}

func TestFromEventMissingD(t *testing.T) {
	if _, ok := FromEvent(Event{ID: "x", PubKey: pk, Kind: KindClassified, Content: "c"}, false); ok {
		t.Fatal("expected skip")
	}
}

func TestFromEventProductZeroQuantity(t *testing.T) {
	l, ok := FromEvent(Event{
		ID: "p1", PubKey: pk, CreatedAt: 3, Kind: KindProduct,
		Content: `{"id":"sku","name":"Mug","description":"ceramic","stall_id":"s1","quantity":0}`,
		Tags:    [][]string{{"d", "sku"}},
	}, false)
	if !ok || l.Status != StatusInactive || l.InactiveReason != ReasonZeroQuantity {
		t.Fatalf("%+v ok=%v", l, ok)
	}
	if l.StallID != "s1" || l.Title != "Mug" {
		t.Fatalf("%+v", l)
	}
}

func TestFromEventMalformedProductJSON(t *testing.T) {
	if _, ok := FromEvent(Event{ID: "p2", PubKey: pk, Kind: KindProduct, Content: "{not-json", Tags: [][]string{{"d", "x"}}}, false); ok {
		t.Fatal("expected skip")
	}
}

func TestFromEventStall(t *testing.T) {
	l, ok := FromEvent(Event{
		ID: "st1", PubKey: pk, CreatedAt: 4, Kind: KindStall,
		Content: `{"id":"stall-a","name":"Market","description":"downtown"}`,
		Tags:    [][]string{{"d", "stall-a"}, {"g", "9q8"}},
	}, false)
	if !ok || l.Title != "Market" || !l.HasGeo {
		t.Fatalf("%+v ok=%v", l, ok)
	}
}

func TestFromEventKind5EAndASamePubkey(t *testing.T) {
	other := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	l, ok := FromEvent(Event{
		ID: "del1", PubKey: pk, Kind: KindDeletion, CreatedAt: 9,
		Tags: [][]string{
			{"e", "deadbeef"},
			{"a", "30402:" + pk + ":item-1"},
			{"a", "30402:" + other + ":nope"},
		},
	}, false)
	if !ok || !l.IsDeletion {
		t.Fatal("deletion")
	}
	if len(l.DeleteEventIDs) != 1 || l.DeleteEventIDs[0] != "deadbeef" {
		t.Fatalf("e tags %v", l.DeleteEventIDs)
	}
	if len(l.DeleteCoords) != 1 || l.DeleteCoords[0] != CoordOf(KindClassified, pk, "item-1") {
		t.Fatalf("a tags %v", l.DeleteCoords)
	}
}

func TestDecodeGeohashSF(t *testing.T) {
	lat, lon, ok := DecodeGeohash("9q8yy")
	if !ok {
		t.Fatal("decode")
	}
	if lat < 37 || lat > 38 || lon < -123 || lon > -122 {
		t.Fatalf("center %f %f", lat, lon)
	}
}

func TestEncodeGeohashRoundTrip(t *testing.T) {
	hash := EncodeGeohash(37.7749, -122.4194, 5)
	if hash != "9q8yy" {
		t.Fatalf("sf hash %q", hash)
	}
	lat, lon, ok := DecodeGeohash(hash)
	if !ok {
		t.Fatal("decode")
	}
	if HaversineKm(37.7749, -122.4194, lat, lon) > 5 {
		t.Fatalf("roundtrip too far %f %f", lat, lon)
	}
}

func TestGeoMatchPrefix(t *testing.T) {
	if got := GeoMatchPrefix("9q8yy", 2); got != "9q" {
		t.Fatalf("got %q", got)
	}
	if got := GeoMatchPrefix("9q", 2); got != "9q" {
		t.Fatalf("got %q", got)
	}
}

func TestGeoOriginLongestPrefix(t *testing.T) {
	lat, lon, ok := GeoOrigin([]string{"9q", "9q8yy"})
	if !ok {
		t.Fatal("origin")
	}
	sfLat, sfLon, _ := DecodeGeohash("9q8yy")
	if lat != sfLat || lon != sfLon {
		t.Fatalf("got %f %f want %f %f", lat, lon, sfLat, sfLon)
	}
}

func TestHaversineKmSanFranciscoLosAngeles(t *testing.T) {
	d := HaversineKm(37.7749, -122.4194, 34.0522, -118.2437)
	if d < 500 || d > 600 {
		t.Fatalf("sf-la %f km", d)
	}
}
