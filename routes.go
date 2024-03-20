package main

import (
	"encoding/json"
	"github.com/cockroachdb/errors"
	"github.com/gorilla/mux"
	"github.com/qup42/loghead/processor"
	"github.com/qup42/loghead/ssh"
	"github.com/qup42/loghead/types"
	"github.com/qup42/loghead/util"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"time"
)

type FailableHandler func(http.ResponseWriter, *http.Request) error

func (fn FailableHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		log.Error().Err(err).Str("path", r.RequestURI).Msg("HTTP Request error")
		http.Error(w, err.Error(), 500)
	}
}

func addSSHRecordingRoutes(
	r *mux.Router,
	rs *ssh.RecordingService) {
	r.Handle("/record", handleSSHRecording(rs))
	r.NotFoundHandler = handleNotFound()
}

func addClientLogsRoutes(
	r *mux.Router,
	c *types.Config,
	fwd *processor.ForwardingService,
	fl *processor.FileLoggerService,
	hi *processor.HostInfoService,
	ms *processor.MetricsService) {

	r.Handle("/c/{collection:[a-zA-Z0-9-_.]+}/{private_id:[0-9a-f]+}", handleTailnodeLogs(fwd, fl, hi, ms)).Methods(http.MethodPost)
	if c.Loghead.Processors.Metrics {
		r.Handle("/metrics", handleMetrics(ms))
	}
	r.NotFoundHandler = handleNotFound()
}

func handleSSHRecording(rec *ssh.RecordingService) http.Handler {
	return FailableHandler(func(w http.ResponseWriter, r *http.Request) error {
		log.Trace().Msg("Starting SSH Session recording")
		rc := http.NewResponseController(w)
		// this is a streaming request, disable the read deadline
		err := rc.SetReadDeadline(time.Time{})
		if err != nil {
			return errors.Errorf("setting read deadline: %w", err)
		}
		err = rec.Record(r.Body)
		if err != nil {
			return errors.Errorf("recording sesion: %w", err)
		}
		log.Trace().Msg("SSH Session recording finished")
		return nil
	})
}

func handleTailnodeLogs(
	fwd *processor.ForwardingService,
	fl *processor.FileLoggerService,
	hi *processor.HostInfoService,
	ms *processor.MetricsService) http.Handler {
	return FailableHandler(func(w http.ResponseWriter, r *http.Request) error {
		vars := mux.Vars(r)
		collection := vars["collection"]
		private_id := vars["private_id"]

		msg, err := io.ReadAll(r.Body)
		if err != nil {
			return errors.Errorf("reading request body: %w", err)
		}
		if r.Header.Get("Content-Encoding") == "zstd" {
			msg = util.ZstdDecode(msg)
		}

		if fwd != nil {
			err := fwd.Forward(msg)
			if err != nil {
				log.Error().Err(err).Msg("error forwarding")
			}
		}

		var maps []map[string]interface{}
		err = json.Unmarshal(msg, &maps)
		if err != nil {
			return errors.Errorf("message unmarshal: %w", err)
		}
		log.Debug().Msgf("Received %d messages for %s/%s", len(maps)+1, collection, private_id)

		for _, m := range maps {
			msg := processor.LogtailMsg{
				Msg:        m,
				Collection: collection,
				PrivateID:  private_id,
			}
			if fl != nil {
				fl.Log(msg)
			}
			if hi != nil {
				hi.Process(msg)
			}
			if ms != nil {
				ms.Process(msg)
			}
		}

		w.WriteHeader(http.StatusOK)
		return nil
	})
}

func handleMetrics(ms *processor.MetricsService) http.Handler {
	return ms.PromHandler()
}

func handleNotFound() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Warn().Msgf("Unknown path %s called", r.RequestURI)

		w.WriteHeader(http.StatusNotFound)
	})
}
