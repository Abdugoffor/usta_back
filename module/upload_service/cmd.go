package upload_cmd

import (
	upload_handler "main_service/module/upload_service/handler"

	"github.com/julienschmidt/httprouter"
)

func Cmd(router *httprouter.Router) {
	upload_handler.NewUploadHandler(router, "/api/v1")
}
