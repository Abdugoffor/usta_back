package country_cmd

import (
	country_handler "main_service/module/country_service/handler"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

func Cmd(router *httprouter.Router, db *pgxpool.Pool) {
	country_handler.NewCountryHandler(router, "/api/v1", db)
}
