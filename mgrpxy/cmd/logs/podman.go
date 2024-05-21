// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	. "github.com/uyuni-project/uyuni-tools/shared/l10n"
	"github.com/uyuni-project/uyuni-tools/shared/podman"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

func podmanLogs(
	globalFlags *types.GlobalFlags,
	flags *logsFlags,
	cmd *cobra.Command,
	args []string,
) error {
	commandArgs := []string{"logs"}
	if flags.follow {
		commandArgs = append(commandArgs, "-f")
	}

	if flags.tail != -1 {
		commandArgs = append(commandArgs, "--tail="+fmt.Sprintf("%d", flags.tail))
	}

	if flags.timestamps {
		commandArgs = append(commandArgs, "--timestamps")
	}

	if flags.since != "" {
		commandArgs = append(commandArgs, fmt.Sprintf("--since=%s", flags.since))
	}

	if len(flags.containers) == 0 {
		commandArgs = append(commandArgs, podman.ProxyContainerNames...)
	} else {
		commandArgs = append(commandArgs, args...)
	}

	log.Trace().Msgf("Running: %s %s", "podman", strings.Join(commandArgs, " "))

	err := utils.RunCmdStdMapping(zerolog.DebugLevel, "podman", commandArgs...)
	if err != nil {
		return fmt.Errorf(L("failed running podman logs %s: %s"), commandArgs, err)
	}

	return nil
}
