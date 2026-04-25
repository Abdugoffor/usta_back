package helper

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
)

type CursorMeta struct {
	Total      int64  `json:"total,omitempty"`
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

type CursorPayload struct {
	ID    int64  `json:"id"`
	Value string `json:"value,omitempty"`
}

func EncodeCursor(id int64) string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.FormatInt(id, 10)))
}

func EncodeCursorPayload(payload CursorPayload) string {
	b, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}

func DecodeCursor(s string) (int64, bool) {
	if s == "" {
		return 0, false
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return 0, false
	}
	id, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func DecodeCursorPayload(s string) (CursorPayload, bool) {
	if s == "" {
		return CursorPayload{}, false
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return CursorPayload{}, false
	}

	var payload CursorPayload
	if err := json.Unmarshal(b, &payload); err == nil && payload.ID > 0 {
		return payload, true
	}

	id, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil || id <= 0 {
		return CursorPayload{}, false
	}
	return CursorPayload{ID: id}, true
}

func ParseCursorPage(r *http.Request) (afterID int64, limit int) {
	cursor := r.URL.Query().Get("cursor")
	afterID, _ = DecodeCursor(cursor)

	limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 200 {
		limit = 10
	}
	return
}

func ParseCursorPayload(r *http.Request) (CursorPayload, int) {
	cursor, ok := DecodeCursorPayload(r.URL.Query().Get("cursor"))

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}
	if !ok {
		return CursorPayload{}, limit
	}
	return cursor, limit
}

func NewCursorMeta(limit int, hasMore bool, lastID int64, total int64) CursorMeta {
	meta := CursorMeta{Total: total, Limit: limit, HasMore: hasMore}
	if hasMore {
		meta.NextCursor = EncodeCursor(lastID)
	}
	return meta
}

func NewCursorMetaWithValue(limit int, hasMore bool, lastID int64, lastValue string, total int64) CursorMeta {
	meta := CursorMeta{Total: total, Limit: limit, HasMore: hasMore}
	if hasMore {
		meta.NextCursor = EncodeCursorPayload(CursorPayload{ID: lastID, Value: lastValue})
	}
	return meta
}
