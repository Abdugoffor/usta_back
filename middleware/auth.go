package middleware

// import (
// 	"context"
// 	"fmt"
// 	"main_service/helper"
// 	"net/http"
// 	"strings"

// 	"github.com/julienschmidt/httprouter"
// )

// type ctxKey string

// const (
// 	CtxUserID ctxKey = "user_id"
// 	CtxRole   ctxKey = "role"
// )

// func Auth(next httprouter.Handle) httprouter.Handle {
// 	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 		header := r.Header.Get("Authorization")
// 		{
// 			if !strings.HasPrefix(header, "Bearer ") {
// 				helper.WriteError(w, http.StatusUnauthorized, "unauthorized")
// 				return
// 			}
// 		}

// 		fmt.Println("auth middleware: ", header)

// 		claims, err := helper.ParseToken(strings.TrimPrefix(header, "Bearer "))
// 		{
// 			if err != nil {
// 				helper.WriteError(w, http.StatusUnauthorized, "invalid token")
// 				return
// 			}
// 		}

// 		ctx := context.WithValue(r.Context(), CtxUserID, claims.UserID)

// 		ctx = context.WithValue(ctx, CtxRole, claims.Role)

// 		next(w, r.WithContext(ctx), ps)
// 	}
// }

// func GetUserID(r *http.Request) int {
// 	v, _ := r.Context().Value(CtxUserID).(int)
// 	return v
// }

// func GetRole(r *http.Request) string {
// 	v, _ := r.Context().Value(CtxRole).(string)
// 	return v
// }
