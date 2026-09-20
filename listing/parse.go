package listing

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/michmich112/conduit-plugin/kinds"
)

// FromEvent parses a stored Nostr event into a Listing. ok=false means skip (malformed or irrelevant).
// Kind roles come from kinds.json (stall / product / listing / listing_draft / deletion).
func FromEvent(ev Event, indexDrafts bool) (Listing, bool) {
	switch {
	case kinds.HasRole(ev.Kind, kinds.RoleDeletion):
		return fromDeletion(ev)
	case kinds.HasRole(ev.Kind, kinds.RoleListing) || kinds.HasRole(ev.Kind, kinds.RoleListingDraft):
		return fromNIP99(ev, indexDrafts)
	case kinds.HasRole(ev.Kind, kinds.RoleProduct):
		return fromProduct(ev)
	case kinds.HasRole(ev.Kind, kinds.RoleStall):
		return fromStall(ev)
	default:
		return Listing{}, false
	}
}

func fromDeletion(ev Event) (Listing, bool) {
	l := Listing{
		EventID:        ev.ID,
		Kind:           KindDeletion,
		PubKey:         ev.PubKey,
		CreatedAt:      ev.CreatedAt,
		IsDeletion:     true,
		Status:         StatusInactive,
		InactiveReason: ReasonDeleted,
	}
	for _, t := range ev.Tags {
		if len(t) < 2 {
			continue
		}
		switch t[0] {
		case "e":
			if t[1] != "" {
				l.DeleteEventIDs = append(l.DeleteEventIDs, t[1])
			}
		case "a":
			kind, pubkey, d, ok := parseATag(t[1])
			if !ok || pubkey != ev.PubKey {
				continue
			}
			l.DeleteCoords = append(l.DeleteCoords, CoordOf(kind, pubkey, d))
		}
	}
	if len(l.DeleteEventIDs) == 0 && len(l.DeleteCoords) == 0 {
		return Listing{}, false
	}
	return l, true
}

func parseATag(v string) (kind int, pubkey, d string, ok bool) {
	parts := strings.SplitN(v, ":", 3)
	if len(parts) != 3 {
		return 0, "", "", false
	}
	k, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", "", false
	}
	if parts[1] == "" || parts[2] == "" {
		return 0, "", "", false
	}
	return k, parts[1], parts[2], true
}

func fromNIP99(ev Event, indexDrafts bool) (Listing, bool) {
	d := tagValue(ev.Tags, "d")
	if d == "" {
		return Listing{}, false
	}
	title := tagValue(ev.Tags, "title")
	summary := tagValue(ev.Tags, "summary")
	body := ev.Content
	if body == "" {
		body = summary
	}
	status := StatusActive
	reason := ""
	st := strings.ToLower(strings.TrimSpace(tagValue(ev.Tags, "status")))
	if st != "" && st != "active" {
		status = StatusInactive
		reason = ReasonNIP99Status
	}
	if ev.Kind == KindClassifiedDraft && !indexDrafts {
		status = StatusInactive
		reason = ReasonDraft
	}
	l := Listing{
		Coord:          CoordOf(ev.Kind, ev.PubKey, d),
		EventID:        ev.ID,
		Kind:           ev.Kind,
		PubKey:         ev.PubKey,
		DTag:           d,
		Status:         status,
		InactiveReason: reason,
		Title:          title,
		Body:           body,
		Tags:           tagValues(ev.Tags, "t"),
		CreatedAt:      ev.CreatedAt,
	}
	applyGeo(&l, tagValue(ev.Tags, "g"))
	l.TextHash = l.ComputeTextHash()
	return l, true
}

type productJSON struct {
	ID          string          `json:"id"`
	StallID     string          `json:"stall_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Quantity    *int            `json:"quantity"`
	Specs       json.RawMessage `json:"specs"`
}

func fromProduct(ev Event) (Listing, bool) {
	var p productJSON
	if strings.TrimSpace(ev.Content) != "" {
		if err := json.Unmarshal([]byte(ev.Content), &p); err != nil {
			return Listing{}, false
		}
	}
	d := tagValue(ev.Tags, "d")
	if d == "" {
		d = p.ID
	}
	if d == "" {
		return Listing{}, false
	}
	status := StatusActive
	reason := ""
	if p.Quantity != nil && *p.Quantity == 0 {
		status = StatusInactive
		reason = ReasonZeroQuantity
	}
	specs := ""
	if len(p.Specs) > 0 && string(p.Specs) != "null" {
		specs = string(p.Specs)
	}
	body := p.Description
	if specs != "" {
		if body != "" {
			body += "\n"
		}
		body += specs
	}
	l := Listing{
		Coord:          CoordOf(ev.Kind, ev.PubKey, d),
		EventID:        ev.ID,
		Kind:           ev.Kind,
		PubKey:         ev.PubKey,
		DTag:           d,
		StallID:        firstNonEmpty(p.StallID, tagValue(ev.Tags, "stall")),
		Status:         status,
		InactiveReason: reason,
		Title:          p.Name,
		Body:           body,
		Tags:           tagValues(ev.Tags, "t"),
		CreatedAt:      ev.CreatedAt,
	}
	applyGeo(&l, tagValue(ev.Tags, "g"))
	l.TextHash = l.ComputeTextHash()
	return l, true
}

type stallJSON struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func fromStall(ev Event) (Listing, bool) {
	var s stallJSON
	if strings.TrimSpace(ev.Content) != "" {
		if err := json.Unmarshal([]byte(ev.Content), &s); err != nil {
			return Listing{}, false
		}
	}
	d := tagValue(ev.Tags, "d")
	if d == "" {
		d = s.ID
	}
	if d == "" {
		return Listing{}, false
	}
	l := Listing{
		Coord:     CoordOf(ev.Kind, ev.PubKey, d),
		EventID:   ev.ID,
		Kind:      ev.Kind,
		PubKey:    ev.PubKey,
		DTag:      d,
		Status:    StatusActive,
		Title:     s.Name,
		Body:      s.Description,
		Tags:      tagValues(ev.Tags, "t"),
		CreatedAt: ev.CreatedAt,
	}
	applyGeo(&l, tagValue(ev.Tags, "g"))
	l.TextHash = l.ComputeTextHash()
	return l, true
}

func applyGeo(l *Listing, gh string) {
	gh = strings.ToLower(strings.TrimSpace(gh))
	if gh == "" {
		return
	}
	lat, lon, ok := DecodeGeohash(gh)
	if !ok {
		return
	}
	l.Geohash = gh
	l.Lat = lat
	l.Lon = lon
	l.HasGeo = true
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
