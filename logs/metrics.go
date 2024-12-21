package logs

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
)

type MetricType int64

const (
	Gauge MetricType = iota
	Counter
)

type Metric struct {
	Name   string
	WireID int
	Value  int
	Type   MetricType
}

type MetricsService struct {
	Registry           *prometheus.Registry
	Metrics            map[string]map[int]Metric
	GaugePromMetrics   map[string]*prometheus.GaugeVec
	CounterPromMetrics map[string]*prometheus.CounterVec
}

func NewMetricsService() *MetricsService {
	return &MetricsService{
		Registry:           prometheus.NewRegistry(),
		Metrics:            map[string]map[int]Metric{},
		GaugePromMetrics:   map[string]*prometheus.GaugeVec{},
		CounterPromMetrics: map[string]*prometheus.CounterVec{},
	}
}

func readByte(in []byte) (byte, int) {
	b := make([]byte, 1)
	hex.Decode(b, []byte(in[0:2]))
	return b[0], 2
}

func readVarint(in []byte) (int, int) {
	bs := make([]byte, 0)
	i := 0

	//	log.Trace().Msgf("Processing: %08b, %s", in[i:i+2], in[i:i+2])
	b, ii := readByte(in[i:])
	bs = append(bs, b)
	i += ii

	for (i+2 <= len(in)) && ((b >> 7) == 1) {
		//        log.Trace().Msgf("Continuation: %08b, %s", in[i:i+2], in[i:i+2])
		b, ii = readByte(in[i:])
		bs = append(bs, b)
		//		log.Trace().Msgf("Got byte %08b", b)
		i += ii
	}
	//	log.Trace().Msgf("Collected bytes %08b", bs)
	n, _ := binary.Varint(bs)
	//	log.Trace().Msgf("Read n=%d", n)
	return int(n), i
}

func (ms *MetricsService) createGauge(m Metric, private_id string) {
	metric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: m.Name,
		},
		[]string{"private_id"})
	metric.With(prometheus.Labels{"private_id": private_id}).Set(float64(m.Value))
	ms.GaugePromMetrics[m.Name] = metric
}

func (ms *MetricsService) createCounter(m Metric, private_id string) {
	metric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: m.Name,
		},
		[]string{"private_id"})
	metric.With(prometheus.Labels{"private_id": private_id}).Add(float64(m.Value))
	ms.CounterPromMetrics[m.Name] = metric
}

func (ms *MetricsService) registerMetric(m Metric, private_id string) {
	switch m.Type {
	case Gauge:
		if _, ok := ms.GaugePromMetrics[m.Name]; ok {
			log.Debug().Msgf("Metric `%s` already registered.", m.Name)
		} else {
			ms.createGauge(m, private_id)
			ms.Registry.MustRegister(ms.GaugePromMetrics[m.Name])
		}
		break
	case Counter:
		if _, ok := ms.CounterPromMetrics[m.Name]; ok {
			log.Debug().Msgf("Metric `%s` already registered.", m.Name)
		} else {
			ms.createCounter(m, private_id)
			ms.Registry.MustRegister(ms.CounterPromMetrics[m.Name])
		}
		break
	}
}

func (ms *MetricsService) updateMetric(m Metric, private_id string, value int) {
	switch m.Type {
	case Gauge:
		log.Debug().Msgf("%s := %d", m.Name, m.Value)
		ms.GaugePromMetrics[m.Name].With(prometheus.Labels{"private_id": private_id}).Add(float64(value))
		break
	case Counter:
		if value < 0 {
			log.Error().Msgf("%s delta is negative (%+d)", m.Name, value)
			return
		}
		log.Debug().Msgf("%s -(%+d)-> %d", m.Name, value, m.Value)
		ms.CounterPromMetrics[m.Name].With(prometheus.Labels{"private_id": private_id}).Add(float64(value))
		break
	}
}

func (ms *MetricsService) Process(msg LogtailMsg) {
	if metrics, ok := msg.Msg["metrics"]; ok {
		metricss := metrics.(string)
		ms.processMetrics(metricss, msg.PrivateID)
	}
}

func (ms *MetricsService) processMetrics(in string, private_id string) {
	var m *Metric = nil
	cmetrics, ok := ms.Metrics[private_id]
	if !ok {
		cmetrics = map[int]Metric{}
		ms.Metrics[private_id] = cmetrics
	}
	i := 0
	for i < len(in) {
		switch in[i] {
		case 'N':
			i += 1
			n, ii := readVarint([]byte(in[i:]))
			i += ii
			name := in[i : i+n]
			var t MetricType
			if strings.HasPrefix(name, "gauge_") {
				t = Gauge
			} else {
				t = Counter
			}
			m = &Metric{name, -1, -1, t}
			i += n
			break
		case 'S':
			i += 1
			w, ii := readVarint([]byte(in[i:]))
			i += ii
			v, ii := readVarint([]byte(in[i:]))
			i += ii
			if entry, ok := cmetrics[w]; ok {
				log.Debug().Msgf("Set `%s` to %d", entry.Name, v)
				entry.Value = v
				cmetrics[w] = entry
			} else if m != nil {
				m.WireID = w
				m.Value = v
				mm := *m
				cmetrics[w] = mm
				log.Info().Msgf("Registered Metric `%s` (%d) with init %d", mm.Name, mm.WireID, mm.Value)
				ms.registerMetric(mm, private_id)
				m = nil
			} else {
				log.Warn().Msgf("WireID %d unknown", w)
			}
			break
		case 'I':
			i += 1
			w, ii := readVarint([]byte(in[i:]))
			i += ii
			v, ii := readVarint([]byte(in[i:]))
			i += ii
			if entry, ok := cmetrics[w]; ok {
				entry.Value += v
				ms.updateMetric(entry, private_id, v)
				cmetrics[w] = entry
			} else {
				log.Warn().Msgf("WireID %d unknown", w)
			}
			break
		}
	}
}

func (ms *MetricsService) PromHandler() http.Handler {
	return promhttp.HandlerFor(ms.Registry, promhttp.HandlerOpts{Registry: ms.Registry})
}
