package listing

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const (
	KindDeletion        = 5
	KindStall           = 30017
	KindProduct         = 30018
	KindClassified      = 30402
	KindClassifiedDraft = 30403
	KindStallParam      = 34550
	KindProductParam    = 34560

	StatusActive   = "active"
	StatusInactive = "inactive"

	ReasonNIP99Status  = "nip99_status"
	ReasonZeroQuantity = "zero_quantity"
	ReasonDraft        = "draft"
	ReasonDeleted      = "deleted"

	embedMax = 8192
)

// Event is the subset of a NIP-01 event needed to build a Listing.
type Event struct {
	ID        string
	PubKey    string
	CreatedAt int64
	Kind      int
	Tags      [][]string
	Content   string
}

// Listing is the marketplace aggregate produced by FromEvent.
type Listing struct {
	Coord          string
	EventID        string
	Kind           int
	PubKey         string
	DTag           string
	StallID        string
	Status         string
	InactiveReason string
	Title          string
	Body           string
	Tags           []string
	TextHash       string
	CreatedAt      int64
	Geohash        string
	Lat            float64
	Lon            float64
	HasGeo         bool
	IsDeletion     bool
	DeleteEventIDs []string
	DeleteCoords   []string
}

// CoordOf formats kind:pubkey:d.
func CoordOf(kind int, pubkey, d string) string {
	return strings.TrimSpace(itoa(kind) + ":" + pubkey + ":" + d)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// EmbedText is title, body, and t-tags joined and truncated to 8k.
func (l Listing) EmbedText() string {
	parts := make([]string, 0, 3)
	if l.Title != "" {
		parts = append(parts, l.Title)
	}
	if l.Body != "" {
		parts = append(parts, l.Body)
	}
	if len(l.Tags) > 0 {
		parts = append(parts, strings.Join(l.Tags, " "))
	}
	s := strings.Join(parts, "\n")
	if len(s) > embedMax {
		s = s[:embedMax]
	}
	return s
}

// ComputeTextHash SHA-256s EmbedText.
func (l Listing) ComputeTextHash() string {
	sum := sha256.Sum256([]byte(l.EmbedText()))
	return hex.EncodeToString(sum[:])
}

// DefaultProductKinds are indexed product kinds.
func DefaultProductKinds() []int { return []int{KindProduct, KindProductParam, KindClassified} }

// DefaultStallKinds are indexed stall kinds.
func DefaultStallKinds() []int { return []int{KindStall, KindStallParam} }

// DefaultDraftKinds is kind 30403.
func DefaultDraftKinds() []int { return []int{KindClassifiedDraft} }

// DefaultDeletionKinds is kind 5.
func DefaultDeletionKinds() []int { return []int{KindDeletion} }

func tagValue(tags [][]string, name string) string {
	for _, t := range tags {
		if len(t) >= 2 && t[0] == name {
			return t[1]
		}
	}
	return ""
}

func tagValues(tags [][]string, name string) []string {
	var out []string
	for _, t := range tags {
		if len(t) >= 2 && t[0] == name {
			out = append(out, t[1])
		}
	}
	return out
}
