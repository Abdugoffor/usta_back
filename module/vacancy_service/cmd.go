package vacancy_cmd

import (
	vacancy_handler "main_service/module/vacancy_service/handler"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

func Cmd(router *httprouter.Router, db *pgxpool.Pool) {
	vacancy_handler.NewVacancyHandler(router, "/api/v1", db)
}
