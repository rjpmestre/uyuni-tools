// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package status

import (
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/shared/kubernetes"
	. "github.com/uyuni-project/uyuni-tools/shared/l10n"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

func kubernetesStatus(
	globalFlags *types.GlobalFlags,
	flags *statusFlags,
	cmd *cobra.Command,
	args []string,
) error {
	namespace, err := kubernetes.GetNamespace("uyuni-proxy")
	if err != nil {
		return utils.Errorf(err, L("failed to find the uyuni-proxy deployment namespace"))
	}

	// Is the pod running? Do we have all the replicas?
	status, err := kubernetes.GetDeploymentStatus(namespace, "uyuni-proxy")
	if err != nil {
		return utils.Errorf(err, L("failed to get deployment status"))
	}
	if status.Replicas != status.ReadyReplicas {
		log.Warn().Msgf(L("Some replicas are not ready: %[1]d / %[2]d"), status.ReadyReplicas, status.Replicas)
	}

	if status.AvailableReplicas == 0 {
		return errors.New(L("the pod is not running"))
	}

	log.Info().Msg(L("Proxy containers up and running"))

	return nil
}
