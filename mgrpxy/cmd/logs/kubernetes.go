// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	. "github.com/uyuni-project/uyuni-tools/shared/l10n"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

func kubernetesLogs(
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
		if isRFC3339(flags.since) {
			commandArgs = append(commandArgs, fmt.Sprintf("--since-time=%s", flags.since))
		} else {
			commandArgs = append(commandArgs, fmt.Sprintf("--since=%s", flags.since))
		}
	}

	if len(flags.containers) == 0 {
		commandArgs = append(commandArgs, "--all-containers=true")
	} else {
		commandArgs = append(commandArgs, args...)
	}

	log.Trace().Msgf("Running: %s %s", "kubectl", strings.Join(commandArgs, " "))

	err := utils.RunCmdStdMapping(zerolog.DebugLevel, "kubectl", commandArgs...)
	if err != nil {
		return fmt.Errorf(L("failed running kubectl logs %s: %s"), commandArgs, err)
	}

	return nil
}

func isRFC3339(timestamp string) bool {
	_, err := time.Parse(time.RFC3339, timestamp)
	return err == nil
}
