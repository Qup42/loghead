package main

import (
	"encoding/json"
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
	"net/http"
	"os"
	"time"
)

type Loghead struct {
	Config     *types.Config
	Processors []processor.LogProcessor
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
		for _, p := range l.Processors {
			p(msg)
		}
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

func main() {
	SetupLogging()

	c := types.LoadConfig()
	zerolog.SetGlobalLevel(c.Log.Level)

	r := mux.NewRouter()

	processors := []processor.LogProcessor{}
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

	l := Loghead{
		Config:     &c,
		Processors: processors,
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
	ss := &http.Server{
		Handler:      sr,
		Addr:         c.SSHRecorder.Addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	g.Go(func() error {
		return ss.ListenAndServe()
	})

	log.Info().Msgf("SSHRecorder Listening on %s", l.Config.SSHRecorder.Addr)

	log.Fatal().Err(g.Wait()).Msg("Error running server")
}
