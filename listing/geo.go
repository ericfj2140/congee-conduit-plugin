package listing

import (
	"math"
	"strings"
)

const geohashAlphabet = "0123456789bcdefghjkmnpqrstuvwxyz"

// DecodeGeohash returns the center lat/lon of a geohash. ok=false if invalid.
func DecodeGeohash(hash string) (lat, lon float64, ok bool) {
	hash = strings.ToLower(strings.TrimSpace(hash))
	if hash == "" {
		return 0, 0, false
	}
	latMin, latMax := -90.0, 90.0
	lonMin, lonMax := -180.0, 180.0
	even := true
	for _, r := range hash {
		idx := strings.IndexRune(geohashAlphabet, r)
		if idx < 0 {
			return 0, 0, false
		}
		for bits := 4; bits >= 0; bits-- {
			bit := (idx >> bits) & 1
			if even {
				mid := (lonMin + lonMax) / 2
				if bit == 1 {
					lonMin = mid
				} else {
					lonMax = mid
				}
			} else {
				mid := (latMin + latMax) / 2
				if bit == 1 {
					latMin = mid
				} else {
					latMax = mid
				}
			}
			even = !even
		}
	}
	return (latMin + latMax) / 2, (lonMin + lonMax) / 2, true
}

// EncodeGeohash encodes lat/lon to a geohash of the given precision (default 5).
func EncodeGeohash(lat, lon float64, precision int) string {
	if precision <= 0 {
		precision = 5
	}
	latMin, latMax := -90.0, 90.0
	lonMin, lonMax := -180.0, 180.0
	even := true
	bit, ch := 0, 0
	var b strings.Builder
	b.Grow(precision)
	for b.Len() < precision {
		if even {
			mid := (lonMin + lonMax) / 2
			if lon >= mid {
				ch |= 1 << (4 - bit)
				lonMin = mid
			} else {
				lonMax = mid
			}
		} else {
			mid := (latMin + latMax) / 2
			if lat >= mid {
				ch |= 1 << (4 - bit)
				latMin = mid
			} else {
				latMax = mid
			}
		}
		even = !even
		bit++
		if bit == 5 {
			b.WriteByte(geohashAlphabet[ch])
			bit = 0
			ch = 0
		}
	}
	return b.String()
}

// HaversineKm is great-circle distance in kilometers.
func HaversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const r = 6371.0
	p1 := lat1 * math.Pi / 180
	p2 := lat2 * math.Pi / 180
	dlat := (lat2 - lat1) * math.Pi / 180
	dlon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(p1)*math.Cos(p2)*math.Sin(dlon/2)*math.Sin(dlon/2)
	return 2 * r * math.Asin(math.Min(1, math.Sqrt(a)))
}

// GeoMatchPrefix is the cell used for `#g` LIKE matching. Longer hashes are
// truncated to minLen so a requester pin (e.g. 9q8yy) still matches the region.
func GeoMatchPrefix(pfx string, minLen int) string {
	pfx = strings.ToLower(strings.TrimSpace(pfx))
	if minLen <= 0 {
		minLen = 2
	}
	if len(pfx) <= minLen {
		return pfx
	}
	return pfx[:minLen]
}

// GeoOrigin is the center of the longest (most precise) geohash in prefixes.
func GeoOrigin(prefixes []string) (lat, lon float64, ok bool) {
	best := ""
	for _, p := range prefixes {
		p = strings.ToLower(strings.TrimSpace(p))
		if len(p) > len(best) {
			best = p
		}
	}
	if best == "" {
		return 0, 0, false
	}
	return DecodeGeohash(best)
}
