// Package service provides the tendermint service interface.
package service

import (
	"context"

	tmpubsub "github.com/tendermint/tendermint/libs/pubsub"
	tmrpctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/oasislabs/oasis-core/go/common/pubsub"
	"github.com/oasislabs/oasis-core/go/common/service"
	"github.com/oasislabs/oasis-core/go/consensus"
	"github.com/oasislabs/oasis-core/go/consensus/tendermint/abci"
	genesis "github.com/oasislabs/oasis-core/go/genesis/api"
)

// TendermintService provides Tendermint access to Oasis core backends.
type TendermintService interface {
	service.BackgroundService
	consensus.Backend

	// Started returns the channel that will be closed when the
	// tendermint service has been started.
	Started() <-chan struct{}

	// RegisterApplication registers an ABCI multiplexer application
	// with this service instance and check that its dependencies are
	// registered.
	RegisterApplication(abci.Application) error

	// GetGenesis will return the oasis genesis document.
	GetGenesis() *genesis.Document

	// GetHeight returns the Tendermint block height.
	GetHeight() (int64, error)

	// GetBlock returns the Tendermint block at the specified height.
	GetBlock(height *int64) (*tmtypes.Block, error)

	// GetBlockResults returns the ABCI results from processing a block
	// at a specific height.
	GetBlockResults(height *int64) (*tmrpctypes.ResultBlockResults, error)

	// WatchBlocks returns a stream of Tendermint blocks as they are
	// returned via the `EventDataNewBlock` query.
	WatchBlocks() (<-chan *tmtypes.Block, *pubsub.Subscription)

	// MarshalTx returns the Tendermint transaction from the inputs
	// that you would pass to BroadcastTx.
	MarshalTx(tag byte, tx interface{}) tmtypes.Tx

	// BroadcastTx broadcasts a transaction for Oasis ABCI application.
	//
	// The CBOR-encodable transaction together with the given application
	// tag is first marshalled and then transmitted using BroadcastTxSync.
	//
	// In case wait is true, the method will wait until the transaction
	// is finalized.
	//
	// Note that there is no implicit timeout -- if you need one, make
	// sure to cancel the context.
	BroadcastTx(ctx context.Context, tag byte, tx interface{}, wait bool) error

	// BroadcastEvidence broadcasts evidence of Tendermint node misbehavior.
	BroadcastEvidence(ctx context.Context, evidence tmtypes.Evidence) error

	// Subscribe subscribes to tendermint events.
	Subscribe(subscriber string, query tmpubsub.Query) (tmtypes.Subscription, error)

	// Unsubscribe unsubscribes from tendermint events.
	Unsubscribe(subscriber string, query tmpubsub.Query) error

	// Pruner returns the ABCI state pruner.
	Pruner() abci.StatePruner
}

// GenesisProvider is a tendermint specific genesis document provider.
type GenesisProvider interface {
	GetTendermintGenesisDocument() (*tmtypes.GenesisDoc, error)
}