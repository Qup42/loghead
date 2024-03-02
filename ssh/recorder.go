package ssh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/qup42/loghead/types"
	"github.com/qup42/loghead/util"
	"io"
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

func (rec *RecordingService) Record(s io.ReadCloser) error {
	// the metadata is the first line
	b, err := readSingleUntil(s, []byte("\n"))
	if err != nil {
		return errors.Errorf("reading stream until metadata: %w", err)
	}
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
	// write the metadata out
	_, err = f.Write(b)
	if err != nil {
		return errors.Errorf("writing ssh session recording: %w", err)
	}
	// stream the rest of the logs directly to the file
	_, err = io.Copy(f, s)
	if err != nil {
		return errors.Errorf("writing ssh session recording: %w", err)
	}
	err = f.Close()
	if err != nil {
		return errors.Errorf("closing ssh session recording file: %w", err)
	}
	return nil
}

func readSingleUntil(r io.Reader, sep []byte) ([]byte, error) {
	// based on the implementation of io.ReadAll
	b := make([]byte, 0, 512)
	for {
		_, err := r.Read(b[len(b) : len(b)+1])
		if err != nil {
			return b, err
		}
		b = b[:len(b)+1]
		if bytes.Contains(b[len(b)-1:len(b)], sep) {
			return b, nil
		}

		if len(b) == cap(b) {
			// Add more capacity (let append pick how much).
			b = append(b, 0)[:len(b)]
		}
	}
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
