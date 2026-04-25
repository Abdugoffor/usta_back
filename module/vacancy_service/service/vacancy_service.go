package vacancy_service

import (
	"context"
	"encoding/json"
	"fmt"
	"main_service/helper"
	vacancy_dto "main_service/module/vacancy_service/dto"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VacancyService interface {
	Create(ctx context.Context, userID int64, req vacancy_dto.CreateVacancyRequest) (*vacancy_dto.VacancyResponse, error)

	GetByID(ctx context.Context, id int64) (*vacancy_dto.VacancyResponse, error)

	GetBySlug(ctx context.Context, slug string) (*vacancy_dto.VacancyResponse, error)

	Update(ctx context.Context, id, userID int64, isAdmin bool, req vacancy_dto.UpdateVacancyRequest) (*vacancy_dto.VacancyResponse, error)

	Delete(ctx context.Context, id, userID int64, isAdmin bool) error

	List(ctx context.Context, f vacancy_dto.VacancyFilter, cursor helper.CursorPayload, limit int) ([]*vacancy_dto.VacancyResponse, bool, error)

	Count(ctx context.Context, f vacancy_dto.VacancyFilter) (int64, error)
}

type vacancyService struct {
	db *pgxpool.Pool
}

func NewVacancyService(db *pgxpool.Pool) VacancyService {
	return &vacancyService{db: db}
}

func normalizeCategoryIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}

	seen := make(map[int64]struct{}, len(ids))

	out := make([]int64, 0, len(ids))

	for _, id := range ids {
		if id <= 0 {
			continue
		}

		if _, dup := seen[id]; dup {
			continue
		}

		seen[id] = struct{}{}

		out = append(out, id)
	}

	return out
}

func (s *vacancyService) attachCategories(ctx context.Context, tx pgx.Tx, vacancyID int64, ids []int64) error {
	for _, catID := range normalizeCategoryIDs(ids) {
		if _, err := tx.Exec(ctx, `INSERT INTO category_vacancy (categorya_id, vacancy_id) VALUES ($1, $2) ON CONFLICT (categorya_id, vacancy_id) DO NOTHING`, catID, vacancyID); err != nil {
			return err
		}
	}

	return nil
}

func (s *vacancyService) replaceCategories(ctx context.Context, tx pgx.Tx, vacancyID int64, ids []int64) error {
	if _, err := tx.Exec(ctx, `DELETE FROM category_vacancy WHERE vacancy_id = $1`, vacancyID); err != nil {
		return err
	}

	return s.attachCategories(ctx, tx, vacancyID, ids)
}

func (s *vacancyService) Create(ctx context.Context, userID int64, req vacancy_dto.CreateVacancyRequest) (*vacancy_dto.VacancyResponse, error) {
	isActive := true

	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	tx, err := s.db.Begin(ctx)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	var id int64

	err = tx.QueryRow(ctx, `INSERT INTO vacancies (user_id, region_id, district_id, mahalla_id, adress, name, title, text, contact, price, is_active) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`,
		userID, req.RegionID, req.DistrictID, req.MahallaID,
		req.Adress, req.Name, req.Title, req.Text, req.Contact,
		req.Price, isActive,
	).Scan(&id)

	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `UPDATE vacancies SET slug = $1 WHERE id = $2`, helper.Slug(req.Name, id), id); err != nil {
		return nil, err
	}

	if err := s.attachCategories(ctx, tx, id, req.CategoryIDs); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, id)
}

func (s *vacancyService) GetByID(ctx context.Context, id int64) (*vacancy_dto.VacancyResponse, error) {
	return s.fetchOne(ctx, `WHERE v.id = $1 AND v.deleted_at IS NULL`, id)
}

func (s *vacancyService) GetBySlug(ctx context.Context, slug string) (*vacancy_dto.VacancyResponse, error) {
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

		defer cancel()

		_, _ = s.db.Exec(bgCtx, `UPDATE vacancies SET views_count = COALESCE(views_count, 0) + 1 WHERE slug = $1`, slug)
	}()

	return s.fetchOne(ctx, `WHERE v.slug = $1 AND v.deleted_at IS NULL`, slug)
}

