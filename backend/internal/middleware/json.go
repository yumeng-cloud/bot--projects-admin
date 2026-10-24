package middleware

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
)

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("json encode error: %v", err)
	}
}

func DecodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func HandleError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	if ae, ok := err.(*apperr.AppError); ok {
		JSON(w, ae.StatusCode, map[string]string{"code": ae.Code, "message": ae.Message})
		return
	}
	log.Printf("internal error: %v", err)
	JSON(w, http.StatusInternalServerError, map[string]string{"code": "internal", "message": "服务器内部错误"})
}
