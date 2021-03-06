// Oasis network integration test harness.
package main

import (
	"github.com/oasislabs/oasis-core/go/oasis-node/cmd/common"
	"github.com/oasislabs/oasis-core/go/oasis-test-runner/cmd"
	"github.com/oasislabs/oasis-core/go/oasis-test-runner/scenario/e2e"
	"github.com/oasislabs/oasis-core/go/oasis-test-runner/scenario/remotesigner"
)

func main() {
	// The general idea is that it should be possible to reuse everything
	// except for the main() function  and specialized test cases to write
	// test drivers that need to exercise the Oasis network or things built
	// up on the Oasis network.
	//
	// Other implementations will likely want to override parts of cmd.rootCmd,
	// in particular the `Use`, `Short`, and `Version` fields.

	// Register all scenarios and scenario parameters.
	for _, register := range []func() error{
		e2e.RegisterScenarios,
		remotesigner.RegisterScenarios,
	} {
		if err := register(); err != nil {
			common.EarlyLogAndExit(err)
		}
	}

	// Execute the command, now that everything has been initialized.
	cmd.Execute()
}
