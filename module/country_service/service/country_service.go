package country_service

import (
	"context"
	"encoding/json"
	"fmt"
	"main_service/helper"
	country_dto "main_service/module/country_service/dto"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var CountrySortMap = helper.SortMap{
	"id":         {Col: "id", Type: "bigint"},
	"name":       {Col: `(name->>'default')`, Type: "text"},
	"parent_id":  {Col: "parent_id", Type: "bigint"},
	"is_active":  {Col: "is_active", Type: "boolean"},
	"created_at": {Col: "created_at", Type: "timestamptz"},
	"updated_at": {Col: "updated_at", Type: "timestamptz"},
}

type CountryService interface {
	Create(ctx context.Context, req country_dto.CreateCountryRequest) (*country_dto.CountryResponse, error)

	GetByID(ctx context.Context, id int64) (*country_dto.CountryResponse, error)

	Update(ctx context.Context, id int64, req country_dto.UpdateCountryRequest) (*country_dto.CountryResponse, error)

	Delete(ctx context.Context, id int64) error

	List(ctx context.Context, f country_dto.CountryFilter, q helper.AdminListQuery) ([]*country_dto.CountryResponse, bool, error)

	Count(ctx context.Context, f country_dto.CountryFilter) (int64, error)

	ListActive(ctx context.Context, parentID int64, lang string) ([]*country_dto.CountryActiveResponse, error)

	GetActiveLanguages(ctx context.Context) ([]string, error)
}

type countryService struct {
	db *pgxpool.Pool
}

func NewCountryService(db *pgxpool.Pool) CountryService {
	return &countryService{db: db}
}

func (s *countryService) GetActiveLanguages(ctx context.Context) ([]string, error) {
	return helper.FetchActiveLanguages(ctx, s.db)
}

func unmarshalName(b []byte) map[string]string {
	if len(b) == 0 {
		return map[string]string{}
	}

	var m map[string]string

	if err := json.Unmarshal(b, &m); err != nil {
		return map[string]string{}
	}

	return m
}

func (s *countryService) Create(ctx context.Context, req country_dto.CreateCountryRequest) (*country_dto.CountryResponse, error) {
	isActive := true

	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	nameJSON, err := json.Marshal(helper.NormalizeMultilang(req.Name))

	if err != nil {
		return nil, err
	}

	var (
		r         country_dto.CountryResponse
		nameBytes []byte
	)

	err = s.db.QueryRow(ctx, `INSERT INTO countries (parent_id, name, is_active) VALUES ($1, $2::jsonb, $3) RETURNING id, parent_id, name, COALESCE(is_active, FALSE), created_at, updated_at, deleted_at`,
		req.ParentID, string(nameJSON), isActive,
	).Scan(&r.ID, &r.ParentID, &nameBytes, &r.IsActive, &r.CreatedAt, &r.UpdatedAt, &r.DeletedAt)

	if err != nil {
		return nil, err
	}

	r.Name = unmarshalName(nameBytes)

	return &r, nil
}

func (s *countryService) GetByID(ctx context.Context, id int64) (*country_dto.CountryResponse, error) {
	var (
		r         country_dto.CountryResponse
		nameBytes []byte
	)

	err := s.db.QueryRow(ctx, `SELECT id, parent_id, name, COALESCE(is_active, FALSE), created_at, updated_at, deleted_at FROM countries WHERE id = $1 AND deleted_at IS NULL`,
		id,
	).Scan(&r.ID, &r.ParentID, &nameBytes, &r.IsActive, &r.CreatedAt, &r.UpdatedAt, &r.DeletedAt)

	if err != nil {
		return nil, err
	}

	r.Name = unmarshalName(nameBytes)

	return &r, nil
}

func (s *countryService) Update(ctx context.Context, id int64, req country_dto.UpdateCountryRequest) (*country_dto.CountryResponse, error) {
	var nameJSON *string

	if req.Name != nil {
		b, err := json.Marshal(helper.NormalizeMultilang(*req.Name))

		if err != nil {
			return nil, err
		}

		v := string(b)

		nameJSON = &v
	}

	var (
		r         country_dto.CountryResponse
		nameBytes []byte
	)

	err := s.db.QueryRow(ctx, `UPDATE countries SET parent_id = COALESCE($1::bigint, parent_id), name = COALESCE($2::jsonb, name), is_active = COALESCE($3::boolean, is_active), updated_at = NOW() WHERE id = $4 AND deleted_at IS NULL RETURNING id, parent_id, name, COALESCE(is_active, FALSE), created_at, updated_at, deleted_at`,
		req.ParentID, nameJSON, req.IsActive, id,
	).Scan(&r.ID, &r.ParentID, &nameBytes, &r.IsActive, &r.CreatedAt, &r.UpdatedAt, &r.DeletedAt)

	if err != nil {
		return nil, fmt.Errorf("country not found")
	}

	r.Name = unmarshalName(nameBytes)

	return &r, nil
}

func (s *countryService) Delete(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `UPDATE countries SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`, time.Now(), id)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("country not found")
	}

	return nil
}

