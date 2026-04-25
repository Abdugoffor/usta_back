package translations_cmd

import (
	translation_handler "main_service/module/translations_service/handler"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

func Cmd(router *httprouter.Router, db *pgxpool.Pool) {
	translation_handler.NewTranslationHandler(router, "/api/v1", db)
}
