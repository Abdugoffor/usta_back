package helper

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type AdminListQuery struct {
	Cursor    CursorPayload
	Limit     int
	SortBy    string
	SortOrder string
}

func ParseAdminList(r *http.Request) AdminListQuery {
	cursor, limit := ParseCursorPayload(r)

	return AdminListQuery{
		Cursor:    cursor,
		Limit:     limit,
		SortBy:    strings.TrimSpace(r.URL.Query().Get("sort_by")),
		SortOrder: strings.TrimSpace(r.URL.Query().Get("sort_order")),
	}
}

func (q AdminListQuery) Order() string {
	if strings.EqualFold(q.SortOrder, "asc") {
		return "ASC"
	}

	return "DESC"
}

type SortSpec struct {
	Col  string
	Type string
}

type SortMap map[string]SortSpec

func (m SortMap) Resolve(sortBy string) (SortSpec, string) {
	if sortBy != "" {
		if spec, ok := m[sortBy]; ok {
			return spec, sortBy
		}
	}

	return m["id"], "id"
}

func EscapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)

	s = strings.ReplaceAll(s, `%`, `\%`)

	s = strings.ReplaceAll(s, `_`, `\_`)

	return s
}

func LikePattern(s string) *string {
	s = strings.TrimSpace(s)

	if s == "" {
		return nil
	}

	if len(s) > 200 {
		s = s[:200]
	}

	v := "%" + EscapeLike(s) + "%"

	return &v
}

func QueryBool(r *http.Request, key string) *bool {
	v := r.URL.Query().Get(key)

	if v == "" {
		return nil
	}

	b, err := strconv.ParseBool(v)

	if err != nil {
		return nil
	}

	return &b
}

func QueryInt64(r *http.Request, key string) *int64 {
	v := r.URL.Query().Get(key)

	if v == "" {
		return nil
	}

	n, err := strconv.ParseInt(v, 10, 64)

	if err != nil {
		return nil
	}

	return &n
}

func QueryInt(r *http.Request, key string) *int {
	v := r.URL.Query().Get(key)

	if v == "" {
		return nil
	}

	n, err := strconv.Atoi(v)

	if err != nil {
		return nil
	}

	return &n
}

func QueryString(r *http.Request, key string, max int) string {
	v := strings.TrimSpace(r.URL.Query().Get(key))

	if max > 0 && len(v) > max {
		v = v[:max]
	}

	return v
}

func BuildCursorClause(startIdx int, idCol string, spec SortSpec, orderDir string, cursor CursorPayload) (string, []any) {
	if cursor.ID <= 0 {
		return "", nil
	}

	op := "<"

	if orderDir == "ASC" {
		op = ">"
	}

	if spec.Col == idCol {
		return fmt.Sprintf("%s %s $%d", idCol, op, startIdx), []any{cursor.ID}
	}

	if cursor.Value == "" {
		return fmt.Sprintf("%s %s $%d", idCol, op, startIdx), []any{cursor.ID}
	}

	typeCast := spec.Type

	if typeCast == "" {
		typeCast = "text"
	}

	clause := fmt.Sprintf(
		`(%s %s $%d::%s OR (%s = $%d::%s AND %s %s $%d))`,
		spec.Col, op, startIdx, typeCast,
		spec.Col, startIdx, typeCast,
		idCol, op, startIdx+1,
	)

	return clause, []any{cursor.Value, cursor.ID}
}

func ParseIDList(raw string, max int) []int64 {
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")

	seen := make(map[int64]struct{}, len(parts))

	result := make([]int64, 0, len(parts))

	for _, p := range parts {
		if max > 0 && len(result) >= max {
			break
		}

		p = strings.TrimSpace(p)

		n, err := strconv.ParseInt(p, 10, 64)

		if err != nil || n <= 0 {
			continue
		}

		if _, dup := seen[n]; dup {
			continue
		}

		seen[n] = struct{}{}

		result = append(result, n)
	}

	return result
}
