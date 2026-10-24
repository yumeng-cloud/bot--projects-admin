package attachments

import (
	"path/filepath"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
)

// 白名单扩展名 → 规范 MIME（上传时以扩展名权威校验；客户端 Content-Type 仅作提示）。
var allowedExtMIME = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
	".pdf":  "application/pdf",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".txt":  "text/plain",
	".md":   "text/markdown",
	".csv":  "text/csv",
}

// IsImageMIME reports whether content should be served inline for preview.
func IsImageMIME(ct string) bool {
	ct = strings.ToLower(strings.TrimSpace(ct))
	switch ct {
	case "image/png", "image/jpeg", "image/gif", "image/webp":
		return true
	default:
		return false
	}
}

func normalizeFilename(name string) string {
	name = strings.TrimSpace(name)
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, "\x00", "")
	if name == "" || name == "." || name == ".." {
		return "file"
	}
	return name
}

func resolveContentType(filename, clientCT string) (ext, mime string, err error) {
	ext = strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return "", "", apperr.BadRequest("不支持的文件类型")
	}
	canonical, ok := allowedExtMIME[ext]
	if !ok {
		return "", "", apperr.BadRequest("不支持的文件类型")
	}
	_ = clientCT // 扩展名权威；不信任客户端 MIME
	return ext, canonical, nil
}

// checkDeclaredSize：size>0 时做预检；size<=0 表示未知，交给落盘 LimitReader。
func checkDeclaredSize(size int64) error {
	if size > MaxBytes {
		return apperr.BadRequest("文件过大，上限 10MB")
	}
	return nil
}
