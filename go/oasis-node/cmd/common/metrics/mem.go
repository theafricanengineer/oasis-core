package metrics

import (
	"fmt"
	"os"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs"
)

const (
	MetricMemVmSizeBytes       = "oasis_node_mem_vm_size_bytes"           // nolint: golint
	MetricMemVmSizeBytesHelp   = "Virtual memory size of worker (bytes)." // nolint: golint
	MetricMemRssAnonBytes      = "oasis_node_mem_rss_anon_bytes"
	MetricMemRssAnonBytesHelp  = "Size of resident anonymous memory of worker as reported by /proc/<PID>/status (bytes)."
	MetricMemRssFileBytes      = "oasis_node_mem_rss_file_bytes"
	MetricMemRssFileBytesHelp  = "Size of resident file mappings of worker as reported by /proc/<PID>/status (bytes)"
	MetricMemRssShmemBytes     = "oasis_node_mem_rss_shmem_bytes"
	MetricMemRssShmemBytesHelp = "Size of resident shared memory of worker."
)

var (
	vmSizeGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: MetricMemVmSizeBytes,
			Help: MetricMemVmSizeBytesHelp,
		},
	)

	rssAnonGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: MetricMemRssAnonBytes,
			Help: MetricMemRssAnonBytesHelp,
		},
	)

	rssFileGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: MetricMemRssFileBytes,
			Help: MetricMemRssFileBytesHelp,
		},
	)

	rssShmemGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: MetricMemRssShmemBytes,
			Help: MetricMemRssShmemBytesHelp,
		},
	)

	memCollectors  = []prometheus.Collector{vmSizeGauge, rssAnonGauge, rssFileGauge, rssShmemGauge}
	memServiceOnce sync.Once
)

type memCollector struct {
	// TODO: Should we monitor memory of children PIDs as well?
	pid int
}

func (m *memCollector) Name() string {
	return "mem"
}

func (m *memCollector) Update() error {
	// Obtain process Memory info.
	proc, err := procfs.NewProc(m.pid)
	if err != nil {
		return fmt.Errorf("memory metric: failed to obtain proc object for PID %d: %w", m.pid, err)
	}
	procStatus, err := proc.NewStatus()
	if err != nil {
		return fmt.Errorf("memory metric: failed to obtain procStatus object %d: %w", m.pid, err)
	}

	vmSizeGauge.Set(float64(procStatus.VmSize))
	rssAnonGauge.Set(float64(procStatus.RssAnon))
	rssFileGauge.Set(float64(procStatus.RssFile))
	rssShmemGauge.Set(float64(procStatus.RssShmem))

	return nil
}

// NewMemService constructs a new memory usage service.
//
// This service will regularly read memory info from process Status file.
func NewMemService() ResourceCollector {
	ms := &memCollector{
		pid: os.Getpid(),
	}

	// Memory metrics are singletons per process. Ensure to register them only once.
	memServiceOnce.Do(func() {
		prometheus.MustRegister(memCollectors...)
	})

	return ms
}
