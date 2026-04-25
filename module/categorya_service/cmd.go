package categorya_cmd

import (
	categorya_handler "main_service/module/categorya_service/handler"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

func Cmd(router *httprouter.Router, db *pgxpool.Pool) {
	categorya_handler.NewCategoryHandler(router, "/api/v1", db)
}
