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
				{"TestTesting", kvtest.TestTesting},
				{"TestStatus", kvtest.TestStatus},
				{"TestInitialState", kvtest.TestInitialState},

				{"TestDirectNetworkEntry", kvtest.TestDirectNetworkEntry},
				{"TestIndirectNetworkEntry", kvtest.TestIndirectNetworkEntry},

				{"TestLeaderElection", kvtest.TestLeaderElection},

				{"TestInitialDirectRead", kvtest.TestInitialDirectRead},
				{"TestInitialIndirectRead", kvtest.TestInitialIndirectRead},
				{"TestInitialDirectReadNotFound", kvtest.TestInitialDirectReadNotFound},
				{"TestInitialIndirectReadNotFound", kvtest.TestInitialIndirectReadNotFound},
			},
			[]testing.InternalBenchmark{},
			[]testing.InternalExample{})
	},
}
