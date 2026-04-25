package language_cmd

import (
	language_handler "main_service/module/language_service/handler"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

func Cmd(router *httprouter.Router, db *pgxpool.Pool) {
	language_handler.NewLanguageHandler(router, "/api/v1", db)
}
