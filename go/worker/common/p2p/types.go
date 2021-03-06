package p2p

import (
	"github.com/oasislabs/oasis-core/go/common"
	"github.com/oasislabs/oasis-core/go/roothash/api/commitment"
)

// NOTE: Bump CommitteeProtocol version in go/common/version if you
//       change any of the structures below.

// Message is a message sent to nodes via P2P transport.
type Message struct {
	// RuntimeID is the identifier of the runtime this message
	// belongs to. It is used as a namespace.
	RuntimeID common.Namespace `json:"runtime_id,omitempty"`

	// GroupVersion is the version of all elected committees (the consensus
	// block height of last processed committee election). Messages with
	// non-matching group versions will be discarded.
	GroupVersion int64 `json:"group_version,omitempty"`

	// Jaeger's span context in binary format.
	SpanContext []byte `json:"span,omitempty"`

	// Message types.

	Ack   *Ack   `json:"ack,omitempty"`
	Error *Error `json:"err,omitempty"`

	SignedTxnSchedulerBatchDispatch *commitment.SignedTxnSchedulerBatchDispatch `json:",omitempty"`
	ExecutorWorkerFinished          *ExecutorWorkerFinished                     `json:",omitempty"`
}

// ExecutorWorkerFinished is the message sent from the executor workers to
// the merge committee after a batch has been processed and is ready to
// be merged.
type ExecutorWorkerFinished struct {
	// Commitment is an executor worker commitment.
	Commitment commitment.ExecutorCommitment
}

// Ack is an acknowledgement that a message was received.
type Ack struct {
}

// Error is an error response.
type Error struct {
	// Message is an error message.
	Message string
}