func (s *vacancyService) fetchOne(ctx context.Context, where string, arg any) (*vacancy_dto.VacancyResponse, error) {
	var (
		v       vacancy_dto.VacancyResponse
		catJSON []byte
	)

	err := s.db.QueryRow(ctx, fmt.Sprintf(`SELECT v.id, v.slug, v.user_id, v.region_id, COALESCE(r.name->>'default', ''), v.district_id, COALESCE(d.name->>'default', ''), v.mahalla_id, COALESCE(m.name->>'default', ''), v.adress, v.name, v.title, v.text, v.contact, v.price, v.views_count, v.is_active, v.created_at, v.updated_at, v.deleted_at, COALESCE((SELECT json_agg(json_build_object('id', c.id, 'name', COALESCE(c.name->>'default', ''), 'is_active', COALESCE(c.is_active, FALSE)) ORDER BY c.id) FROM categories c JOIN category_vacancy cv ON cv.categorya_id = c.id WHERE cv.vacancy_id = v.id AND c.deleted_at IS NULL), '[]'::json) AS categories FROM vacancies v LEFT JOIN countries r ON r.id = v.region_id AND r.deleted_at IS NULL LEFT JOIN countries d ON d.id = v.district_id AND d.deleted_at IS NULL LEFT JOIN countries m ON m.id = v.mahalla_id AND m.deleted_at IS NULL %s`, where), arg).Scan(
		&v.ID, &v.Slug, &v.UserID,
		&v.RegionID, &v.RegionName,
		&v.DistrictID, &v.DistrictName,
		&v.MahallaID, &v.MahallaName,
		&v.Adress, &v.Name, &v.Title, &v.Text, &v.Contact,
		&v.Price, &v.ViewsCount, &v.IsActive,
		&v.CreatedAt, &v.UpdatedAt, &v.DeletedAt,
		&catJSON,
	)

	if err != nil {
		return nil, err
	}

	v.Categories = parseCategoryShortList(catJSON)

	return &v, nil
}

func parseCategoryShortList(b []byte) []vacancy_dto.CategoryShort {
	if len(b) == 0 {
		return []vacancy_dto.CategoryShort{}
	}

	var out []vacancy_dto.CategoryShort

	if err := json.Unmarshal(b, &out); err != nil {
		return []vacancy_dto.CategoryShort{}
	}

	if out == nil {
		return []vacancy_dto.CategoryShort{}
	}

	return out
}

func (s *vacancyService) Update(ctx context.Context, id, userID int64, isAdmin bool, req vacancy_dto.UpdateVacancyRequest) (*vacancy_dto.VacancyResponse, error) {
	tx, err := s.db.Begin(ctx)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	args := []any{
		req.RegionID, req.DistrictID, req.MahallaID, req.Adress, req.Name, req.Title,
		req.Text, req.Contact, req.Price, req.IsActive, id,
	}

	ownerCond := ""

	if !isAdmin {
		ownerCond = ` AND user_id = $12`

		args = append(args, userID)
	}

	var retID int64

	if err := tx.QueryRow(ctx, fmt.Sprintf(`UPDATE vacancies SET region_id = COALESCE($1::bigint, region_id), district_id = COALESCE($2::bigint, district_id), mahalla_id = COALESCE($3::bigint, mahalla_id), adress = COALESCE($4::text, adress), name = COALESCE($5::varchar, name), title = COALESCE($6::varchar, title), text = COALESCE($7::text, text), contact = COALESCE($8::varchar, contact), price = COALESCE($9::bigint, price), is_active = COALESCE($10::boolean, is_active), updated_at = NOW() WHERE id = $11 AND deleted_at IS NULL%s RETURNING id`, ownerCond), args...).Scan(&retID); err != nil {
		return nil, fmt.Errorf("vacancy not found or access denied")
	}

	if req.CategoryIDs != nil {
		if err := s.replaceCategories(ctx, tx, retID, req.CategoryIDs); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, retID)
}

