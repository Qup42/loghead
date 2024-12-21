package logs

import (
	"reflect"
	"testing"
)

func TestReadVarint(t *testing.T) {
	type Result struct {
		n int
		c int
	}
	tests := []struct {
		in  string
		out Result
	}{
		{in: "02", out: Result{1, 2}},
		{in: "0202", out: Result{1, 2}},
		{in: "01", out: Result{-1, 2}},
		{in: "00", out: Result{0, 2}},
		{in: "46", out: Result{35, 2}},
		{in: "44c401", out: Result{34, 2}},
		{in: "c401", out: Result{98, 4}},
		{in: "c401ff", out: Result{98, 4}},
		{in: "feff7f", out: Result{1048575, 6}},
		{in: "ffff7f", out: Result{-1048576, 6}},
		{in: "feff7f01", out: Result{1048575, 6}},
		{in: "ffff7f01", out: Result{-1048576, 6}},
	}

	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			n, c := readVarint([]byte(tc.in))

			if (n != tc.out.n) || (c != tc.out.c) {
				t.Fatalf(`readVarint("%s") = (%d, %d), want (%d, %d)`, tc.in, n, c, tc.out.n, tc.out.c)
			}
		})
	}
}

func TestProcessMetrics(t *testing.T) {
	tests := []struct {
		in  string
		out map[int]Metric
	}{
		// Test cases extracted from real data. The data is arranged to make good test cases.
		{in: "N2anetmon_link_change_eqS0202", out: map[int]Metric{1: Metric{"netmon_link_change_eq", 1, 1, Counter}}},
		{in: "N20portmap_pcp_sentS0404", out: map[int]Metric{2: Metric{"portmap_pcp_sent", 2, 2, Counter}}},
		{in: "N20portmap_pmp_sentS0604", out: map[int]Metric{3: Metric{"portmap_pmp_sent", 3, 2, Counter}}},
		{in: "N22portmap_upnp_sentS0804", out: map[int]Metric{4: Metric{"portmap_upnp_sent", 4, 2, Counter}}},
		{in: "N1enetcheck_reportS0a04", out: map[int]Metric{5: Metric{"netcheck_report", 5, 2, Counter}}},
		{in: "N28netcheck_report_fullS0c02", out: map[int]Metric{6: Metric{"netcheck_report_full", 6, 1, Counter}}},
		{in: "N2enetcheck_stun_send_ipv4S0e62", out: map[int]Metric{7: Metric{"netcheck_stun_send_ipv4", 7, 49, Counter}}},
		{in: "N2enetcheck_stun_send_ipv6S1062", out: map[int]Metric{8: Metric{"netcheck_stun_send_ipv6", 8, 49, Counter}}},
		{in: "N2enetcheck_stun_recv_ipv4S1262", out: map[int]Metric{9: Metric{"netcheck_stun_recv_ipv4", 9, 49, Counter}}},
		{in: "N2enetcheck_stun_recv_ipv6S1462", out: map[int]Metric{10: Metric{"netcheck_stun_recv_ipv6", 10, 49, Counter}}},
		{in: "N4egauge_controlclient_map_requests_activeS1602", out: map[int]Metric{11: Metric{"gauge_controlclient_map_requests_active", 11, 1, Gauge}}},
		{in: "N34controlclient_map_requestsS180a", out: map[int]Metric{12: Metric{"controlclient_map_requests", 12, 5, Counter}}},
		{in: "N3econtrolclient_map_requests_liteS1a08", out: map[int]Metric{13: Metric{"controlclient_map_requests_lite", 13, 4, Counter}}},
		{in: "N20derp_home_changeS4a02", out: map[int]Metric{37: Metric{"derp_home_change", 37, 1, Counter}}},
		{in: "N24magicsock_send_udpS44c401", out: map[int]Metric{34: Metric{"magicsock_send_udp", 34, 98, Counter}}},
		{in: "N46gauge_dns_manager_linux_mode_directS4602", out: map[int]Metric{35: Metric{"gauge_dns_manager_linux_mode_direct", 35, 1, Gauge}}},
		{in: "N2anetmon_link_change_eqS0202N20portmap_pcp_sentS0404", out: map[int]Metric{1: Metric{"netmon_link_change_eq", 1, 1, Counter}, 2: Metric{"portmap_pcp_sent", 2, 2, Counter}}},
		{in: "N2anetmon_link_change_eqS0202I0202", out: map[int]Metric{1: Metric{"netmon_link_change_eq", 1, 2, Counter}}},
		{in: "N3cgauge_magicsock_num_derp_connsS3802I3802I3801", out: map[int]Metric{28: Metric{"gauge_magicsock_num_derp_conns", 28, 1, Gauge}}},
		{in: "N3cgauge_magicsock_num_derp_connsS3802I3802I3803", out: map[int]Metric{28: Metric{"gauge_magicsock_num_derp_conns", 28, 0, Gauge}}},
	}

	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			ms := NewMetricsService()
			ms.processMetrics(tc.in, "")
			out := map[string]map[int]Metric{"": tc.out}

			if !reflect.DeepEqual(out, ms.Metrics) {
				t.Fatalf(`processMetrics("%s") = %+v, want %+v`, tc.in, ms.Metrics, out)
			}
		})
	}
}
