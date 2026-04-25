package upload_handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"main_service/helper"
	"main_service/middleware"
	"net/http"
	"os"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
)

const maxUploadSize = 3 << 20 // 3 MB

var allowedMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

func NewUploadHandler(router *httprouter.Router, group string) {
	router.POST(group+"/upload", middleware.CheckRole(upload))
}

func upload(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		helper.WriteError(w, http.StatusBadRequest, "fayl hajmi 3MB dan oshmasligi kerak")
		return
	}

	file, _, err := r.FormFile("photo")
	if err != nil {
		helper.WriteError(w, http.StatusBadRequest, "photo maydoni talab etiladi")
		return
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		helper.WriteInternalError(w, err)
		return
	}

	mimeType := http.DetectContentType(buf[:n])
	ext, ok := allowedMIME[mimeType]
	if !ok {
		helper.WriteError(w, http.StatusBadRequest, fmt.Sprintf("faqat rasm fayllari qabul qilinadi (jpeg, png, gif, webp), yuborilgan: %s", mimeType))
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		helper.WriteInternalError(w, err)
		return
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		helper.WriteInternalError(w, err)
		return
	}
	filename := hex.EncodeToString(b) + ext

	if err := os.MkdirAll("uploads", 0755); err != nil {
		helper.WriteInternalError(w, err)
		return
	}

	dst, err := os.Create(filepath.Join("uploads", filename))
	if err != nil {
		helper.WriteInternalError(w, err)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		helper.WriteInternalError(w, err)
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]string{"url": "/uploads/" + filename})
}