func (s *vacancyService) Delete(ctx context.Context, id, userID int64, isAdmin bool) error {
	var (
		tag pgconn.CommandTag
		err error
	)

	if isAdmin {
		tag, err = s.db.Exec(ctx, `UPDATE vacancies SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`, time.Now(), id)
	} else {
		tag, err = s.db.Exec(ctx, `UPDATE vacancies SET deleted_at = $1 WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL`, time.Now(), id, userID)
	}

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("vacancy not found or access denied")
	}

	return nil
}

func vacancyOrderConfig(sortBy, sortOrder string) (string, string, string) {
	switch sortBy {
	case "price":
		if normalizedOrder(sortOrder) == "ASC" {
			return "COALESCE(v.price, 9223372036854775807)", "9223372036854775807", "int64"
		}

		return "COALESCE(v.price, -1)", "-1", "int64"
	default:
		return "v.id", "", ""
	}
}

func normalizedOrder(order string) string {
	if strings.EqualFold(order, "asc") {
		return "ASC"
	}

	return "DESC"
}

func vacancyCursorClause(idx int, orderCol, fallbackValue, valueKind, orderDir string, cursor helper.CursorPayload) (string, []any) {
	if cursor.ID <= 0 {
		return "", nil
	}

	op := "<"

	if orderDir == "ASC" {
		op = ">"
	}

	if orderCol == "v.id" {
		return fmt.Sprintf("v.id %s $%d", op, idx), []any{cursor.ID}
	}

	cursorValue := fallbackValue

	if cursor.Value != "" {
		cursorValue = cursor.Value
	}

	clause := fmt.Sprintf("(%s %s $%d OR (%s = $%d AND v.id %s $%d))", orderCol, op, idx, orderCol, idx, op, idx+1)

	if valueKind == "int64" {
		n, err := strconv.ParseInt(cursorValue, 10, 64)

		if err != nil {
			n, _ = strconv.ParseInt(fallbackValue, 10, 64)
		}

		return clause, []any{n, cursor.ID}
	}

	return clause, []any{cursorValue, cursor.ID}
}

const vacancyListWhere = `v.deleted_at IS NULL AND ($1::bigint IS NULL OR v.user_id = $1) AND ($2::bigint IS NULL OR v.region_id = $2) AND ($3::bigint IS NULL OR v.district_id = $3) AND ($4::bigint IS NULL OR v.mahalla_id = $4) AND ($5::varchar IS NULL OR v.name ILIKE $5) AND ($6::varchar IS NULL OR v.title ILIKE $6) AND ($7::varchar IS NULL OR (v.name ILIKE $7 OR v.title ILIKE $7 OR v.text ILIKE $7)) AND ($8::boolean IS NULL OR v.is_active = $8) AND ($9::bigint IS NULL OR v.price >= $9) AND ($10::bigint IS NULL OR v.price <= $10) AND ($11::bigint IS NULL OR EXISTS (SELECT 1 FROM category_vacancy cv WHERE cv.vacancy_id = v.id AND cv.categorya_id = $11)) AND ($12::bigint[] IS NULL OR EXISTS (SELECT 1 FROM category_vacancy cv WHERE cv.vacancy_id = v.id AND cv.categorya_id = ANY($12)))`

func vacancyListArgs(f vacancy_dto.VacancyFilter) []any {
	var name, title, search *string

	if f.Name != "" {
		v := "%" + helper.EscapeLike(f.Name) + "%"

		name = &v
	}

	if f.Title != "" {
		v := "%" + helper.EscapeLike(f.Title) + "%"

		title = &v
	}

	if f.Search != "" {
		v := "%" + helper.EscapeLike(f.Search) + "%"

		search = &v
	}

	return []any{
		f.UserID, f.RegionID, f.DistrictID, f.MahallaID,
		name, title, search,
		f.IsActive, f.MinPrice, f.MaxPrice,
		f.CategoryID, f.CategoryIDs,
	}
}

