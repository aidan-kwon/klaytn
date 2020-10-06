package work

import (
	"github.com/klaytn/klaytn/blockchain/state"
	"github.com/klaytn/klaytn/blockchain/types"
)

type FakeWorker struct{}

// NewFakeWorker disables mining and block processing
//
// worker and istanbul engine will not be started.
func NewFakeWorker() *FakeWorker {
	logger.Warn("worker is disabled; no processing according to consensus logic")
	return &FakeWorker{}
}

func (*FakeWorker) Start()                                  {}
func (*FakeWorker) Stop()                                   {}
func (*FakeWorker) Register(Agent)                          {}
func (*FakeWorker) Mining() bool                            { return false }
func (*FakeWorker) HashRate() (tot int64)                   { return 0 }
func (*FakeWorker) SetExtra([]byte) error                   { return nil }
func (*FakeWorker) Pending() (*types.Block, *state.StateDB) { return nil, nil }
func (*FakeWorker) PendingBlock() *types.Block              { return nil }
