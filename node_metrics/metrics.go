package node_metrics

import (
	"github.com/cockroachdb/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/qup42/loghead/types"
	"github.com/rs/zerolog/log"
	"net"
	"net/http"
	"net/url"
)

type NodeMetricsService struct {
	Gatherers prometheus.Gatherers
}

func NewNodeMetricsService(config types.NodeMetricsConfig) (*NodeMetricsService, error) {
	gatherers := make([]prometheus.Gatherer, 0, len(config.Targets))
	for _, target := range config.Targets {
		name, port, err := net.SplitHostPort(target)
		if err != nil {
			return nil, errors.Errorf("invalid target \"%s\": %w", target, err)
		}
		if port == "" && name != "100.100.100.100" {
			port = "5252"
			log.Debug().Msgf("Defaulting port to \"5252\" for target: %s", target)
		}
		host := net.JoinHostPort(name, port)
		var targetUrl url.URL
		targetUrl.Scheme = "http"
		targetUrl.Host = host
		targetUrl.Path = "metrics"
		log.Debug().Msgf("Adding target: %s", targetUrl.String())

		gatherers = append(gatherers, prometheus.GathererFunc(func() ([]*dto.MetricFamily, error) {
			metrics, err := CollectTarget(targetUrl.String())
			if err != nil {
				return nil, err
			}
			labelName := "target"
			metricsList := make([]*dto.MetricFamily, 0, len(metrics))
			for _, metricFamily := range metrics {
				for _, metric := range metricFamily.Metric {
					metric.Label = append(metric.Label, &dto.LabelPair{
						Name:  &labelName,
						Value: &host,
					})
				}
				metricsList = append(metricsList, metricFamily)
			}
			return metricsList, nil
		}))
	}
	return &NodeMetricsService{
		Gatherers: gatherers,
	}, nil
}

func (n *NodeMetricsService) PromHandler() http.Handler {
	return promhttp.HandlerFor(n.Gatherers, promhttp.HandlerOpts{})
}

func CollectTarget(target string) (map[string]*dto.MetricFamily, error) {
	resp, err := http.Get(target)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	parser := expfmt.TextParser{}
	metrics, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}
