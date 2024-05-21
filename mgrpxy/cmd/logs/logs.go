package logs

import (
	"errors"
	"os/exec"
	"strings"

	. "github.com/uyuni-project/uyuni-tools/shared/l10n"

	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/shared/podman"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

type logsFlags struct {
	containers []string
	follow     bool
	timestamps bool
	tail       int
	since      string
}

// NewCommand to get the logs of the server.
func NewCommand(globalFlags *types.GlobalFlags) *cobra.Command {
	var flags logsFlags

	cmd := &cobra.Command{
		Use:   "logs [container…]",
		Short: L("Get the proxy logs"),
		Long:  L("Get the proxy logs"),
		RunE: func(cmd *cobra.Command, args []string) error {
			flags.containers = cmd.Flags().Args()
			return utils.CommandHelper(globalFlags, cmd, args, &flags, logs)
		},
		ValidArgsFunction: getContainerNames,
	}

	cmd.Flags().BoolVarP(&flags.follow, "follow", "f", false, L("specify if logs should be followed"))
	cmd.Flags().BoolVarP(&flags.timestamps, "timestamps", "t", false, L("show timestamps in the log outputs"))
	cmd.Flags().IntVar(&flags.tail, "tail", -1, L("number of lines to show from the end of the logs"))
	cmd.Flags().Lookup("tail").NoOptDefVal = "-1"
	cmd.Flags().StringVar(&flags.since, "since", "", L("show logs since a specific time or duration. Supports Go duration strings and RFC3339 format (e.g. 5s, 2m, 3h, 2023-01-02T15:04:05)"))

	cmd.SetUsageTemplate(cmd.UsageTemplate())
	return cmd
}

func logs(globalFlags *types.GlobalFlags, flags *logsFlags, cmd *cobra.Command, args []string) error {
	if podman.HasService(podman.ProxyService) {
		return podmanLogs(globalFlags, flags, cmd, args)
	}

	if utils.IsInstalled("kubectl") && utils.IsInstalled("helm") {
		return kubernetesLogs(globalFlags, flags, cmd, args)
	}

	return errors.New(L("no installed proxy detected"))
}

func getContainerNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var getContainerNamesCmd *exec.Cmd

	if podman.HasService(podman.ProxyService) {
		getContainerNamesCmd = exec.Command("podman", "ps", "--format", "{{.Names}}")
	} else if utils.IsInstalled("kubectl") && utils.IsInstalled("helm") {
		getContainerNamesCmd = exec.Command("kubectl", "get", "pods", "--no-headers", "-o", "custom-columns=:.metadata.name")
	}

	out, err := getContainerNamesCmd.Output()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	names := strings.Split(strings.TrimSpace(string(out)), "\n")
	return names, cobra.ShellCompDirectiveNoFileComp
}
