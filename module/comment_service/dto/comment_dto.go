package comment_dto

import "time"

// ─── Request DTOs ────────────────────────────────────────────────────────────

type CreateCommentRequest struct {
	ParentID    *int64 `json:"parent_id"`
	VakansiyaID *int64 `json:"vakansiya_id"`
	ResumeID    *int64 `json:"resume_id"`
	Type        string `json:"type" validate:"required,oneof=comment review"`
	Text        string `json:"text" validate:"required,min=1,max=5000"`
}

type UpdateCommentRequest struct {
	Type *string `json:"type" validate:"omitempty,oneof=comment review"`
	Text *string `json:"text" validate:"omitempty,min=1,max=5000"`
}

// ─── Filter ──────────────────────────────────────────────────────────────────

type CommentFilter struct {
	VakansiyaID *int64
	ResumeID    *int64
	ParentID    *int64
	UserID      *int64
	Type        string
}

// ─── Response DTOs ───────────────────────────────────────────────────────────

type CommentResponse struct {
	ID          int64             `json:"id"`
	ParentID    *int64            `json:"parent_id"`
	UserID      *int64            `json:"user_id"`
	VakansiyaID *int64            `json:"vakansiya_id"`
	ResumeID    *int64            `json:"resume_id"`
	Type        *string           `json:"type"`
	Text        *string           `json:"text"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	DeletedAt   *time.Time        `json:"deleted_at,omitempty"`
	Children    []*CommentResponse `json:"children,omitempty"`
}
