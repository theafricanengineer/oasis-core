package abci

import (
	"context"

	epochtime "github.com/oasislabs/oasis-core/go/epochtime/api"
	upgrade "github.com/oasislabs/oasis-core/go/upgrade/api"
)

// MockABCIMux exports some of the muxer's internal methods for testing use.
type MockABCIMux struct {
	*abciMux
}

// MockRegisterApp is used to register apps with this muxer during testing.
func (mux *MockABCIMux) MockRegisterApp(app Application) error {
	return mux.doRegister(app)
}

// MockSetEpochtime sets the timesource used by this muxer when testing.
func (mux *MockABCIMux) MockSetEpochtime(epochTime epochtime.Backend) {
	mux.state.timeSource = epochTime
}

// MockSetTransactionAuthHandler sets the transaction auth hander used by
// this muxer when testing.
func (mux *MockABCIMux) MockSetTransactionAuthHandler(handler TransactionAuthHandler) {
	mux.state.txAuthHandler = handler
}

// MockClose cleans up the muxer's state; it must be called once the muxer is no longer needed.
func (mux *MockABCIMux) MockClose() {
	mux.doCleanup()
}

// NewMockMux creates a new ABCI mux suitable for testing.
func NewMockMux(ctx context.Context, upgrader upgrade.Backend, cfg *ApplicationConfig) (*MockABCIMux, error) {
	mux, err := newABCIMux(ctx, upgrader, cfg)
	if err != nil {
		return nil, err
	}
	mockMux := &MockABCIMux{
		abciMux: mux,
	}
	return mockMux, nil
}
