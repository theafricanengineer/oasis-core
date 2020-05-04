package registry

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/oasislabs/oasis-core/go/common/crypto/signature"
	fileSigner "github.com/oasislabs/oasis-core/go/common/crypto/signature/signers/file"
	"github.com/oasislabs/oasis-core/go/common/identity"
	"github.com/oasislabs/oasis-core/go/common/logging"
	consensus "github.com/oasislabs/oasis-core/go/consensus/api"
	cmdCommon "github.com/oasislabs/oasis-core/go/oasis-node/cmd/common"
	"github.com/oasislabs/oasis-core/go/registry/api"
)

const (
	metricsUpdateInterval = 10 * time.Second

	MetricRegistryNodes              = "oasis_registry_nodes" // godoc: metric
	MetricRegistryNodesHelp          = "Number of registry nodes."
	MetricRegistryEntities           = "oasis_registry_entities" // godoc: metric
	MetricRegistryEntitiesHelp       = "Number of registry entities."
	MetricRegistryRuntimes           = "oasis_registry_runtimes" // godoc: metric
	MetricRegistryRuntimesHelp       = "Number of registry runtimes."
	MetricRegistryNodeRegistered     = "oasis_registry_node_registered" // godoc: metric
	MetricRegistryNodeRegisteredHelp = "Is oasis node registered (binary)."
)

var (
	registryNodes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: MetricRegistryNodes,
			Help: MetricRegistryNodesHelp,
		},
	)
	registryEntities = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: MetricRegistryEntities,
			Help: MetricRegistryEntitiesHelp,
		},
	)
	registryRuntimes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: MetricRegistryRuntimes,
			Help: MetricRegistryRuntimesHelp,
		},
	)
	registryNodeRegistered = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: MetricRegistryNodeRegistered,
			Help: MetricRegistryNodeRegisteredHelp,
		},
	)
	registryCollectors = []prometheus.Collector{
		registryNodes,
		registryEntities,
		registryRuntimes,
		registryNodeRegistered,
	}

	metricsOnce sync.Once
)

// MetricsUpdater is a registry metric updater.
type MetricsUpdater struct {
	logger *logging.Logger

	backend api.Backend

	closeOnce sync.Once
	closeCh   chan struct{}
	closedCh  chan struct{}
}

// Cleanup performs cleanup.
func (m *MetricsUpdater) Cleanup() {
	m.closeOnce.Do(func() {
		close(m.closeCh)
		<-m.closedCh
	})
}

func (m *MetricsUpdater) worker(ctx context.Context) {
	defer close(m.closedCh)

	t := time.NewTicker(metricsUpdateInterval)
	defer t.Stop()

	runtimeCh, sub, err := m.backend.WatchRuntimes(ctx)
	if err != nil {
		m.logger.Error("failed to watch runtimes, metrics will not be updated",
			"err", err,
		)
		return
	}
	defer sub.Close()

	for {
		select {
		case <-m.closeCh:
			return
		case <-runtimeCh:
			registryRuntimes.Inc()
			continue
		case <-t.C:
		}

		m.updatePeriodicMetrics(ctx)
	}
}

func (m *MetricsUpdater) updatePeriodicMetrics(ctx context.Context) {
	nodes, err := m.backend.GetNodes(ctx, consensus.HeightLatest)
	if err == nil {
		registryNodes.Set(float64(len(nodes)))
	}

	entities, err := m.backend.GetEntities(ctx, consensus.HeightLatest)
	if err == nil {
		registryEntities.Set(float64(len(entities)))
	}

	// Check, if node is a registered validator.
	// XXX: Can we obtain existing nodeIdentity from consensus code directly?
	dataDir, err := cmdCommon.DataDirOrPwd()
	if err != nil {
		m.logger.Error("failed to query data directory",
			"err", err,
		)
		return
	}
	nodeSignerFactory, err := fileSigner.NewFactory(dataDir, signature.SignerNode, signature.SignerP2P, signature.SignerConsensus)
	if err != nil {
		m.logger.Error("failed to create node identity signer factory",
			"err", err,
		)
		return
	}
	nodeIdentity, err := identity.Load(dataDir, nodeSignerFactory)
	if err != nil {
		m.logger.Error("failed to load node identity",
			"err", err,
		)
		return
	}
	registered := false
	for _, node := range nodes {
		if node.ID.Equal(nodeIdentity.NodeSigner.Public()) {
			registered = true
			break
		}
	}
	if registered {
		registryNodeRegistered.Set(float64(1.0))
	} else {
		registryNodeRegistered.Set(float64(0.0))
	}
}

// NewMetricsUpdater creates a new registry metrics updater.
func NewMetricsUpdater(ctx context.Context, backend api.Backend) *MetricsUpdater {
	metricsOnce.Do(func() {
		prometheus.MustRegister(registryCollectors...)
	})

	m := &MetricsUpdater{
		logger:   logging.GetLogger("go/registry/metrics"),
		backend:  backend,
		closeCh:  make(chan struct{}),
		closedCh: make(chan struct{}),
	}

	m.updatePeriodicMetrics(ctx)
	go m.worker(ctx)

	return m
}
