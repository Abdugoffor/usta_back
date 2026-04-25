package middleware

import (
	"context"
	"main_service/helper"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type ctxKey string

const (
	CtxUserID ctxKey = "user_id"
	CtxRole   ctxKey = "role"
)

func CheckRole(next httprouter.Handle, roles ...string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		header := strings.TrimSpace(r.Header.Get("Authorization"))
		{
			if header == "" {
				helper.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
		}

		if !strings.HasPrefix(header, "Bearer ") {
			helper.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		{
			if token == "" {
				helper.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
		}

		claims, err := helper.ParseToken(token)
		{
			if err != nil {
				helper.WriteError(w, http.StatusUnauthorized, "invalid token")
				return
			}
		}

		if len(roles) > 0 {
			allowed := false

			for _, role := range roles {
				if strings.TrimSpace(role) == claims.Role {
					allowed = true
					break
				}
			}

			if !allowed {
				helper.WriteError(w, http.StatusForbidden, "forbidden")
				return
			}
		}

		ctx := context.WithValue(r.Context(), CtxUserID, claims.UserID)
		ctx = context.WithValue(ctx, CtxRole, claims.Role)

		next(w, r.WithContext(ctx), ps)
	}
}

func GetUserID(r *http.Request) int {
	v, _ := r.Context().Value(CtxUserID).(int)
	return v
}

func GetRole(r *http.Request) string {
	v, _ := r.Context().Value(CtxRole).(string)
	return v
}
