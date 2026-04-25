package user_cmd

import (
	user_handler "main_service/module/user_service/handler"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

func Cmd(router *httprouter.Router, db *pgxpool.Pool) {
	user_handler.NewUserHandler(router, "/api/v1", db)
}
