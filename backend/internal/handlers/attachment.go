package handlers

import (
	"database/sql"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/attachments"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
)

// AttachmentHandler 仅挂载上传 + 元数据 + 内容流。
// List / Delete HTTP 由 Issue 域挂载，调用 attachments.Service。
type AttachmentHandler struct {
	DB  *sql.DB
	Svc *attachments.Service
}

func NewAttachmentHandler(db *sql.DB, storeRoot string) *AttachmentHandler {
	return &AttachmentHandler{DB: db, Svc: attachments.New(db, storeRoot)}
}

// UploadIssue POST …/issues/{issueID}/attachments  multipart field "file"
func (h *AttachmentHandler) UploadIssue(w http.ResponseWriter, r *http.Request) {
	ws, mid, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	issueID := r.PathValue("issueID")
	att, err := h.readAndUpload(r, ws, mid, issueID, "")
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{"attachment": att})
}

// UploadComment POST …/issues/{issueID}/comments/{commentID}/attachments
func (h *AttachmentHandler) UploadComment(w http.ResponseWriter, r *http.Request) {
	ws, mid, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	issueID := r.PathValue("issueID")
	commentID := r.PathValue("commentID")
	att, err := h.readAndUpload(r, ws, mid, issueID, commentID)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{"attachment": att})
}

func (h *AttachmentHandler) readAndUpload(r *http.Request, ws, mid, issueID, commentID string) (*attachments.Attachment, error) {
	// 限制 multipart 内存缓冲；总大小由 service MaxBytes 再挡
	if err := r.ParseMultipartForm(attachments.MaxBytes + (1 << 20)); err != nil {
		return nil, apperr.BadRequest("无效的上传表单")
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		return nil, apperr.BadRequest("缺少文件字段 file")
	}
	defer file.Close()

	filename := hdr.Filename
	if v := strings.TrimSpace(r.FormValue("filename")); v != "" {
		filename = v
	}
	ct := hdr.Header.Get("Content-Type")
	size := hdr.Size // 可能为 -1（未知）；service 会按流式 LimitReader 限制

	if commentID == "" {
		return h.Svc.UploadIssue(r.Context(), ws, issueID, mid, filename, ct, file, size)
	}
	return h.Svc.UploadComment(r.Context(), ws, issueID, commentID, mid, filename, ct, file, size)
}

// GetMeta GET …/attachments/{attachmentID}
func (h *AttachmentHandler) GetMeta(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	att, err := h.Svc.Get(r.Context(), ws, r.PathValue("attachmentID"))
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"attachment": att})
}

// GetContent GET …/attachments/{attachmentID}/content
func (h *AttachmentHandler) GetContent(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	att, rc, err := h.Svc.OpenContent(r.Context(), ws, r.PathValue("attachmentID"))
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", att.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(att.SizeBytes, 10))
	disp := "attachment"
	if attachments.IsImageMIME(att.ContentType) {
		disp = "inline"
	}
	// RFC 5987 filename*
	safeName := filepath.Base(att.Filename)
	w.Header().Set("Content-Disposition", disp+`; filename="`+escapeDisposition(safeName)+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, rc)
}

func escapeDisposition(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
