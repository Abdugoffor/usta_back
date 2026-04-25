package comment_cmd

import (
	comment_handler "main_service/module/comment_service/handler"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

func Cmd(router *httprouter.Router, db *pgxpool.Pool) {
	comment_handler.NewCommentHandler(router, "/api/v1", db)
}
