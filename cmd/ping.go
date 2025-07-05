package cmd

import (
	"fmt"
	"sync"
	"github.com/settlectl/settle-core/common"
	"github.com/settlectl/settle-core/inventory/parser"
	"github.com/settlectl/settle-core/inventory/ssh"
	"github.com/spf13/cobra"
)

var (
	filterHost string
	filterGroup string
)


var pingCmd = &cobra.Command{
	Use: "ping",
	Short: "Check ssh connectivity to hosts",
	Run: func(cmd *cobra.Command, args []string) {
		hosts, err := parser.ParseHosts("hosts.stl")
		if err != nil {
			fmt.Printf("Error parsing hosts file: %v\n", err)
			return
		}

		if len(hosts) == 0 {
			fmt.Println("No hosts found")
			return
		}

		successCount := 0
		failureCount := 0

		var wg sync.WaitGroup


		for _, host := range hosts {
			if filterHost != "" && host.Name != filterHost {
				continue
			}

			if filterGroup != "" && host.Group != filterGroup {
				continue
			}

			wg.Add(1)
			go func(h *common.Host) {
				defer wg.Done()
				err := ssh.PingHost(h)
				if err != nil {
					fmt.Printf("Failed to ping %s: %v\n", host.Name, err)
					failureCount++
				} else {
					fmt.Printf("Successfully pinged %s\n", host.Name)
					successCount++
				}
			}(&host)
		}
		wg.Wait()

		fmt.Printf("Ping results:\n")
		fmt.Printf("Success: %d\n", successCount)
		fmt.Printf("Failure: %d\n", failureCount)
	},
}

func init() {

	pingCmd.Flags().StringVarP(&filterHost, "host", "H", "", "Filter hosts by name")
	pingCmd.Flags().StringVarP(&filterGroup, "group", "G", "", "Filter hosts by group")
	rootCmd.AddCommand(pingCmd)
}