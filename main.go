package main

import (
	"context"
	"fmt"
	"github.com/efekarakus/termcolor"
	"github.com/gorilla/mux"
	"github.com/qup42/loghead/processor"
	"github.com/qup42/loghead/ssh"
	"github.com/qup42/loghead/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
	"os"
	"os/signal"
	"tailscale.com/tsnet"
	"time"
)

func SetupLogging() {
	var colors bool
	switch l := termcolor.SupportLevel(os.Stderr); l {
	case termcolor.Level16M:
		colors = true
	case termcolor.Level256:
		colors = true
	case termcolor.LevelBasic:
		colors = true
	case termcolor.LevelNone:
		colors = false
	default:
		// no color, return text as is.
		colors = false
	}

	// Adhere to no-color.org manifesto of allowing users to
	// turn off color in cli/services
	if _, noColorIsSet := os.LookupEnv("NO_COLOR"); noColorIsSet {
		colors = false
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
		NoColor:    !colors,
	}).With().Caller().Logger()
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

func serve(ctx context.Context, r *mux.Router, ln net.Listener) error {
	s := http.Server{
		Handler:      r,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- s.Serve(ln)
	}()

	var err error
	select {
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = s.Shutdown(ctx)
	case err = <-serveErr:
	}
	return err
}

func makeTS(ctx context.Context, c types.ListenerConfig) (*tsnet.Server, error) {
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
		return nil, fmt.Errorf("create tsnet.Server: %w", err)
	}

	return &s, nil
}

func main() {
	SetupLogging()

	c, err := types.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	zerolog.SetGlobalLevel(c.Log.Level)

	log.Debug().Msgf("Config: %+v", c)

	var fls *processor.FileLoggerService
	var fwd *processor.ForwardingService
	var hs *processor.HostInfoService
	var ms *processor.MetricsService
	var rs *ssh.RecordingService
	if c.Processors.FileLogger {
		fls, err = processor.NewFileLoggerService(c.FileLogger)
		if err != nil {
			log.Fatal().Err(err)
		}
	}
	if c.Processors.Metrics {
		ms = processor.NewMetricsService()
	}
	if c.Processors.Hostinfo {
		hs = &processor.HostInfoService{}
	}
	if c.Processors.Forward != "" {
		log.Info().Msgf("Enableing forwarder to %s", c.Processors.Forward)
		fwd = processor.NewForwardingService(c.Processors.Forward)
	}
	rs, err = ssh.NewRecordingService(c.SSHRecorder)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not create SSH Recorder")
	}

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	// logtail
	ltr := mux.NewRouter()
	addClientLogsRoutes(ltr, c, fwd, fls, hs, ms)

	var ln net.Listener
	switch c.Listener.Type {
	case "plain":
		ln, err = net.Listen("tcp", net.JoinHostPort(c.Listener.Addr, c.Listener.Port))
		defer ln.Close()
		if err != nil {
			log.Fatal().Err(err).Msg("starting listener")
		}
		log.Info().Msgf("loghead Listening on %s:%s", c.Listener.Addr, c.Listener.Port)
	case "tsnet":
		s, err := makeTS(ctx, c.Listener)
		defer s.Close()
		if err != nil {
			log.Fatal().Err(err).Msg("starting ts")
		}

		ln, err = s.Listen("tcp", fmt.Sprintf(":%s", c.Listener.Port))
		defer ln.Close()
		if err != nil {
			log.Fatal().Err(err).Msg("starting ts listener")
		}
		log.Info().Msgf("loghead Listening over tailscale on :%s", c.Listener.Port)
	default:
		log.Fatal().Msgf("unknown listener type %s", c.Listener.Type)
	}
	g.Go(func() error {
		return serve(ctx, ltr, ln)
	})

	// SSH session recording
	sr := mux.NewRouter()
	addSSHRecordingRoutes(sr, rs)

	//	var ln net.Listener
	switch c.SSHRecorder.Listener.Type {
	case "plain":
		ln, err = net.Listen("tcp", net.JoinHostPort(c.SSHRecorder.Listener.Addr, c.SSHRecorder.Listener.Port))
		defer ln.Close()
		if err != nil {
			log.Fatal().Err(err).Msg("starting listener")
		}
		log.Info().Msgf("SSHRecorder Listening on %s:%s", c.SSHRecorder.Listener.Addr, c.SSHRecorder.Listener.Port)
	case "tsnet":
		s, err := makeTS(ctx, c.SSHRecorder.Listener)
		defer s.Close()
		if err != nil {
			log.Fatal().Err(err).Msg("starting ts")
		}

		ln, err = s.Listen("tcp", fmt.Sprintf(":%s", c.SSHRecorder.Listener.Port))
		defer ln.Close()
		if err != nil {
			log.Fatal().Err(err).Msg("starting ts listener")
		}
		log.Info().Msgf("SSHRecorder Listening over tailscale on :%s", c.SSHRecorder.Listener.Port)
	default:
		log.Fatal().Msgf("unknown listener type %s", c.Listener.Type)
	}
	g.Go(func() error {
		return serve(ctx, sr, ln)
	})

	err = g.Wait()
	if err != nil {
		log.Fatal().Err(err).Msg("error running server")
	}
}
