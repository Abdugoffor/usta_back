package helper

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func FetchActiveLanguages(ctx context.Context, db *pgxpool.Pool) ([]string, error) {
	rows, err := db.Query(ctx, `SELECT name FROM languages WHERE is_active = TRUE AND deleted_at IS NULL ORDER BY id`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	codes := make([]string, 0)

	for rows.Next() {
		var c string

		if err := rows.Scan(&c); err != nil {
			return nil, err
		}

		codes = append(codes, strings.ToLower(strings.TrimSpace(c)))
	}

	return codes, rows.Err()
}

func SanitizeMultilang(name map[string]string) {
	for k, v := range name {
		name[k] = strings.TrimSpace(v)
	}
}

func NormalizeMultilang(name map[string]string) map[string]string {
	if name == nil {
		return map[string]string{}
	}

	out := make(map[string]string, len(name))

	for k, v := range name {
		out[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
	}

	return out
}

func ValidateMultilang(name map[string]string, activeLangs []string, requireDefault bool) map[string]string {
	errs := map[string]string{}

	if requireDefault && strings.TrimSpace(name["default"]) == "" {
		errs["name.default"] = "default qiymati majburiy"
	}

	for _, lang := range activeLangs {
		if strings.TrimSpace(name[lang]) == "" {
			errs["name."+lang] = lang + " tilida kiritish majburiy"
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

func FilterMultilangByActive(name map[string]string, activeLangs []string) map[string]string {
	allowed := make(map[string]bool, len(activeLangs)+1)

	allowed["default"] = true

	for _, l := range activeLangs {
		allowed[l] = true
	}

	out := make(map[string]string, len(name))

	for k, v := range name {
		if allowed[k] {
			out[k] = v
		}
	}

	return out
}
