package types

import (
	"context"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"
	"net"
	"tailscale.com/tsnet"
	"time"
)

type Listener struct {
	Listener net.Listener
	TSServer *tsnet.Server
}

func waitTSReady(ctx context.Context, s *tsnet.Server) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := s.Up(ctx)
	if err != nil {
		return err
	}

	return nil
}

func makeTS(ctx context.Context, c ListenerConfig) (*tsnet.Server, error) {
	s := tsnet.Server{
		Logf: func(format string, v ...any) {
			log.Trace().Str("ts", c.TS.HostName).Msgf(format, v...)
		},
		AuthKey:    c.TS.AuthKey,
		ControlURL: c.TS.ControllURL,
		Hostname:   c.TS.HostName,
		Dir:        c.TS.Dir,
	}

	err := waitTSReady(ctx, &s)
	if err != nil {
		// TODO: handle error during teardown
		_ = s.Close()
		return nil, errors.Errorf("create tsnet.Server: %w", err)
	}

	return &s, nil
}

func MakeListener(ctx context.Context, c ListenerConfig, componentName string) (*Listener, error) {
	switch c.Type {
	case "plain":
		ln, err := net.Listen("tcp", net.JoinHostPort(c.Addr, c.Port))
		if err != nil {
			return nil, errors.Errorf("binding port: %w", err)
		}
		log.Info().Msgf("%s Listening on %s:%s", componentName, c.Addr, c.Port)
		return &Listener{ln, nil}, nil
	case "tsnet":
		s, err := makeTS(ctx, c)
		if err != nil {
			return nil, errors.Errorf("starting ts server: %w", err)
		}

		ln, err := s.Listen("tcp", fmt.Sprintf(":%s", c.Port))
		if err != nil {
			return nil, errors.Errorf("binding port on ts listener: %w", err)
		}
		log.Info().Msgf("%s Listening over tailscale on :%s", componentName, c.Port)
		return &Listener{ln, s}, nil
	default:
		return nil, errors.Errorf("unknown listener type %s", c.Type)
	}
}

func (ln *Listener) Close() error {
	errListener := ln.Listener.Close()
	var errTSServer error
	if ln.TSServer != nil {
		errTSServer = ln.TSServer.Close()
	}
	return errors.Join(errListener, errTSServer)
}
