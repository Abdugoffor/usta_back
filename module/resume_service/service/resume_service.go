package resume_service

import (
	"context"
	"encoding/json"
	"fmt"
	"main_service/helper"
	resume_dto "main_service/module/resume_service/dto"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ResumeService interface {
	Create(ctx context.Context, userID int64, req resume_dto.CreateResumeRequest) (*resume_dto.ResumeResponse, error)

	GetByID(ctx context.Context, id int64) (*resume_dto.ResumeResponse, error)

	GetBySlug(ctx context.Context, slug string) (*resume_dto.ResumeResponse, error)

	Update(ctx context.Context, id, userID int64, isAdmin bool, req resume_dto.UpdateResumeRequest) (*resume_dto.ResumeResponse, error)

	Delete(ctx context.Context, id, userID int64, isAdmin bool) error

	List(ctx context.Context, f resume_dto.ResumeFilter, cursor helper.CursorPayload, limit int) ([]*resume_dto.ResumeResponse, bool, error)

	Count(ctx context.Context, f resume_dto.ResumeFilter) (int64, error)
}

type resumeService struct {
	db *pgxpool.Pool
}

func NewResumeService(db *pgxpool.Pool) ResumeService {
	return &resumeService{db: db}
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

func (s *resumeService) attachCategories(ctx context.Context, tx pgx.Tx, resumeID int64, ids []int64) error {
	for _, catID := range normalizeCategoryIDs(ids) {
		if _, err := tx.Exec(ctx, `INSERT INTO category_resume (categorya_id, resume_id) VALUES ($1, $2) ON CONFLICT (categorya_id, resume_id) DO NOTHING`, catID, resumeID); err != nil {
			return err
		}
	}

	return nil
}

func (s *resumeService) replaceCategories(ctx context.Context, tx pgx.Tx, resumeID int64, ids []int64) error {
	if _, err := tx.Exec(ctx, `DELETE FROM category_resume WHERE resume_id = $1`, resumeID); err != nil {
		return err
	}

	return s.attachCategories(ctx, tx, resumeID, ids)
}

func (s *resumeService) Create(ctx context.Context, userID int64, req resume_dto.CreateResumeRequest) (*resume_dto.ResumeResponse, error) {
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

	err = tx.QueryRow(ctx, `INSERT INTO resumes (user_id, region_id, district_id, mahalla_id, adress, name, photo, title, text, contact, price, experience_year, skills, is_active) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING id`,
		userID, req.RegionID, req.DistrictID, req.MahallaID,
		req.Adress, req.Name, req.Photo, req.Title, req.Text, req.Contact,
		req.Price, req.ExperienceYear, req.Skills, isActive,
	).Scan(&id)

	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `UPDATE resumes SET slug = $1 WHERE id = $2`, helper.Slug(req.Name, id), id); err != nil {
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

func (s *resumeService) GetByID(ctx context.Context, id int64) (*resume_dto.ResumeResponse, error) {
	return s.fetchOne(ctx, `WHERE rs.id = $1 AND rs.deleted_at IS NULL`, id)
}

func (s *resumeService) GetBySlug(ctx context.Context, slug string) (*resume_dto.ResumeResponse, error) {
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

		defer cancel()

		_, _ = s.db.Exec(bgCtx, `UPDATE resumes SET views_count = COALESCE(views_count, 0) + 1 WHERE slug = $1`, slug)
	}()

	return s.fetchOne(ctx, `WHERE rs.slug = $1 AND rs.deleted_at IS NULL`, slug)
}

func (s *resumeService) fetchOne(ctx context.Context, where string, arg any) (*resume_dto.ResumeResponse, error) {
	var (
		r       resume_dto.ResumeResponse
		catJSON []byte
	)

	err := s.db.QueryRow(ctx, fmt.Sprintf(`SELECT rs.id, rs.slug, rs.user_id, rs.region_id, COALESCE(r.name->>'default', ''), rs.district_id, COALESCE(d.name->>'default', ''), rs.mahalla_id, COALESCE(m.name->>'default', ''), rs.adress, rs.name, rs.photo, rs.title, rs.text, rs.contact, rs.price, rs.experience_year, rs.skills, rs.views_count, rs.is_active, rs.created_at, rs.updated_at, rs.deleted_at, COALESCE((SELECT json_agg(json_build_object('id', c.id, 'name', COALESCE(c.name->>'default', ''), 'is_active', COALESCE(c.is_active, FALSE)) ORDER BY c.id) FROM categories c JOIN category_resume cr ON cr.categorya_id = c.id WHERE cr.resume_id = rs.id AND c.deleted_at IS NULL), '[]'::json) AS categories FROM resumes rs LEFT JOIN countries r ON r.id = rs.region_id AND r.deleted_at IS NULL LEFT JOIN countries d ON d.id = rs.district_id AND d.deleted_at IS NULL LEFT JOIN countries m ON m.id = rs.mahalla_id AND m.deleted_at IS NULL %s`, where), arg).Scan(
		&r.ID, &r.Slug, &r.UserID,
		&r.RegionID, &r.RegionName,
		&r.DistrictID, &r.DistrictName,
		&r.MahallaID, &r.MahallaName,
		&r.Adress, &r.Name, &r.Photo, &r.Title, &r.Text, &r.Contact,
		&r.Price, &r.ExperienceYear, &r.Skills,
		&r.ViewsCount, &r.IsActive,
		&r.CreatedAt, &r.UpdatedAt, &r.DeletedAt,
		&catJSON,
	)

	if err != nil {
		return nil, err
	}

	r.Categories = parseCategoryShortList(catJSON)

	return &r, nil
}

func parseCategoryShortList(b []byte) []resume_dto.CategoryShort {
	if len(b) == 0 {
		return []resume_dto.CategoryShort{}
	}

	var out []resume_dto.CategoryShort

	if err := json.Unmarshal(b, &out); err != nil {
		return []resume_dto.CategoryShort{}
	}

	if out == nil {
		return []resume_dto.CategoryShort{}
	}

	return out
}

func (s *resumeService) Update(ctx context.Context, id, userID int64, isAdmin bool, req resume_dto.UpdateResumeRequest) (*resume_dto.ResumeResponse, error) {
	tx, err := s.db.Begin(ctx)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	args := []any{
		req.RegionID, req.DistrictID, req.MahallaID, req.Adress, req.Name, req.Photo,
		req.Title, req.Text, req.Contact, req.Price, req.ExperienceYear, req.Skills,
		req.IsActive, id,
	}

	ownerCond := ""

	if !isAdmin {
		ownerCond = ` AND user_id = $15`

		args = append(args, userID)
	}

	var retID int64

	if err := tx.QueryRow(ctx, fmt.Sprintf(`UPDATE resumes SET region_id = COALESCE($1::bigint, region_id), district_id = COALESCE($2::bigint, district_id), mahalla_id = COALESCE($3::bigint, mahalla_id), adress = COALESCE($4::text, adress), name = COALESCE($5::varchar, name), photo = COALESCE($6::text, photo), title = COALESCE($7::varchar, title), text = COALESCE($8::text, text), contact = COALESCE($9::varchar, contact), price = COALESCE($10::bigint, price), experience_year = COALESCE($11::int, experience_year), skills = COALESCE($12::text, skills), is_active = COALESCE($13::boolean, is_active), updated_at = NOW() WHERE id = $14 AND deleted_at IS NULL%s RETURNING id`, ownerCond), args...).Scan(&retID); err != nil {
		return nil, fmt.Errorf("resume not found or access denied")
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

func (s *resumeService) Delete(ctx context.Context, id, userID int64, isAdmin bool) error {
	var (
		tag pgconn.CommandTag
		err error
	)

	if isAdmin {
		tag, err = s.db.Exec(ctx, `UPDATE resumes SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`, time.Now(), id)
	} else {
		tag, err = s.db.Exec(ctx, `UPDATE resumes SET deleted_at = $1 WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL`, time.Now(), id, userID)
	}

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("resume not found or access denied")
	}

	return nil
}

func resumeOrderConfig(sortBy, sortOrder string) (string, string, string) {
	switch sortBy {
	case "price":
		if normalizedOrder(sortOrder) == "ASC" {
			return "COALESCE(rs.price, 9223372036854775807)", "9223372036854775807", "int64"
		}

		return "COALESCE(rs.price, -1)", "-1", "int64"
	case "experience_year":
		if normalizedOrder(sortOrder) == "ASC" {
			return "COALESCE(rs.experience_year, 2147483647)", "2147483647", "int"
		}

		return "COALESCE(rs.experience_year, -1)", "-1", "int"
	default:
		return "rs.id", "", ""
	}
}

func normalizedOrder(order string) string {
	if strings.EqualFold(order, "asc") {
		return "ASC"
	}

	return "DESC"
}

func resumeCursorClause(idx int, orderCol, fallbackValue, valueKind, orderDir string, cursor helper.CursorPayload) (string, []any) {
	if cursor.ID <= 0 {
		return "", nil
	}

	op := "<"

	if orderDir == "ASC" {
		op = ">"
	}

	if orderCol == "rs.id" {
		return fmt.Sprintf("rs.id %s $%d", op, idx), []any{cursor.ID}
	}

	cursorValue := fallbackValue

	if cursor.Value != "" {
		cursorValue = cursor.Value
	}

	clause := fmt.Sprintf("(%s %s $%d OR (%s = $%d AND rs.id %s $%d))", orderCol, op, idx, orderCol, idx, op, idx+1)

	switch valueKind {
	case "int64":
		n, err := strconv.ParseInt(cursorValue, 10, 64)

		if err != nil {
			n, _ = strconv.ParseInt(fallbackValue, 10, 64)
		}

		return clause, []any{n, cursor.ID}
	case "int":
		n, err := strconv.Atoi(cursorValue)

		if err != nil {
			n, _ = strconv.Atoi(fallbackValue)
		}

		return clause, []any{n, cursor.ID}
	}

	return clause, []any{cursorValue, cursor.ID}
}

const resumeListWhere = `rs.deleted_at IS NULL AND ($1::bigint IS NULL OR rs.user_id = $1) AND ($2::bigint IS NULL OR rs.region_id = $2) AND ($3::bigint IS NULL OR rs.district_id = $3) AND ($4::bigint IS NULL OR rs.mahalla_id = $4) AND ($5::varchar IS NULL OR rs.name ILIKE $5) AND ($6::varchar IS NULL OR rs.title ILIKE $6) AND ($7::varchar IS NULL OR (rs.name ILIKE $7 OR rs.title ILIKE $7 OR rs.skills ILIKE $7 OR rs.text ILIKE $7)) AND ($8::boolean IS NULL OR rs.is_active = $8) AND ($9::bigint IS NULL OR rs.price >= $9) AND ($10::bigint IS NULL OR rs.price <= $10) AND ($11::int IS NULL OR rs.experience_year >= $11) AND ($12::bigint IS NULL OR EXISTS (SELECT 1 FROM category_resume cr WHERE cr.resume_id = rs.id AND cr.categorya_id = $12)) AND ($13::bigint[] IS NULL OR EXISTS (SELECT 1 FROM category_resume cr WHERE cr.resume_id = rs.id AND cr.categorya_id = ANY($13)))`

func resumeListArgs(f resume_dto.ResumeFilter) []any {
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
		f.IsActive, f.MinPrice, f.MaxPrice, f.MinExperience,
		f.CategoryID, f.CategoryIDs,
	}
}

func (s *resumeService) Count(ctx context.Context, f resume_dto.ResumeFilter) (int64, error) {
	var total int64

	err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM resumes rs WHERE `+resumeListWhere, resumeListArgs(f)...).Scan(&total)

	return total, err
}

func (s *resumeService) loadCategories(ctx context.Context, ids []int64) (map[int64][]resume_dto.CategoryShort, error) {
	out := make(map[int64][]resume_dto.CategoryShort, len(ids))

	if len(ids) == 0 {
		return out, nil
	}

	rows, err := s.db.Query(ctx, `SELECT cr.resume_id, c.id, COALESCE(c.name->>'default', ''), COALESCE(c.is_active, FALSE) FROM category_resume cr JOIN categories c ON c.id = cr.categorya_id WHERE cr.resume_id = ANY($1) AND c.deleted_at IS NULL ORDER BY cr.resume_id, c.id`, ids)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			resumeID int64
			cat      resume_dto.CategoryShort
		)

		if err := rows.Scan(&resumeID, &cat.ID, &cat.Name, &cat.IsActive); err != nil {
			return nil, err
		}

		out[resumeID] = append(out[resumeID], cat)
	}

	return out, rows.Err()
}

func (s *resumeService) List(ctx context.Context, f resume_dto.ResumeFilter, cursor helper.CursorPayload, limit int) ([]*resume_dto.ResumeResponse, bool, error) {
	args := resumeListArgs(f)

	orderCol, fallbackValue, valueKind := resumeOrderConfig(f.SortBy, f.SortOrder)

	orderDir := normalizedOrder(f.SortOrder)

	cursorSQL, cursorArgs := resumeCursorClause(len(args)+1, orderCol, fallbackValue, valueKind, orderDir, cursor)

	if cursorSQL == "" {
		cursorSQL = "TRUE"
	}

	args = append(args, cursorArgs...)

	args = append(args, limit+1)

	rows, err := s.db.Query(ctx, fmt.Sprintf(`SELECT rs.id, rs.slug, rs.user_id, rs.region_id, COALESCE(r.name->>'default', ''), rs.district_id, COALESCE(d.name->>'default', ''), rs.mahalla_id, COALESCE(m.name->>'default', ''), rs.adress, rs.name, rs.photo, rs.title, rs.text, rs.contact, rs.price, rs.experience_year, rs.skills, rs.views_count, rs.is_active, rs.created_at, rs.updated_at, rs.deleted_at FROM resumes rs LEFT JOIN countries r ON r.id = rs.region_id AND r.deleted_at IS NULL LEFT JOIN countries d ON d.id = rs.district_id AND d.deleted_at IS NULL LEFT JOIN countries m ON m.id = rs.mahalla_id AND m.deleted_at IS NULL WHERE %s AND %s ORDER BY %s %s, rs.id %s LIMIT $%d`,
		resumeListWhere, cursorSQL, orderCol, orderDir, orderDir, len(args),
	), args...)

	if err != nil {
		return nil, false, err
	}

	defer rows.Close()

	items := make([]*resume_dto.ResumeResponse, 0, limit)

	for rows.Next() {
		var rs resume_dto.ResumeResponse

		if err := rows.Scan(
			&rs.ID, &rs.Slug, &rs.UserID,
			&rs.RegionID, &rs.RegionName,
			&rs.DistrictID, &rs.DistrictName,
			&rs.MahallaID, &rs.MahallaName,
			&rs.Adress, &rs.Name, &rs.Photo, &rs.Title, &rs.Text, &rs.Contact,
			&rs.Price, &rs.ExperienceYear, &rs.Skills,
			&rs.ViewsCount, &rs.IsActive,
			&rs.CreatedAt, &rs.UpdatedAt, &rs.DeletedAt,
		); err != nil {
			return nil, false, err
		}

		items = append(items, &rs)
	}

	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore := len(items) > limit

	if hasMore {
		items = items[:limit]
	}

	ids := make([]int64, 0, len(items))

	for _, rs := range items {
		ids = append(ids, rs.ID)
	}

	cats, err := s.loadCategories(ctx, ids)

	if err != nil {
		return nil, false, err
	}

	for _, rs := range items {
		rs.Categories = cats[rs.ID]

		if rs.Categories == nil {
			rs.Categories = []resume_dto.CategoryShort{}
		}
	}

	return items, hasMore, nil
}
