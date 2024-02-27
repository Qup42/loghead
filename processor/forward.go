package processor

import (
	"bytes"
	"fmt"
	"net/http"
)

type ForwardingService struct {
	Addr string
}

func NewForwardingService(addr string) *ForwardingService {
	return &ForwardingService{addr}
}

func (fwd *ForwardingService) Forward(m []byte) error {
	req, err := http.NewRequest("POST", fwd.Addr, bytes.NewReader(m))
	if err != nil {
		return fmt.Errorf("Error creating forward request")
	}
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("Error forwarding log")
	}
	return nil
}
