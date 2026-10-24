package attachments

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
)

// DiskStore 本地落盘：root/{workspace_id}/{attachment_id}
type DiskStore struct {
	Root string
}

func NewDiskStore(root string) *DiskStore {
	if strings.TrimSpace(root) == "" {
		root = "./data/attachments"
	}
	return &DiskStore{Root: root}
}

func (s *DiskStore) abs(key string) (string, error) {
	key = filepath.Clean(key)
	if key == "." || key == ".." || strings.HasPrefix(key, "..") || filepath.IsAbs(key) {
		return "", apperr.Internal("非法存储键")
	}
	full := filepath.Join(s.Root, key)
	// 防止跳出 root
	rel, err := filepath.Rel(s.Root, full)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", apperr.Internal("非法存储键")
	}
	return full, nil
}

// Save 将 r 写入 key；成功返回实际写入字节数。失败时尽量清理半成品。
func (s *DiskStore) Save(key string, r io.Reader, maxBytes int64) (written int64, err error) {
	full, err := s.abs(key)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return 0, apperr.Internal("创建附件目录失败")
	}
	f, err := os.OpenFile(full, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return 0, apperr.Internal("写入附件失败")
	}
	defer func() {
		_ = f.Close()
		if err != nil {
			_ = os.Remove(full)
		}
	}()

	limited := io.LimitReader(r, maxBytes+1)
	written, err = io.Copy(f, limited)
	if err != nil {
		return written, apperr.Internal("写入附件失败")
	}
	if written > maxBytes {
		err = apperr.BadRequest("文件过大，上限 10MB")
		return written, err
	}
	if written <= 0 {
		err = apperr.BadRequest("不支持的文件类型")
		return written, err
	}
	if err = f.Sync(); err != nil {
		err = apperr.Internal("写入附件失败")
		return written, err
	}
	return written, nil
}

func (s *DiskStore) Open(key string) (*os.File, error) {
	full, err := s.abs(key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(full)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, apperr.NotFound("附件文件不存在")
		}
		return nil, apperr.Internal("打开附件失败")
	}
	return f, nil
}

func (s *DiskStore) Delete(key string) error {
	full, err := s.abs(key)
	if err != nil {
		return err
	}
	err = os.Remove(full)
	if err != nil && !os.IsNotExist(err) {
		return apperr.Internal("删除附件文件失败")
	}
	return nil
}
