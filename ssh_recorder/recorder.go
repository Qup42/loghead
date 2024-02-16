package ssh_recorder

import (
	"fmt"
	"github.com/qup42/loghead/types"
	"github.com/qup42/loghead/util"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
	"time"
)

type SSHRecorder struct {
	Dir string
}

func NewSSHRecorder(c types.SSHRecorderConfig) SSHRecorder {
	err := util.EnsureFolderExists(c.Dir)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create ssh recording directory")
	}
	return SSHRecorder{c.Dir}
}

func (rec *SSHRecorder) Handle(w http.ResponseWriter, r *http.Request) {
	log.Trace().Msg("Starting SSH Session recording")
	fn := fmt.Sprintf("ssh-session-%v-*.cast", time.Now().UnixNano())
	f, err := os.CreateTemp(rec.Dir, fn)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open file for ssh session recording")
	}
	body, _ := io.ReadAll(r.Body)
	_, err = f.Write(body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to write ssh session recording to file")
	}
	log.Debug().Msgf("Recorded %s", f.Name())
	err = f.Close()
	if err != nil {
		log.Error().Err(err).Msg("Could not close ssh session recording file")
	}
	log.Trace().Msg("SSH Session recording finished")
}
