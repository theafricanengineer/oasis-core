package debug

import (
	"crypto/rand"
	"math"
	"math/big"

	"github.com/oasislabs/oasis-core/go/common/crypto/signature"
	memorySigner "github.com/oasislabs/oasis-core/go/common/crypto/signature/signers/memory"
	"github.com/oasislabs/oasis-core/go/common/quantity"
	"github.com/oasislabs/oasis-core/go/staking/api"
)

var (
	DebugStateTestTotalSupply = QtyFromInt(math.MaxInt64)

	DebugGenesisState = api.Genesis{
		Parameters: api.ConsensusParameters{
			DebondingInterval: 1,
			Thresholds: map[api.ThresholdKind]quantity.Quantity{
				api.KindEntity:            QtyFromInt(1),
				api.KindNodeValidator:     QtyFromInt(2),
				api.KindNodeCompute:       QtyFromInt(3),
				api.KindNodeStorage:       QtyFromInt(4),
				api.KindNodeKeyManager:    QtyFromInt(5),
				api.KindRuntimeCompute:    QtyFromInt(6),
				api.KindRuntimeKeyManager: QtyFromInt(7),
			},
			Slashing: map[api.SlashReason]api.Slash{
				api.SlashDoubleSigning: api.Slash{
					Amount:         QtyFromInt(math.MaxInt64), // Slash everything.
					FreezeInterval: 1,
				},
			},
			MinDelegationAmount:     QtyFromInt(10),
			FeeSplitWeightVote:      QtyFromInt(1),
			RewardFactorEpochSigned: QtyFromInt(1),
			// Zero RewardFactorBlockProposed is normal.
		},
		TotalSupply: DebugStateTestTotalSupply,
		Ledger: map[signature.PublicKey]*api.Account{
			DebugStateSrcID: &api.Account{
				General: api.GeneralAccount{
					Balance: DebugStateTestTotalSupply,
				},
			},
		},
	}

	DebugStateSrcSigner = mustGenerateSigner()
	DebugStateSrcID     = DebugStateSrcSigner.Public()
	destSigner          = mustGenerateSigner()
	DebugStateDestID    = destSigner.Public()
)

func QtyFromInt(n int) quantity.Quantity {
	q := quantity.NewQuantity()
	if err := q.FromBigInt(big.NewInt(int64(n))); err != nil {
		panic(err)
	}
	return *q
}

func mustGenerateSigner() signature.Signer {
	k, err := memorySigner.NewSigner(rand.Reader)
	if err != nil {
		panic(err)
	}

	return k
}
