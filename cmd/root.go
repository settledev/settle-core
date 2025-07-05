package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "settlectl",
	Short: "Settle — agentless, stateful configuration automation",
	Long: `Settle is a modern, minimal, state-aware configuration tool built for 
infrastructure and platform engineers who want the power of Ansible 
and the safety of Terraform — without the overhead.

🔧 Agentless. No daemons or remote installations. Settle uses native SSH 
   to connect to your servers, containers, or virtual machines.

📄 Declarative. Define your desired configuration using a simple, 
   readable DSL. Settle parses these .stl files to create, refresh, 
   plan, and apply state changes.

💡 Safe-by-default. With built-in lifecycle awareness and a local state 
   file, Settle helps you avoid unintended changes and drift.

⛏️ Composable. Resources like package, service, and file are 
   first-class primitives that you can extend or override.

🔥 Designed to be small, understandable, and fast to adopt.

Get started with:
  settlectl ping             # check SSH connectivity
  settlectl plan             # see what would change
  settlectl apply            # apply changes from your config
  settlectl drop             # safely remove config and reverse state
  settlectl refresh          # pull real state into local snapshot

Settle is early but growing fast. Open source. Built in Go. Made for you.`,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}