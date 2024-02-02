// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package podman

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uyuni-project/uyuni-tools/mgradm/cmd/migrate/shared"
	"github.com/uyuni-project/uyuni-tools/mgradm/shared/podman"
	podman_utils "github.com/uyuni-project/uyuni-tools/shared/podman"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

func migrateToPodman(globalFlags *types.GlobalFlags, flags *podmanMigrateFlags, cmd *cobra.Command, args []string) error {
	if _, err := exec.LookPath("podman"); err != nil {
		log.Fatal().Err(err).Msg("install podman before running this command")
	}

	// Find the SSH Socket and paths for the migration
	sshAuthSocket := shared.GetSshAuthSocket()
	sshConfigPath, sshKnownhostsPath := shared.GetSshPaths()

	scriptDir := shared.GenerateMigrationScript(args[0], false)
	defer os.RemoveAll(scriptDir)

	extraArgs := []string{
		"--security-opt", "label:disable",
		"-e", "SSH_AUTH_SOCK",
		"-v", filepath.Dir(sshAuthSocket) + ":" + filepath.Dir(sshAuthSocket),
		"-v", scriptDir + ":/var/lib/uyuni-tools/",
	}

	if sshConfigPath != "" {
		extraArgs = append(extraArgs, "-v", sshConfigPath+":/tmp/ssh_config")
	}

	if sshKnownhostsPath != "" {
		extraArgs = append(extraArgs, "-v", sshKnownhostsPath+":/etc/ssh/ssh_known_hosts")
	}

	serverImage, err := utils.ComputeImage(flags.Image.Name, flags.Image.Tag)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to compute image URL")
	}
	podman_utils.PrepareImage(serverImage, flags.Image.PullPolicy)

	log.Info().Msg("Migrating server")
	runContainer("uyuni-migration", serverImage, extraArgs,
		[]string{"/var/lib/uyuni-tools/migrate.sh"})

	// Read the extracted data
	tz, oldPgVersion, newPgVersion := shared.ReadContainerData(scriptDir)

	if oldPgVersion != newPgVersion {
		var migrationImage types.ImageFlags
		migrationImage.Name = flags.MigrationImage.Name
		if migrationImage.Name == "" {
			migrationImage.Name = fmt.Sprintf("%s-migration-%s-%s", flags.Image.Name, oldPgVersion, newPgVersion)
		}
		migrationImage.Tag = flags.MigrationImage.Tag
		log.Info().Msgf("Using migration image %s:%s", migrationImage.Name, migrationImage.Tag)

		image, err := utils.ComputeImage(migrationImage.Name, migrationImage.Tag)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to compute image URL")
		}
		podman_utils.PrepareImage(image, flags.Image.PullPolicy)
		shared.GeneratePgMigrationScript(scriptDir, oldPgVersion, newPgVersion, false)
		runContainer("uyuni-pg-migration", image, extraArgs,
			[]string{"/var/lib/uyuni-tools/migrate.sh"})
	}

	shared.GenerateFinalizePostgresMigrationScript(scriptDir, true, oldPgVersion != newPgVersion, true, true, false)
	runContainer("uyuni-migration", serverImage, extraArgs,
		[]string{"/var/lib/uyuni-tools/migrate.sh"})

	podman.GenerateSystemdService(tz, serverImage, false, viper.GetStringSlice("podman.arg"))

	// Start the service
	if err := podman_utils.EnableService(podman_utils.ServerService); err != nil {
		return err
	}

	log.Info().Msg("Server migrated")

	podman_utils.EnablePodmanSocket()

	return nil
}

func runContainer(name string, image string, extraArgs []string, cmd []string) {

	podmanArgs := append([]string{"run", "--name", name}, podman.GetCommonParams()...)
	podmanArgs = append(podmanArgs, extraArgs...)

	for volumeName, containerPath := range utils.VOLUMES {
		podmanArgs = append(podmanArgs, "-v", volumeName+":"+containerPath)
	}

	podmanArgs = append(podmanArgs, image)
	podmanArgs = append(podmanArgs, cmd...)

	err := utils.RunCmdStdMapping("podman", podmanArgs...)

	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to run %s container", name)
	}
}
