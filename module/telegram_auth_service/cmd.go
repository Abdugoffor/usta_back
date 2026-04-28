package telegram_auth_cmd

import (
	"context"
	"main_service/helper"
	telegram_auth_handler "main_service/module/telegram_auth_service/handler"
	telegram_auth_service "main_service/module/telegram_auth_service/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

func Cmd(router *httprouter.Router, db *pgxpool.Pool) {
	botUsername := helper.ENV("TELEGRAM_BOT_USERNAME")
	botToken := helper.ENV("TELEGRAM_BOT_TOKEN")

	svc := telegram_auth_service.New(db, botUsername)

	telegram_auth_handler.New(router, "/api/v1", svc)

	telegram_auth_service.StartPoller(context.Background(), botToken, svc)
}
