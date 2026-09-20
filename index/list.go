package index

import (
	"context"
	"fmt"
	"strings"
)

func clampListQuery(limit, offset int) (int, int) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func (s *sqlStore) ListListings(ctx context.Context, q ListQuery) (ListingPage, error) {
	limit, offset := clampListQuery(q.Limit, q.Offset)
	where := "1=1"
	args := []any{}
	n := 1
	status := strings.TrimSpace(q.Status)
	if status == "active" || status == "inactive" {
		where = "l.status = " + s.ph(n)
		args = append(args, status)
		n++
	}
	var total int64
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM listings l WHERE `+where, args...).Scan(&total); err != nil {
		return ListingPage{}, err
	}
	sqlStr := fmt.Sprintf(`SELECT l.coord, l.event_id, l.kind, l.pubkey, l.d_tag, COALESCE(l.stall_id,''), l.status, COALESCE(l.inactive_reason,''), COALESCE(l.title,''), l.created_at, COALESCE(g.geohash,''), CASE WHEN e.coord IS NULL THEN 0 ELSE 1 END
FROM listings l
LEFT JOIN listing_geo g ON g.coord = l.coord
LEFT JOIN listing_embeddings e ON e.coord = l.coord
WHERE %s
ORDER BY l.updated_at DESC, l.coord
LIMIT %s OFFSET %s`, where, s.ph(n), s.ph(n+1))
	args = append(args, limit, offset)
	rows, err := s.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return ListingPage{}, err
	}
	defer rows.Close()
	items := []ListingListItem{}
	for rows.Next() {
		var it ListingListItem
		var hasEmb int
		if err := rows.Scan(&it.Coord, &it.EventID, &it.Kind, &it.PubKey, &it.DTag, &it.StallID, &it.Status, &it.InactiveReason, &it.Title, &it.CreatedAt, &it.Geohash, &hasEmb); err != nil {
			return ListingPage{}, err
		}
		it.HasEmbedding = hasEmb != 0
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return ListingPage{}, err
	}
	return ListingPage{Items: items, Total: total}, nil
}

func (s *sqlStore) ListEmbeddings(ctx context.Context, q ListQuery) (EmbeddingPage, error) {
	limit, offset := clampListQuery(q.Limit, q.Offset)
	var total int64
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM listing_embeddings`).Scan(&total); err != nil {
		return EmbeddingPage{}, err
	}
	sqlStr := fmt.Sprintf(`SELECT e.coord, e.model, e.dim, COALESCE(l.title,''), COALESCE(l.status,''), COALESCE(l.kind, 0), COALESCE(l.event_id,''), COALESCE(l.d_tag,''), COALESCE(l.pubkey,'')
FROM listing_embeddings e
LEFT JOIN listings l ON l.coord = e.coord
ORDER BY e.coord
LIMIT %s OFFSET %s`, s.ph(1), s.ph(2))
	rows, err := s.db.QueryContext(ctx, sqlStr, limit, offset)
	if err != nil {
		return EmbeddingPage{}, err
	}
	defer rows.Close()
	items := []EmbeddingListItem{}
	for rows.Next() {
		var it EmbeddingListItem
		if err := rows.Scan(&it.Coord, &it.Model, &it.Dim, &it.Title, &it.Status, &it.Kind, &it.EventID, &it.DTag, &it.PubKey); err != nil {
			return EmbeddingPage{}, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return EmbeddingPage{}, err
	}
	return EmbeddingPage{Items: items, Total: total}, nil
}
