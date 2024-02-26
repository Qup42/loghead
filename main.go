package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/efekarakus/termcolor"
	"github.com/gorilla/mux"
	"github.com/qup42/loghead/processor"
	"github.com/qup42/loghead/ssh_recorder"
	"github.com/qup42/loghead/types"
	"github.com/qup42/loghead/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"io"
	"net"
	"net/http"
	"os"
	"tailscale.com/tsnet"
	"time"
)

type Loghead struct {
	Config        *types.Config
	MsgProcessors []processor.MsgProcessor
	LogProcessors []processor.LogProcessor
}

type FailableHandler func(http.ResponseWriter, *http.Request) error

func (fn FailableHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		log.Error().Err(err).Str("path", r.RequestURI).Msg("HTTP Request error")
		http.Error(w, err.Error(), 500)
	}
}

func (l *Loghead) LogHandler(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	collection := vars["collection"]
	private_id := vars["private_id"]

	msg, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("reading request body: %w", err)
	}
	if r.Header.Get("Content-Encoding") == "zstd" {
		msg = util.ZstdDecode(msg)
	}
	var maps []map[string]interface{}
	err = json.Unmarshal(msg, &maps)
	if err != nil {
		return fmt.Errorf("message unmarshal: %w", err)
	}
	log.Debug().Msgf("Received %d messages for %s/%s", len(maps)+1, collection, private_id)

	for _, m := range maps {
		msg := processor.LogtailMsg{
			Msg:       m,
			PrivateID: private_id,
		}
		for _, p := range l.MsgProcessors {
			p(msg)
		}
	}
	for _, p := range l.LogProcessors {
		p(msg)
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	log.Warn().Msgf("Unknown path %s called", r.RequestURI)

	w.WriteHeader(http.StatusNotFound)
}

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

func WaitTSReady(ctx context.Context, s *tsnet.Server) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := s.Up(ctx)
	if err != nil {
		return err
	}

	return nil
}

func startTSListener(ctx context.Context, r *mux.Router, c types.ListenerConfig) error {
	s := tsnet.Server{
		Logf:       func(string, ...any) {},
		AuthKey:    c.TS_AuthKey,
		ControlURL: c.TS_ControllURL,
	}
	defer s.Close()

	err := WaitTSReady(ctx, &s)
	if err != nil {
		return err
	}

	ln, err := s.Listen("tcp", fmt.Sprintf(":%s", c.Port))
	if err != nil {
		return err
	}
	defer ln.Close()

	ss := &http.Server{
		Handler:      r,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- ss.Serve(ln)
	}()
	log.Debug().Msg("TS ready and listening")

	select {
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = ss.Shutdown(ctx)
	case err = <-serveErr:
	}

	return err
}

func startPlainListener(ctx context.Context, r *mux.Router, c types.ListenerConfig) error {
	ss := &http.Server{
		Handler:      r,
		Addr:         net.JoinHostPort(c.Addr, c.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- ss.ListenAndServe()
	}()

	var err error
	select {
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = ss.Shutdown(ctx)
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

	r := mux.NewRouter()

	processors := []processor.MsgProcessor{}
	if c.Processors.FileLogger {
		fl, err := processor.NewFileLogger(c.FileLogger)
		if err != nil {
			log.Fatal().Err(err)
		}
		processors = append(processors, fl.Process)
	}
	if c.Processors.Metrics {
		cm := processor.NewClientMetrics()
		processors = append(processors, cm.Process)
		r.Path("/metrics").Handler(cm.PromHandler())
	}
	if c.Processors.Hostinfo {
		processors = append(processors, processor.Process)
	}

	logProcessors := []processor.LogProcessor{}
	if c.Processors.Forward != "" {
		log.Info().Msgf("Enableing forwarder to %s", c.Processors.Forward)
		fwd := processor.NewForwarder(c.Processors.Forward)
		logProcessors = append(logProcessors, fwd.Process)
	}

	l := Loghead{
		Config:        c,
		MsgProcessors: processors,
		LogProcessors: logProcessors,
	}

	g, ctx := errgroup.WithContext(context.Background())

	r.Handle("/c/{collection:[a-zA-Z0-9-_.]+}/{private_id:[0-9a-f]+}", FailableHandler(l.LogHandler)).Methods(http.MethodPost)
	r.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

	switch c.Listener.Type {
	case "plain":
		g.Go(func() error {
			return startPlainListener(ctx, r, c.Listener)
		})
		log.Info().Msgf("loghead Listening on %s:%s", c.Listener.Addr, c.Listener.Port)
		break
	case "tsnet":
		g.Go(func() error {
			return startTSListener(ctx, r, c.Listener)
		})
		log.Info().Msgf("loghead Listening over Tailscale on :%s", c.Listener.Port)
		break
	}

	rec, err := ssh_recorder.NewSSHRecorder(c.SSHRecorder)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not create SSH Recorder")
	}
	sr := mux.NewRouter()
	sr.Handle("/record", FailableHandler(rec.Handle))
	sr.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

	switch c.SSHRecorder.Listener.Type {
	case "plain":
		g.Go(func() error {
			return startPlainListener(ctx, sr, c.SSHRecorder.Listener)
		})
		log.Info().Msgf("SSHRecorder Listening on %s:%s", c.SSHRecorder.Listener.Addr, c.SSHRecorder.Listener.Port)
		break
	case "tsnet":
		g.Go(func() error {
			return startTSListener(ctx, sr, c.SSHRecorder.Listener)
		})
		log.Info().Msgf("SSHRecorder Listening over Tailscale on :%s", c.SSHRecorder.Listener.Port)
		break
	}

	err = g.Wait()
	if err != nil {
		log.Fatal().Err(err).Msg("error running server")
	}
}