const countryListWhere = `c.deleted_at IS NULL AND ($1::bigint IS NULL OR c.parent_id = $1) AND ($2::boolean IS NULL OR (($2 = TRUE AND c.parent_id IS NOT NULL) OR ($2 = FALSE AND c.parent_id IS NULL))) AND ($3::text IS NULL OR c.name::text ILIKE $3) AND ($4::boolean IS NULL OR c.is_active = $4)`

func countryListArgs(f country_dto.CountryFilter) []any {
	return []any{f.ParentID, f.HasParent, helper.LikePattern(f.Name), f.IsActive}
}

func (s *countryService) Count(ctx context.Context, f country_dto.CountryFilter) (int64, error) {
	var total int64

	err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM countries c WHERE `+countryListWhere, countryListArgs(f)...).Scan(&total)

	return total, err
}

func (s *countryService) ListActive(ctx context.Context, parentID int64, lang string) ([]*country_dto.CountryActiveResponse, error) {
	rows, err := s.db.Query(ctx, `SELECT id, parent_id, name FROM countries WHERE deleted_at IS NULL AND is_active = TRUE AND parent_id = $1 ORDER BY (name->>'default') ASC NULLS LAST, id ASC`, parentID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	lang = strings.ToLower(strings.TrimSpace(lang))

	items := make([]*country_dto.CountryActiveResponse, 0)

	for rows.Next() {
		var (
			id        int64
			parent    *int64
			nameBytes []byte
		)

		if err := rows.Scan(&id, &parent, &nameBytes); err != nil {
			return nil, err
		}

		name := unmarshalName(nameBytes)

		value := ""

		if lang != "" {
			if v, ok := name[lang]; ok && strings.TrimSpace(v) != "" {
				value = v
			}
		}

		if value == "" {
			value = strings.TrimSpace(name["default"])
		}

		items = append(items, &country_dto.CountryActiveResponse{ID: id, ParentID: parent, Name: value})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (s *countryService) List(ctx context.Context, f country_dto.CountryFilter, q helper.AdminListQuery) ([]*country_dto.CountryResponse, bool, error) {
	spec, _ := CountrySortMap.Resolve(q.SortBy)

	orderDir := q.Order()

	args := countryListArgs(f)

	cursorSQL, cursorArgs := helper.BuildCursorClause(len(args)+1, "id", spec, orderDir, q.Cursor)

	if cursorSQL == "" {
		cursorSQL = "TRUE"
	}

	args = append(args, cursorArgs...)

	args = append(args, q.Limit+1)

	rows, err := s.db.Query(ctx, fmt.Sprintf(`SELECT id, parent_id, name, is_active, created_at, updated_at, deleted_at FROM (SELECT c.id, c.parent_id, c.name, COALESCE(c.is_active, FALSE) AS is_active, c.created_at, c.updated_at, c.deleted_at FROM countries c WHERE %s) c WHERE %s ORDER BY %s %s NULLS LAST, id %s LIMIT $%d`,
		countryListWhere, cursorSQL, spec.Col, orderDir, orderDir, len(args),
	), args...)

	if err != nil {
		return nil, false, err
	}

	defer rows.Close()

	items := make([]*country_dto.CountryResponse, 0, q.Limit)

	for rows.Next() {
		var (
			r         country_dto.CountryResponse
			nameBytes []byte
		)

		if err := rows.Scan(&r.ID, &r.ParentID, &nameBytes, &r.IsActive, &r.CreatedAt, &r.UpdatedAt, &r.DeletedAt); err != nil {
			return nil, false, err
		}

		r.Name = unmarshalName(nameBytes)

		items = append(items, &r)
	}

	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore := len(items) > q.Limit

	if hasMore {
		items = items[:q.Limit]
	}

	return items, hasMore, nil
}
