package metrics

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs"
)

const (
	MetricNetReceiveBytesTotal        = "oasis_node_net_receive_bytes_total" // godoc: metric
	MetricNetReceiveBytesTotalHelp    = "Received data for each network device as reported by /proc/net/dev (bytes)."
	MetricNetReceivePacketsTotal      = "oasis_node_net_receive_packets_total" // godoc: metric
	MetricNetReceivePacketsTotalHelp  = "Received data for each network device as reported by /proc/net/dev (packets)."
	MetricNetTransmitBytesTotal       = "oasis_node_net_transmit_bytes_total" // godoc: metric
	MetricNetTransmitBytesTotalHelp   = "Transmitted data for each network device as reported by /proc/net/dev (bytes)."
	MetricNetTransmitPacketsTotal     = "oasis_node_net_transmit_packets_total" // godoc: metric
	MetricNetTransmitPacketsTotalHelp = "Transmitted data for each network device as reported by /proc/net/dev (packets)."
)

var (
	receiveBytesGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: MetricNetReceiveBytesTotal,
			Help: MetricNetReceiveBytesTotalHelp,
		},
		[]string{
			// Interface name, e.g. eth0.
			"device",
		})

	receivePacketsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: MetricNetReceivePacketsTotal,
			Help: MetricNetReceivePacketsTotalHelp,
		},
		[]string{
			// Interface name, e.g. eth0.
			"device",
		})

	transmitBytesGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: MetricNetTransmitBytesTotal,
			Help: MetricNetTransmitBytesTotalHelp,
		},
		[]string{
			// Interface name, e.g. eth0.
			"device",
		})

	transmitPacketsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: MetricNetTransmitPacketsTotal,
			Help: MetricNetTransmitPacketsTotalHelp,
		},
		[]string{
			// Interface name, e.g. eth0.
			"device",
		})

	netCollectors  = []prometheus.Collector{receiveBytesGauge, receivePacketsGauge, transmitBytesGauge, transmitPacketsGauge}
	netServiceOnce sync.Once
)

type netCollector struct {
}

func (n *netCollector) Name() string {
	return "net"
}

func (n *netCollector) Update() error {
	// Obtain process Network info.
	proc, err := procfs.NewDefaultFS()
	if err != nil {
		return fmt.Errorf("network metric: failed to obtain proc object: %w", err)
	}
	netDevs, err := proc.NetDev()
	if err != nil {
		return fmt.Errorf("network metric: failed to obtain netDevs object: %w", err)
	}

	for _, netDev := range netDevs {
		receiveBytesGauge.WithLabelValues(netDev.Name).Set(float64(netDev.RxBytes))
		receivePacketsGauge.WithLabelValues(netDev.Name).Set(float64(netDev.RxPackets))
		transmitBytesGauge.WithLabelValues(netDev.Name).Set(float64(netDev.TxBytes))
		transmitPacketsGauge.WithLabelValues(netDev.Name).Set(float64(netDev.TxPackets))
	}

	return nil
}

// NewNetService constructs a new network statistics service.
//
// This service will regularly read info from /proc/net/dev file.
func NewNetService() ResourceCollector {
	ns := &netCollector{}

	netServiceOnce.Do(func() {
		prometheus.MustRegister(netCollectors...)
	})

	return ns
}
