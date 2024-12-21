package main

import (
	"context"
	"github.com/efekarakus/termcolor"
	"github.com/gorilla/mux"
	"github.com/qup42/loghead/logs"
	"github.com/qup42/loghead/node_metrics"
	"github.com/qup42/loghead/ssh"
	"github.com/qup42/loghead/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
	"os"
	"os/signal"
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

func main() {
	SetupLogging()

	c, err := types.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	zerolog.SetGlobalLevel(c.Log.Level)

	log.Debug().Msgf("Config: %+v", c)

	var fls *logs.FileLoggerService
	var fwd *logs.ForwardingService
	var hs *logs.HostInfoService
	var ms *logs.MetricsService
	var rs *ssh.RecordingService
	if c.Loghead.Processors.FileLogger.Enabled {
		fls, err = logs.NewFileLoggerService(c.Loghead.Processors.FileLogger)
		if err != nil {
			log.Fatal().Err(err)
		}
	}
	if c.Loghead.Processors.Metrics {
		ms = logs.NewMetricsService()
	}
	if c.Loghead.Processors.Hostinfo {
		hs = &logs.HostInfoService{}
	}
	if c.Loghead.Processors.Forward.Enabled {
		log.Info().Msgf("Enableing forwarder to %s", c.Loghead.Processors.Forward.Addr)
		fwd = logs.NewForwardingService(c.Loghead.Processors.Forward.Addr)
	}
	rs, err = ssh.NewRecordingService(c.SSHRecorder)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not create SSH Recorder")
	}
	var nms *node_metrics.NodeMetricsService
	nms, err = node_metrics.NewNodeMetricsService(c.NodeMetrics)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not create Node Metrics")
	}

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	// logtail
	ltr := mux.NewRouter()
	addClientLogsRoutes(ltr, c, fwd, fls, hs, ms)

	logheadListener, err := types.MakeListener(ctx, c.Loghead.Listener, "loghead")
	if err != nil {
		log.Fatal().Err(err).Msg("Creating loghead listener")
	}
	defer logheadListener.Close()
	g.Go(func() error {
		return serve(ctx, ltr, logheadListener.Listener)
	})

	// SSH session recording
	sr := mux.NewRouter()
	addSSHRecordingRoutes(sr, rs)

	sshListener, err := types.MakeListener(ctx, c.SSHRecorder.Listener, "SSHRecorder")
	if err != nil {
		log.Fatal().Err(err).Msg("Creating SSHRecorder listener")
	}
	defer sshListener.Close()
	g.Go(func() error {
		return serve(ctx, sr, sshListener.Listener)
	})

	// Node metrics
	if c.NodeMetrics.Enabled {
		nm := mux.NewRouter()
		addNodeMetricsRoutes(nm, c, nms)

		nodeMetricsListener, err := types.MakeListener(ctx, c.NodeMetrics.Listener, "NodeMetrics")
		if err != nil {
			log.Fatal().Err(err).Msg("Creating NodeMetrics listener")
		}
		defer nodeMetricsListener.Close()
		g.Go(func() error {
			return serve(ctx, nm, nodeMetricsListener.Listener)
		})
	}

	err = g.Wait()
	if err != nil {
		log.Fatal().Err(err).Msg("error running server")
	}
}
