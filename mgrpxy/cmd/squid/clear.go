package squid

import (
	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/shared"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

type squidClearFlags struct {
	Backend string
}

func NewClearCmd(globalFlags *types.GlobalFlags) *cobra.Command {
	var clearCmd = &cobra.Command{
		Use:   "clear",
		Short: "Clear the Squid cache",
		Long:  "Clear the Squid cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			var flags squidClearFlags
			return utils.CommandHelper(globalFlags, cmd, args, &flags, clear)
		},
	}

	return clearCmd
}

func clear(globalFlags *types.GlobalFlags, flags *squidClearFlags, cmd *cobra.Command, args []string) error {
	fn, err := shared.ChooseProxyPodmanOrKubernetes(cmd.Flags(), podmanSquidClear, kubernetesSquidClear)
	if err != nil {
		return err
	}

	return fn(globalFlags, flags, cmd, args)
}
