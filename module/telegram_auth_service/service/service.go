package telegram_auth_service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"main_service/helper"
	user_model "main_service/module/user_service/model"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const sessionTTL = 5 * time.Minute

type Service interface {
	StartLogin(ctx context.Context) (token, deepLink string, err error)
	GetStatus(ctx context.Context, token string) (status, jwt string, user *user_model.User, err error)
	HandleTelegramStart(ctx context.Context, loginToken string, tgUserID int64, tgUsername, firstName, lastName string) error
}

type svc struct {
	db          *pgxpool.Pool
	botUsername string
}

func New(db *pgxpool.Pool, botUsername string) Service {
	return &svc{db: db, botUsername: botUsername}
}

func randomToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *svc) StartLogin(ctx context.Context) (string, string, error) {
	token, err := randomToken()
	if err != nil {
		return "", "", err
	}
	expires := time.Now().Add(sessionTTL)

	_, err = s.db.Exec(ctx,
		`INSERT INTO tg_login_sessions (token, status, expires_at) VALUES ($1, 'pending', $2)`,
		token, expires)
	if err != nil {
		return "", "", err
	}

	deep := fmt.Sprintf("https://t.me/%s?start=%s", s.botUsername, token)
	return token, deep, nil
}

func (s *svc) GetStatus(ctx context.Context, token string) (string, string, *user_model.User, error) {
	var (
		status  string
		userID  *int64
		expires time.Time
	)
	err := s.db.QueryRow(ctx,
		`SELECT status, user_id, expires_at FROM tg_login_sessions WHERE token = $1`, token).
		Scan(&status, &userID, &expires)
	if errors.Is(err, pgx.ErrNoRows) {
		return "expired", "", nil, nil
	}
	if err != nil {
		return "", "", nil, err
	}

	if status == "pending" && time.Now().After(expires) {
		_, _ = s.db.Exec(ctx, `UPDATE tg_login_sessions SET status='expired' WHERE token=$1`, token)
		return "expired", "", nil, nil
	}
	if status != "done" || userID == nil {
		return status, "", nil, nil
	}

	var u user_model.User
	err = s.db.QueryRow(ctx,
		`SELECT id, full_name, photo, COALESCE(phone, ''), role, is_active, created_at, updated_at, deleted_at
		 FROM users WHERE id = $1`, *userID).
		Scan(&u.ID, &u.FullName, &u.Photo, &u.Phone, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt)
	if err != nil {
		return "", "", nil, err
	}

	jwt, err := helper.GenerateToken(int(u.ID), u.Role)
	if err != nil {
		return "", "", nil, err
	}
	return "done", jwt, &u, nil
}

func (s *svc) HandleTelegramStart(ctx context.Context, loginToken string, tgUserID int64, tgUsername, firstName, lastName string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status string
	var expires time.Time
	err = tx.QueryRow(ctx,
		`SELECT status, expires_at FROM tg_login_sessions WHERE token=$1 FOR UPDATE`, loginToken).
		Scan(&status, &expires)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("session not found")
	}
	if err != nil {
		return err
	}
	if status != "pending" || time.Now().After(expires) {
		return fmt.Errorf("session not active")
	}

	var userID int64
	err = tx.QueryRow(ctx,
		`SELECT id FROM users WHERE telegram_id = $1 AND deleted_at IS NULL`, tgUserID).
		Scan(&userID)

	if errors.Is(err, pgx.ErrNoRows) {
		fullName := strings.TrimSpace(firstName + " " + lastName)
		if fullName == "" {
			fullName = "Telegram user"
		}
		var tgUsernamePtr *string
		if tgUsername != "" {
			tgUsernamePtr = &tgUsername
		}
		err = tx.QueryRow(ctx,
			`INSERT INTO users (full_name, telegram_id, telegram_username, role, is_active)
			 VALUES ($1, $2, $3, 'user', true) RETURNING id`,
			fullName, tgUserID, tgUsernamePtr).Scan(&userID)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		var tgUsernamePtr *string
		if tgUsername != "" {
			tgUsernamePtr = &tgUsername
		}
		_, _ = tx.Exec(ctx,
			`UPDATE users SET telegram_username=$1, updated_at=NOW() WHERE id=$2`,
			tgUsernamePtr, userID)
	}

	_, err = tx.Exec(ctx,
		`UPDATE tg_login_sessions SET telegram_id=$1, user_id=$2, status='done' WHERE token=$3`,
		tgUserID, userID, loginToken)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
