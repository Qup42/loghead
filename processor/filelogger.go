package processor

import (
	"encoding/json"
	"fmt"
	"github.com/qup42/loghead/types"
	"github.com/qup42/loghead/util"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
)

type FileLogger struct {
	BaseDir string
}

func NewFileLogger(c types.FileLoggerConfig) (*FileLogger, error) {
	err := util.EnsureFolderExists(filepath.Join(c.Dir, TailnodeCollection))
	if err != nil {
		return nil, fmt.Errorf("init FileLogger: %w", err)
	}
	return &FileLogger{c.Dir}, nil
}

func (fl *FileLogger) Process(m LogtailMsg) {
	p := filepath.Join(fl.BaseDir, TailnodeCollection, m.PrivateID)
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error().Err(err).Str("path", p).Msg("Error opening log file")
	}

	b, _ := json.Marshal(m.Msg)
	if _, err := f.Write(b); err != nil {
		log.Error().Err(err).Str("path", p).Msg("Failed to write to file")
	}
	if _, err := f.Write([]byte("\n")); err != nil {
		log.Error().Err(err).Str("path", p).Msg("Failed to write to file")
	}

	if err := f.Close(); err != nil {
		log.Error().Err(err).Str("path", p).Msg("Failed to close file")
	}
}
