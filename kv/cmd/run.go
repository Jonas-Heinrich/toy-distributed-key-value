package cmd

import (
	"fmt"
	"net"
	"os"

	"github.com/Jonas-Heinrich/toy-distributed-key-value/kv"
	"github.com/spf13/cobra"
)

var leader bool
var networkMemberAddress string

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().BoolVarP(&leader, "leader", "l", false, "leader")
	runCmd.PersistentFlags().StringVarP(&networkMemberAddress, "networkMemberAddress", "a", "", "IP Address of the initial network member node")
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the kv store",
	Long:  `Run the kv store`,
	Run: func(cmd *cobra.Command, args []string) {
		var nodeAddress net.IP
		if release {
			nodeAddress = net.ParseIP(networkMemberAddress)
			if nodeAddress == nil {
				fmt.Println(fmt.Errorf("an initial network member ip address has to be provided in release mode"))
				os.Exit(1)
			}
		} else {
			// Ignore leader ip address, should be commanded by test
			nodeAddress = nil
		}

		keyValueStore := kv.InitKeyValueStore(leader, nodeAddress)
		keyValueStore.Serve(release)
	},
}
