package ssh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/qup42/loghead/types"
	"github.com/qup42/loghead/util"
	"github.com/rs/zerolog/log"
	"os"
	"path"
	"time"
)

type RecordingService struct {
	Dir string
}

// See https://docs.asciinema.org/manual/asciicast/v2/
// and https://github.com/tailscale/tailscale/blob/main/ssh/tailssh/tailssh.go#L1718-L1740
type CastMetadata struct {
	Version       int      `json:"version"`
	Width         int      `json:"width"`
	Height        int      `json:"height"`
	Timestamp     UnixTime `json:"timestamp"`
	Command       string   `json:"command,omitempty"`
	SrcNode       string   `json:"srcNode"`
	SrcNodeID     string   `json:"srcNodeID"`
	SrcNodeTags   string   `json:"srcNodeTags,omitempty"`
	SrcNodeUser   string   `json:"srcNodeUser,omitempty"`
	SrcNodeUserID int64    `json:"srcNodeUserID,omitempty"`
	SSHUser       string   `json:"sshUser"`
	LocalUser     string   `json:"localUser"`
	ConnectionID  string   `json:"connectionID"`
}

// Taken from https://ikso.us/posts/unmarshal-timestamp-as-time/
type UnixTime struct {
	time.Time
}

func (u *UnixTime) UnmarshalJSON(b []byte) error {
	var timestamp int64
	err := json.Unmarshal(b, &timestamp)
	if err != nil {
		return err
	}
	u.Time = time.Unix(timestamp, 0)
	return nil
}

func (u UnixTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", (u.Time.Unix()))), nil
}

func NewRecordingService(c types.SSHRecorderConfig) (*RecordingService, error) {
	err := util.EnsureFolderExists(c.Dir)
	if err != nil {
		return nil, errors.Errorf("init SSHRecorder: %w", err)
	}
	return &RecordingService{c.Dir}, nil
}

func (rec *RecordingService) Record(b []byte) error {
	if len(b) == 0 {
		log.Info().Msg("Discarding empty recording")
		return nil
	}
	// TODO: processing the metadata should be done during the recording
	meta, err := readCastMetadata(b)
	if err != nil {
		return errors.Errorf("reading recording metadata: %w", err)
	}
	// use the accesing node's stable node id,
	// tailscale instead uses the target's stable node id
	// https://tailscale.com/kb/1246/tailscale-ssh-session-recording?q=.cast#session-recordings
	recDir := path.Join(rec.Dir, meta.SrcNodeID)
	err = util.EnsureFolderExists(recDir)
	if err != nil {
		return errors.Errorf("creating recording directory: %w", err)
	}
	recP := path.Join(recDir, meta.Timestamp.Format(time.RFC3339)+".cast")
	// os.O_CREATE|os.O_EXCL ensures that no recordings are overwriten
	f, err := os.OpenFile(recP, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return errors.Errorf("opening ssh session recording file: %w", err)
	}
	_, err = f.Write(b)
	if err != nil {
		return errors.Errorf("writing ssh session recording: %w", err)
	}
	err = f.Close()
	if err != nil {
		return errors.Errorf("closing ssh session recording file: %w", err)
	}
	return nil
}

func readCastMetadata(b []byte) (*CastMetadata, error) {
	i := bytes.Index(b, []byte("\n"))
	if i == -1 {
		return nil, errors.Errorf("no metadata found")
	}
	metadataBytes := b[:i]
	var metadata CastMetadata
	err := json.Unmarshal(metadataBytes, &metadata)
	if err != nil {
		return nil, errors.Errorf("unmarshaling cast metadata: %w", err)
	}
	return &metadata, nil
}
