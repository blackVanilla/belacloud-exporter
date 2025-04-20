// Belabox Cloud stats exporter
// Made by https://twitch.tv/murr1to
package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Stats struct {
	Publishers map[string]Publisher `json:"publishers"`
}

type Publisher struct {
	Connected   bool    `json:"connected"`
	Latency     float64 `json:"latency"`
	Network     float64 `json:"network"`
	Bitrate     float64 `json:"bitrate"`
	RTT         float64 `json:"rtt"`
	DroppedPkts float64 `json:"dropped_pkts"`
}

var (
	connectedGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "publisher_connected",
		Help: "Publisher connection status",
	}, []string{"url", "key", "server"})

	latencyGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "publisher_latency_ms",
		Help: "Latency in ms",
	}, []string{"url", "key", "server"})

	networkGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "publisher_network",
		Help: "Network usage",
	}, []string{"url", "key", "server"})

	bitrateGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "publisher_bitrate",
		Help: "Bitrate",
	}, []string{"url", "key", "server"})

	rttGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "publisher_rtt",
		Help: "Round Trip Time",
	}, []string{"url", "key", "server"})

	droppedPktsGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "publisher_dropped_packets",
		Help: "Dropped packets count",
	}, []string{"url", "key", "server"})
)

func init() {

}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	inputURL := r.URL.Query().Get("url")
	if inputURL == "" {
		http.Error(w, "url param required", http.StatusBadRequest)
		return
	}

	u, err := url.Parse(inputURL)
	if err != nil {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	key := parts[len(parts)-1]
	server := u.Hostname()

	resp, err := http.Get(inputURL)
	if err != nil {
		http.Error(w, "failed to fetch data", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusInternalServerError)
		return
	}

	var stats Stats
	if err := json.Unmarshal(body, &stats); err != nil {
		http.Error(w, "failed to parse json", http.StatusInternalServerError)
		return
	}

	publisher, ok := stats.Publishers[key]
	if !ok {
		http.Error(w, "publisher key not found", http.StatusNotFound)
		return
	}

	labels := prometheus.Labels{"url": inputURL, "key": key, "server": server}

	registry := prometheus.NewRegistry()
	connectedGaugeLocal := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "publisher_connected",
		Help:        "Publisher connection status",
		ConstLabels: labels,
	})
	connectedGaugeLocal.Set(boolToFloat(publisher.Connected))
	registry.MustRegister(connectedGaugeLocal)

	latencyGaugeLocal := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "publisher_latency_ms",
		Help:        "Latency in ms",
		ConstLabels: labels,
	})
	latencyGaugeLocal.Set(publisher.Latency)
	registry.MustRegister(latencyGaugeLocal)

	networkGaugeLocal := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "publisher_network",
		Help:        "Network usage",
		ConstLabels: labels,
	})
	networkGaugeLocal.Set(publisher.Network)
	registry.MustRegister(networkGaugeLocal)

	bitrateGaugeLocal := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "publisher_bitrate",
		Help:        "Bitrate",
		ConstLabels: labels,
	})
	bitrateGaugeLocal.Set(publisher.Bitrate)
	registry.MustRegister(bitrateGaugeLocal)

	rttGaugeLocal := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "publisher_rtt",
		Help:        "Round Trip Time",
		ConstLabels: labels,
	})
	rttGaugeLocal.Set(publisher.RTT)
	registry.MustRegister(rttGaugeLocal)

	droppedPktsGaugeLocal := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "publisher_dropped_packets",
		Help:        "Dropped packets count",
		ConstLabels: labels,
	})
	droppedPktsGaugeLocal.Set(publisher.DroppedPkts)
	registry.MustRegister(droppedPktsGaugeLocal)

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func main() {
	http.HandleFunc("/probe", metricsHandler)
	log.Println("Exporter started at :9090")
	log.Fatal(http.ListenAndServe(":9090", nil))
}
