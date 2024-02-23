package main

import (
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

func (l *Loghead) LogHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collection := vars["collection"]
	private_id := vars["private_id"]

	msg, _ := io.ReadAll(r.Body)
	if r.Header.Get("Content-Encoding") == "zstd" {
		msg = util.ZstdDecode(msg)
	}
	var maps []map[string]interface{}
	err := json.Unmarshal(msg, &maps)
	if err != nil {
		log.Error().Err(err).Msg("Unmarshal failed")
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

func startTSListener(r *mux.Router, c types.ListenerConfig) error {
	s := tsnet.Server{
		Logf:       func(string, ...any) {},
		AuthKey:    c.TS_AuthKey,
		ControlURL: c.TS_ControllURL,
	}
	defer s.Close()

	ln, err := s.Listen("tcp", fmt.Sprintf(":%s", c.Port))
	if err != nil {
		log.Error().Err(err)
	}
	defer ln.Close()

	return http.Serve(ln, r)
}

func startPlainListener(r *mux.Router, c types.ListenerConfig) error {
	ss := &http.Server{
		Handler:      r,
		Addr:         net.JoinHostPort(c.Addr, c.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return ss.ListenAndServe()
}

func main() {
	SetupLogging()

	c := types.LoadConfig()
	zerolog.SetGlobalLevel(c.Log.Level)

	r := mux.NewRouter()

	processors := []processor.MsgProcessor{}
	if c.Processors.FileLogger {
		fl := processor.NewFileLogger(c.FileLogger)
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
		Config:        &c,
		MsgProcessors: processors,
		LogProcessors: logProcessors,
	}

	g := new(errgroup.Group)

	r.HandleFunc("/c/{collection:[a-zA-Z0-9-_.]+}/{private_id:[0-9a-f]+}", l.LogHandler).Methods(http.MethodPost)
	r.NotFoundHandler = http.HandlerFunc(NotFoundHandler)
	srv := &http.Server{
		Handler:      r,
		Addr:         c.Addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	g.Go(func() error {
		return srv.ListenAndServe()
	})

	log.Info().Msgf("Listening on %s", l.Config.Addr)

	rec := ssh_recorder.NewSSHRecorder(c.SSHRecorder)
	sr := mux.NewRouter()
	sr.HandleFunc("/record", rec.Handle)
	sr.NotFoundHandler = http.HandlerFunc(rec.Handle)

	switch c.SSHRecorder.Listener.Type {
	case "plain":
		g.Go(func() error {
			return startPlainListener(sr, c.SSHRecorder.Listener)
		})
		log.Info().Msgf("SSHRecorder Listening on %s:%s", c.SSHRecorder.Listener.Addr, c.SSHRecorder.Listener.Port)
		break
	case "tsnet":
		g.Go(func() error {
			return startTSListener(sr, c.SSHRecorder.Listener)
		})
		log.Info().Msgf("SSHRecorder Listening over Tailscale on :%s", c.SSHRecorder.Listener.Port)
		break
	}

	log.Fatal().Err(g.Wait()).Msg("Error running server")
}