func (s *vacancyService) Count(ctx context.Context, f vacancy_dto.VacancyFilter) (int64, error) {
	var total int64

	err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM vacancies v WHERE `+vacancyListWhere, vacancyListArgs(f)...).Scan(&total)

	return total, err
}

func (s *vacancyService) loadCategories(ctx context.Context, ids []int64) (map[int64][]vacancy_dto.CategoryShort, error) {
	out := make(map[int64][]vacancy_dto.CategoryShort, len(ids))

	if len(ids) == 0 {
		return out, nil
	}

	rows, err := s.db.Query(ctx, `SELECT cv.vacancy_id, c.id, COALESCE(c.name->>'default', ''), COALESCE(c.is_active, FALSE) FROM category_vacancy cv JOIN categories c ON c.id = cv.categorya_id WHERE cv.vacancy_id = ANY($1) AND c.deleted_at IS NULL ORDER BY cv.vacancy_id, c.id`, ids)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			vacancyID int64
			cat       vacancy_dto.CategoryShort
		)

		if err := rows.Scan(&vacancyID, &cat.ID, &cat.Name, &cat.IsActive); err != nil {
			return nil, err
		}

		out[vacancyID] = append(out[vacancyID], cat)
	}

	return out, rows.Err()
}

func (s *vacancyService) List(ctx context.Context, f vacancy_dto.VacancyFilter, cursor helper.CursorPayload, limit int) ([]*vacancy_dto.VacancyResponse, bool, error) {
	args := vacancyListArgs(f)

	orderCol, fallbackValue, valueKind := vacancyOrderConfig(f.SortBy, f.SortOrder)

	orderDir := normalizedOrder(f.SortOrder)

	cursorSQL, cursorArgs := vacancyCursorClause(len(args)+1, orderCol, fallbackValue, valueKind, orderDir, cursor)

	if cursorSQL == "" {
		cursorSQL = "TRUE"
	}

	args = append(args, cursorArgs...)

	args = append(args, limit+1)

	rows, err := s.db.Query(ctx, fmt.Sprintf(`SELECT v.id, v.slug, v.user_id, v.region_id, COALESCE(r.name->>'default', ''), v.district_id, COALESCE(d.name->>'default', ''), v.mahalla_id, COALESCE(m.name->>'default', ''), v.adress, v.name, v.title, v.text, v.contact, v.price, v.views_count, v.is_active, v.created_at, v.updated_at, v.deleted_at FROM vacancies v LEFT JOIN countries r ON r.id = v.region_id AND r.deleted_at IS NULL LEFT JOIN countries d ON d.id = v.district_id AND d.deleted_at IS NULL LEFT JOIN countries m ON m.id = v.mahalla_id AND m.deleted_at IS NULL WHERE %s AND %s ORDER BY %s %s, v.id %s LIMIT $%d`,
		vacancyListWhere, cursorSQL, orderCol, orderDir, orderDir, len(args),
	), args...)

	if err != nil {
		return nil, false, err
	}

	defer rows.Close()

	items := make([]*vacancy_dto.VacancyResponse, 0, limit)

	for rows.Next() {
		var v vacancy_dto.VacancyResponse

		if err := rows.Scan(
			&v.ID, &v.Slug, &v.UserID,
			&v.RegionID, &v.RegionName,
			&v.DistrictID, &v.DistrictName,
			&v.MahallaID, &v.MahallaName,
			&v.Adress, &v.Name, &v.Title, &v.Text, &v.Contact,
			&v.Price, &v.ViewsCount, &v.IsActive,
			&v.CreatedAt, &v.UpdatedAt, &v.DeletedAt,
		); err != nil {
			return nil, false, err
		}

		items = append(items, &v)
	}

	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore := len(items) > limit

	if hasMore {
		items = items[:limit]
	}

	ids := make([]int64, 0, len(items))

	for _, v := range items {
		ids = append(ids, v.ID)
	}

	cats, err := s.loadCategories(ctx, ids)

	if err != nil {
		return nil, false, err
	}

	for _, v := range items {
		v.Categories = cats[v.ID]

		if v.Categories == nil {
			v.Categories = []vacancy_dto.CategoryShort{}
		}
	}

	return items, hasMore, nil
}
