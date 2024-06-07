package squid

import (
	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/shared/types"
)

func NewCommand(globalFlags *types.GlobalFlags) *cobra.Command {
	var squidCmd = &cobra.Command{
		Use:   "squid",
		Short: "Manage Squid cache",
		Long:  `Manage Squid cache`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	squidCmd.AddCommand(NewClearCmd(globalFlags))
	return squidCmd
}
