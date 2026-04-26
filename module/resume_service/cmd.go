package resume_cmd

import (
	resume_handler "main_service/module/resume_service/handler"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

func Cmd(router *httprouter.Router, db *pgxpool.Pool) {
	resume_handler.NewResumeHandler(router, "/api/v1", db)
	resume_handler.NewClientResumeHandler(router, "/api/v1", db)
}
