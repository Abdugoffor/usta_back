package comment_service

import (
	"context"
	"fmt"
	comment_dto "main_service/module/comment_service/dto"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ─── Interface ───────────────────────────────────────────────────────────────

type CommentService interface {
	Create(ctx context.Context, userID int64, req comment_dto.CreateCommentRequest) (*comment_dto.CommentResponse, error)
	GetByID(ctx context.Context, id int64) (*comment_dto.CommentResponse, error)
	Update(ctx context.Context, id, userID int64, req comment_dto.UpdateCommentRequest) (*comment_dto.CommentResponse, error)
	Delete(ctx context.Context, id, userID int64) error
	List(ctx context.Context, f comment_dto.CommentFilter, page, limit int, sortCol, sortOrder string) ([]*comment_dto.CommentResponse, error)
	Count(ctx context.Context, f comment_dto.CommentFilter) (int64, error)
}

// ─── Implementation ──────────────────────────────────────────────────────────

type commentService struct {
	db *pgxpool.Pool
}

func NewCommentService(db *pgxpool.Pool) CommentService {
	return &commentService{db: db}
}

// ─── Create ──────────────────────────────────────────────────────────────────

func (s *commentService) Create(ctx context.Context, userID int64, req comment_dto.CreateCommentRequest) (*comment_dto.CommentResponse, error) {
	var id int64
	err := s.db.QueryRow(ctx, `
		INSERT INTO comments (parent_id, user_id, vakansiya_id, resume_id, type, text)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, req.ParentID, userID, req.VakansiyaID, req.ResumeID, req.Type, req.Text).Scan(&id)
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, id)
}

// ─── GetByID ─────────────────────────────────────────────────────────────────

func (s *commentService) GetByID(ctx context.Context, id int64) (*comment_dto.CommentResponse, error) {
	var c comment_dto.CommentResponse
	err := s.db.QueryRow(ctx, `
		SELECT id, parent_id, user_id, vakansiya_id, resume_id, type, text, created_at, updated_at, deleted_at
		FROM comments
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&c.ID, &c.ParentID, &c.UserID, &c.VakansiyaID, &c.ResumeID,
		&c.Type, &c.Text, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
	)
	return &c, err
}

// ─── Update ──────────────────────────────────────────────────────────────────

func (s *commentService) Update(ctx context.Context, id, userID int64, req comment_dto.UpdateCommentRequest) (*comment_dto.CommentResponse, error) {
	var retID int64

	err := s.db.QueryRow(ctx, `
		UPDATE comments
		SET type       = COALESCE($1::varchar, type),
		    text       = COALESCE($2::text, text),
		    updated_at = NOW()
		WHERE id = $3 AND user_id = $4 AND deleted_at IS NULL
		RETURNING id
	`, req.Type, req.Text, id, userID).Scan(&retID)

	if err != nil {
		return nil, fmt.Errorf("comment not found or access denied")
	}

	return s.GetByID(ctx, retID)
}

// ─── Delete ──────────────────────────────────────────────────────────────────

func (s *commentService) Delete(ctx context.Context, id, userID int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE comments SET deleted_at = $1
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`, time.Now(), id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("comment not found or access denied")
	}
	return nil
}

var validSortCols = map[string]string{
	"id":         "c.id",
	"type":       "c.type",
	"created_at": "c.created_at",
	"updated_at": "c.updated_at",
}

const commentListWhere = `c.deleted_at IS NULL AND c.parent_id IS NULL AND ($1::bigint IS NULL OR c.vakansiya_id = $1) AND ($2::bigint IS NULL OR c.resume_id = $2) AND ($3::bigint IS NULL OR c.user_id = $3) AND ($4::varchar IS NULL OR c.type = $4)`

func commentListArgs(f comment_dto.CommentFilter) []any {
	var commentType *string

	if f.Type != "" {
		commentType = &f.Type
	}

	return []any{f.VakansiyaID, f.ResumeID, f.UserID, commentType}
}

func (s *commentService) Count(ctx context.Context, f comment_dto.CommentFilter) (int64, error) {
	var total int64

	err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM comments c WHERE `+commentListWhere, commentListArgs(f)...).Scan(&total)

	return total, err
}

// ─── List (threaded) ─────────────────────────────────────────────────────────

func (s *commentService) List(ctx context.Context, f comment_dto.CommentFilter, page, limit int, sortCol, sortOrder string) ([]*comment_dto.CommentResponse, error) {
	col := validSortCols[sortCol]

	if col == "" {
		col = "c.id"
	}

	if sortOrder != "DESC" {
		sortOrder = "ASC"
	}

	args := commentListArgs(f)

	args = append(args, limit, (page-1)*limit)

	rootRows, err := s.db.Query(ctx, fmt.Sprintf(`SELECT c.id FROM comments c WHERE %s ORDER BY %s %s LIMIT $5 OFFSET $6`, commentListWhere, col, sortOrder), args...)

	if err != nil {
		return nil, err
	}

	defer rootRows.Close()

	var rootIDs []int64
	for rootRows.Next() {
		var id int64
		if err := rootRows.Scan(&id); err != nil {
			return nil, err
		}
		rootIDs = append(rootIDs, id)
	}
	rootRows.Close()

	if len(rootIDs) == 0 {
		return []*comment_dto.CommentResponse{}, nil
	}

	allRows, err := s.db.Query(ctx, `
		WITH RECURSIVE tree AS (
			SELECT id, parent_id, user_id, vakansiya_id, resume_id, type, text, created_at, updated_at, deleted_at
			FROM comments
			WHERE id = ANY($1) AND deleted_at IS NULL
			UNION ALL
			SELECT c.id, c.parent_id, c.user_id, c.vakansiya_id, c.resume_id, c.type, c.text, c.created_at, c.updated_at, c.deleted_at
			FROM comments c
			JOIN tree t ON c.parent_id = t.id
			WHERE c.deleted_at IS NULL
		)
		SELECT id, parent_id, user_id, vakansiya_id, resume_id, type, text, created_at, updated_at, deleted_at
		FROM tree
		ORDER BY parent_id NULLS FIRST, id
	`, rootIDs)
	if err != nil {
		return nil, err
	}
	defer allRows.Close()

	nodeMap := map[int64]*comment_dto.CommentResponse{}
	for allRows.Next() {
		var c comment_dto.CommentResponse
		if err := allRows.Scan(
			&c.ID, &c.ParentID, &c.UserID, &c.VakansiyaID, &c.ResumeID,
			&c.Type, &c.Text, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
		); err != nil {
			return nil, err
		}
		cc := c
		nodeMap[c.ID] = &cc
	}
	if err := allRows.Err(); err != nil {
		return nil, err
	}

	rootSet := make(map[int64]bool, len(rootIDs))
	for _, id := range rootIDs {
		rootSet[id] = true
	}

	for _, node := range nodeMap {
		if !rootSet[node.ID] && node.ParentID != nil {
			if parent, ok := nodeMap[*node.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	roots := make([]*comment_dto.CommentResponse, 0, len(rootIDs))
	for _, id := range rootIDs {
		if node, ok := nodeMap[id]; ok {
			roots = append(roots, node)
		}
	}
	return roots, nil
}
