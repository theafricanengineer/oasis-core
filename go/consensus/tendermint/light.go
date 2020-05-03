package tendermint

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	tmlite "github.com/tendermint/tendermint/lite2"
	tmliteprovider "github.com/tendermint/tendermint/lite2/provider"
	tmtypes "github.com/tendermint/tendermint/types"

	cmnGrpc "github.com/oasislabs/oasis-core/go/common/grpc"
	consensusAPI "github.com/oasislabs/oasis-core/go/consensus/api"
)

// lightClientProvider implements Tendermint's light client provider interface using the Oasis Core
// light client API.
type lightClientProvider struct {
	ctx context.Context

	chainID string
	client  consensusAPI.LightClientBackend
}

// Implements tmliteprovider.Provider.
func (lp *lightClientProvider) ChainID() string {
	return lp.chainID
}

// Implements tmliteprovider.Provider.
func (lp *lightClientProvider) SignedHeader(height int64) (*tmtypes.SignedHeader, error) {
	shdr, err := lp.client.GetSignedHeader(lp.ctx, height)
	switch {
	case err == nil:
	case errors.Is(err, consensusAPI.ErrVersionNotFound):
		return nil, tmliteprovider.ErrSignedHeaderNotFound
	default:
		return nil, fmt.Errorf("failed to fetch signed header: %w", err)
	}

	// Decode Tendermint-specific signed header.
	var sh tmtypes.SignedHeader
	if err = aminoCodec.UnmarshalBinaryBare(shdr.Meta, &sh); err != nil {
		return nil, fmt.Errorf("received malformed header: %w", err)
	}

	if lp.chainID != sh.ChainID {
		return nil, fmt.Errorf("incorrect chain ID (expected: %s got: %s)",
			lp.chainID,
			sh.ChainID,
		)
	}

	return &sh, nil
}

// Implements tmliteprovider.Provider.
func (lp *lightClientProvider) ValidatorSet(height int64) (*tmtypes.ValidatorSet, error) {
	vs, err := lp.client.GetValidatorSet(lp.ctx, height)
	switch {
	case err == nil:
	case errors.Is(err, consensusAPI.ErrVersionNotFound):
		return nil, tmliteprovider.ErrValidatorSetNotFound
	default:
		return nil, fmt.Errorf("failed to fetch validator set: %w", err)
	}

	// Decode Tendermint-specific validator set.
	var vals tmtypes.ValidatorSet
	if err = aminoCodec.UnmarshalBinaryBare(vs.Meta, &vals); err != nil {
		return nil, fmt.Errorf("received malformed validator set: %w", err)
	}

	return &vals, nil
}

// Implements tmliteprovider.Provider.
func (lp *lightClientProvider) ReportEvidence(ev tmtypes.Evidence) error {
	// TODO: Implement SubmitEvidence.
	return fmt.Errorf("not yet implemented")
}

// newLightClientProvider creates a new provider for the Tendermint's light client.
//
// The provided chain ID must be the Tendermint chain ID.
func newLightClientProvider(
	ctx context.Context,
	chainID string,
	address string,
) (tmliteprovider.Provider, error) {
	creds := credentials.NewTLS(&tls.Config{
		// XXX: As a defense in depth measure this should validate certificates. However the light
		//      endpoint is untrusted and is verified through other means.
		//
		//      This should be easier once we start support validating only by Ed25519 public keys
		//      as those are easier to input than full certificates (see oasis-core#2556).
		InsecureSkipVerify: true,
	})
	conn, err := cmnGrpc.Dial(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to dial public consensus service endpoint %s: %w", address, err)
	}

	return &lightClientProvider{
		ctx:     ctx,
		chainID: chainID,
		client:  consensusAPI.NewConsensusLightClient(conn),
	}, nil
}

// lightService is a Tendermint consensus service that uses the light client API to talk with a
// remote Tendermint node and verify responses.
//
// This should eventually become a replacement for the full node tendermintService.
type lightService struct {
	// lc is the Tendermint light client used for verifying headers.
	lc *tmlite.Client
	// client is the consensus light client backend connected to a remote node.
	client consensusAPI.LightClientBackend
}

func (ls *lightService) getParameters(ctx context.Context, height int64) (*tmtypes.ConsensusParams, error) {
	p, err := ls.client.GetParameters(ctx, height)
	if err != nil {
		return nil, err
	}
	if p.Height <= 0 {
		return nil, fmt.Errorf("malformed height in response: %d", p.Height)
	}

	// Decode Tendermint-specific parameters.
	var params tmtypes.ConsensusParams
	if err = aminoCodec.UnmarshalBinaryBare(p.Meta, &params); err != nil {
		return nil, fmt.Errorf("malformed parameters: %w", err)
	}
	if err = params.Validate(); err != nil {
		return nil, fmt.Errorf("malformed parameters: %w", err)
	}

	// Fetch the header from the light client.
	h, err := ls.lc.VerifyHeaderAtHeight(p.Height, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch header %d from light client: %w", p.Height, err)
	}

	// Verify hash.
	if localHash := params.Hash(); !bytes.Equal(localHash, h.ConsensusHash) {
		return nil, fmt.Errorf("mismatched parameters hash (expected: %X got: %X)",
			h.ConsensusHash,
			localHash,
		)
	}

	return &params, nil
}

// newLightService creates a light Tendermint consensus service.
func newLightService(client consensusAPI.LightClientBackend, lc *tmlite.Client) (*lightService, error) {
	return &lightService{
		lc:     lc,
		client: client,
	}, nil
}
