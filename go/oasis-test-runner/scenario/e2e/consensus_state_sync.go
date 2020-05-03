package e2e

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	consensus "github.com/oasislabs/oasis-core/go/consensus/api"
	"github.com/oasislabs/oasis-core/go/oasis-test-runner/env"
	"github.com/oasislabs/oasis-core/go/oasis-test-runner/oasis"
	"github.com/oasislabs/oasis-core/go/oasis-test-runner/scenario"
)

var (
	// ConsensusStateSync is the consensus state sync scenario.
	ConsensusStateSync scenario.Scenario = &consensusStateSyncImpl{
		runtimeImpl: *newRuntimeImpl("consensus-state-sync", "", nil),
	}
)

type consensusStateSyncImpl struct {
	runtimeImpl
}

func (sc *consensusStateSyncImpl) Clone() scenario.Scenario {
	return &consensusStateSyncImpl{
		runtimeImpl: *sc.runtimeImpl.Clone().(*runtimeImpl),
	}
}

func (sc *consensusStateSyncImpl) Fixture() (*oasis.NetworkFixture, error) {
	f, err := sc.runtimeImpl.Fixture()
	if err != nil {
		return nil, err
	}

	// We don't need any runtimes.
	f.Runtimes = nil
	f.Keymanagers = nil
	f.StorageWorkers = nil
	f.ComputeWorkers = nil
	f.Sentries = nil

	// Enable public consensus RPC services worker so that the nodes can be used for light clients.
	f.Validators = []oasis.ValidatorFixture{
		oasis.ValidatorFixture{Entity: 1, Consensus: oasis.ConsensusFixture{EnableConsensusRPCWorker: true}},
		oasis.ValidatorFixture{Entity: 1, Consensus: oasis.ConsensusFixture{EnableConsensusRPCWorker: true}},
		oasis.ValidatorFixture{Entity: 1, Consensus: oasis.ConsensusFixture{EnableConsensusRPCWorker: true}},
		oasis.ValidatorFixture{Entity: 1, Consensus: oasis.ConsensusFixture{EnableConsensusRPCWorker: true}},
	}

	return f, nil
}

func (sc *consensusStateSyncImpl) Run(childEnv *env.Env) error {
	if err := sc.net.Start(); err != nil {
		return err
	}

	sc.logger.Info("waiting for network to come up")
	ctx := context.Background()
	if err := sc.net.Controller().WaitNodesRegistered(ctx, len(sc.net.Validators())); err != nil {
		return err
	}

	// Stop one of the validators.
	val := sc.net.Validators()[2]
	if err := val.Stop(); err != nil {
		return fmt.Errorf("failed to stop validator: %w", err)
	}

	// Let the network run for 50 blocks. This should generate some checkpoints.
	blockCh, blockSub, err := sc.net.Controller().Consensus.WatchBlocks(ctx)
	if err != nil {
		return err
	}
	defer blockSub.Close()

	sc.logger.Info("waiting for some blocks")
	var blk *consensus.Block
	for {
		select {
		case blk = <-blockCh:
			if blk.Height < 50 {
				continue
			}
		case <-time.After(30 * time.Second):
			return fmt.Errorf("timed out waiting for blocks")
		}

		break
	}

	sc.logger.Info("got some blocks, starting the validator back",
		"trust_height", blk.Height,
		"trust_hash", hex.EncodeToString(blk.Hash),
	)

	// Configure state sync for the consensus validator.
	val.SetConsensusStateSync(&oasis.ConsensusStateSyncCfg{
		ConsensusNodes: []string{
			sc.net.Validators()[0].ExternalGRPCAddress(),
			sc.net.Validators()[1].ExternalGRPCAddress(),
		},
		TrustHeight: uint64(blk.Height),
		TrustHash:   hex.EncodeToString(blk.Hash),
	})

	if err := val.Start(); err != nil {
		return fmt.Errorf("failed to start validator back: %w", err)
	}

	// Wait for the validator to finish syncing.
	sc.logger.Info("waiting for the validator to sync")
	valCtrl, err := oasis.NewController(val.SocketPath())
	if err != nil {
		return err
	}
	if err = valCtrl.WaitSync(ctx); err != nil {
		return err
	}

	return nil
}
