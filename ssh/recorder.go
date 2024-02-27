package ssh

import (
	"fmt"
	"github.com/qup42/loghead/types"
	"github.com/qup42/loghead/util"
	"os"
	"time"
)

type RecordingService struct {
	Dir string
}

func NewRecordingService(c types.SSHRecorderConfig) (*RecordingService, error) {
	err := util.EnsureFolderExists(c.Dir)
	if err != nil {
		return nil, fmt.Errorf("init SSHRecorder: %w", err)
	}
	return &RecordingService{c.Dir}, nil
}

func (rec *RecordingService) Record(b []byte) error {
	fn := fmt.Sprintf("ssh-session-%v-*.cast", time.Now().UnixNano())
	f, err := os.CreateTemp(rec.Dir, fn)
	if err != nil {
		return fmt.Errorf("opening ssh session recording file: %w", err)
	}
	_, err = f.Write(b)
	if err != nil {
		return fmt.Errorf("writing ssh session recording: %w", err)
	}
	err = f.Close()
	if err != nil {
		return fmt.Errorf("closing ssh session recording file: %w", err)
	}
	return nil
}
