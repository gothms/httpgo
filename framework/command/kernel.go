package command

import "github.com/gothms/httpgo/framework/cobra"

// AddKernelCommands will add all command/ * to root command
func AddKernelCommands(root *cobra.Command) {
	root.AddCommand(DemoCommand)
	root.AddCommand(initAppCommand())
}
