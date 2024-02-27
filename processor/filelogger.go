package processor

import (
	"encoding/json"
	"fmt"
	"github.com/qup42/loghead/types"
	"github.com/qup42/loghead/util"
	"os"
	"path/filepath"
)

type FileLoggerService struct {
	BaseDir string
}

func NewFileLoggerService(c types.FileLoggerConfig) (*FileLoggerService, error) {
	err := util.EnsureFolderExists(filepath.Join(c.Dir, TailnodeCollection))
	if err != nil {
		return nil, fmt.Errorf("init FileLogger: %w", err)
	}
	return &FileLoggerService{c.Dir}, nil
}

func (fl *FileLoggerService) Log(m LogtailMsg) error {
	p := filepath.Join(fl.BaseDir, TailnodeCollection, m.PrivateID)
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening %s: %w", p, err)
	}

	b, _ := json.Marshal(m.Msg)
	if _, err := f.Write(b); err != nil {
		return fmt.Errorf("writing to %s: %w", p, err)
	}
	if _, err := f.Write([]byte("\n")); err != nil {
		return fmt.Errorf("writing to %s: %w", p, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("closing %s: %w", p, err)
	}
	return nil
}
