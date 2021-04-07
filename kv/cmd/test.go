package cmd

import (
	"flag"
	"testing"
	"time"

	kvtest "github.com/Jonas-Heinrich/toy-distributed-key-value/test"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(testCmd)
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the test container",
	Long:  `Run the code specified in "/test/"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Wait for other containers to boot
		time.Sleep(1 * time.Second)

		flag.Set("test.v", "true")
		testing.Main(func(pat, str string) (bool, error) { return true, nil },
			[]testing.InternalTest{
				// Essential tests: cannot be left out
				{"TestTesting", kvtest.TestTesting},
				{"TestStatus", kvtest.TestStatus},
				{"TestInitialState", kvtest.TestInitialState},

				// Essential tests: cannot be left out
				{"TestDirectNetworkEntry", kvtest.TestDirectNetworkEntry},
				{"TestIndirectNetworkEntry", kvtest.TestIndirectNetworkEntry},
				{"TestNetworkEntryLogReplication", kvtest.TestNetworkEntryLogReplication},
				{"TestRemainingNetworkEntry", kvtest.TestRemainingNetworkEntry},

				// Leader Election
				{"TestLeaderElection", kvtest.TestLeaderElection},

				// Read
				{"TestInitialDirectRead", kvtest.TestInitialDirectRead},
				{"TestInitialIndirectRead", kvtest.TestInitialIndirectRead},
				{"TestInitialDirectReadNotFound", kvtest.TestInitialDirectReadNotFound},
				{"TestInitialIndirectReadNotFound", kvtest.TestInitialIndirectReadNotFound},

				// Write
				{"TestDirectWrite", kvtest.TestDirectWrite},
				{"TestIndirectWrite", kvtest.TestIndirectWrite},
			},
			[]testing.InternalBenchmark{},
			[]testing.InternalExample{})
	},
}
