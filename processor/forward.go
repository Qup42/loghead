package processor

import (
	"bytes"
	"github.com/rs/zerolog/log"
	"net/http"
)

type Forwarder struct {
	Addr string
}

func NewForwarder(addr string) Forwarder {
	return Forwarder{addr}
}

func (fwd *Forwarder) Process(m []byte) {
	req, err := http.NewRequest("POST", fwd.Addr, bytes.NewReader(m))
	if err != nil {
		log.Error().Err(err).Msg("Error creating forward request")
	}
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Error forwarding log")
	}
}
