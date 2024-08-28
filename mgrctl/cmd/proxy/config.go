// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/shared/api"
	"github.com/uyuni-project/uyuni-tools/shared/api/proxy"
	. "github.com/uyuni-project/uyuni-tools/shared/l10n"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

// Flag names by key.
const (
	rootCA          = "rootCA"
	intermediateCAs = "intermediateCAs"
	proxyCrt        = "proxyCrt"
	proxyKey        = "proxyKey"
)

// Set of required fields for validation.
var ProxyCreateConfigRequiredFields = [6]string{
	proxyName,
	server,
	email,
	rootCA,
	proxyCrt,
	proxyKey,
}

// Common flags for proxy create config commands.
type proxyCreateConfigBaseFlags struct {
	ConnectionDetails api.ConnectionDetails `mapstructure:"api"`
	ProxyName         string
	ProxyPort         int
	Server            string
	MaxCache          int
	Email             string
	Output            string
}

// Flags for proxy create config command.
type proxyCreateConfigFlags struct {
	proxyCreateConfigBaseFlags `mapstructure:",squash"`
	RootCA                     string
	IntermediateCAs            []string
	ProxyCrt                   string
	ProxyKey                   string
}

// CreateCommand entry command for managing cache.
// Setup for subcommand to clear (the cache).
func NewConfigCommand(globalFlags *types.GlobalFlags) *cobra.Command {
	var flags proxyCreateConfigFlags

	createConfigCmd := &cobra.Command{
		Use:   "config",
		Short: L("Create a proxy configuration file"),
		Long:  L("Create a proxy configuration file"),
		Example: `  Create a proxy configuration file providing only required parameters
	
    $ mgrctl proxy create config --proxyName="proxy.example.com" --server="server.example.com" --email="admin@example.com" --rootCA="root_ca.pem" --proxyCrt="proxy_crt.pem" --proxyKey="proxy_key.pem"

  Create a proxy configuration file providing all parameters
	
	$ mgrctl proxy create config --proxyName="proxy.example.com" --server="server.example.com" --email="admin@example.com" --rootCA="root_ca.pem" --proxyCrt="proxy_crt.pem" --proxyKey="proxy_key.pem" --intermediateCAs="intermediateCA_1.pem,intermediateCA_2.pem,intermediateCA_3.pem" -o="proxy-config"
	
  or an alternative format:

	$ mgrctl proxy create config --proxyName="proxy.example.com" --server="server.example.com" --email="admin@example.com" --rootCA="root_ca.pem" --proxyCrt="proxy_crt.pem" --proxyKey="proxy_key.pem" --intermediateCAs "intermediateCA_1.pem" --intermediateCAs "intermediateCA_2.pem" --intermediateCAs "intermediateCA_3.pem" -o="proxy-config"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return utils.CommandHelper(globalFlags, cmd, args, &flags, proxyCreateConfig)
		},
	}

	createConfigCmd.Flags().String(proxyName, "", L("Unique DNS-resolvable FQDN of this proxy. Mandatory."))
	createConfigCmd.Flags().Int(proxyPort, 8022, L("SSH port the proxy listens one."))
	createConfigCmd.Flags().String(server, "", L("FQDN of the server to connect to proxy to. Mandatory."))
	createConfigCmd.Flags().Int(maxCache, 102400, L("Maximum cache size in MB."))
	createConfigCmd.Flags().String(email, "", L("Email of the proxy administrator"))
	createConfigCmd.Flags().StringP(output, "o", "", L("Filename to write the configuration to (without extension)."))

	createConfigCmd.Flags().String(rootCA, "", L("Path to the root CA used to sign the proxy certificate in PEM format. Mandatory."))
	createConfigCmd.Flags().String(proxyCrt, "", L("Path to the proxy certificate in PEM format. Mandatory."))
	createConfigCmd.Flags().String(proxyKey, "", L("Path to the proxy certificate private key in PEM format. Mandatory."))
	createConfigCmd.Flags().StringSliceP(intermediateCAs, "i", []string{}, L("Path to an intermediate CA used to sign the proxy certicate in PEM format. May be provided multiple times or separated by commas."))

	createConfigCmd.AddCommand(NewConfigGenerateCommand(globalFlags))

	utils.ValidateMandatoryFlags(createConfigCmd, ProxyCreateConfigRequiredFields[:])

	api.AddAPIFlags(createConfigCmd)
	return createConfigCmd
}

func proxyCreateConfig(globalFlags *types.GlobalFlags, flags *proxyCreateConfigFlags, cmd *cobra.Command, args []string) error {
	// handle file paths
	rootCA := string(utils.ReadFile(flags.RootCA))
	proxyCrt := string(utils.ReadFile(flags.ProxyCrt))
	proxyKey := string(utils.ReadFile(flags.ProxyKey))

	var intermediateCAs []string
	for _, v := range flags.IntermediateCAs {
		intermediateCAs = append(intermediateCAs, string(utils.ReadFile(v)))
	}

	_, err := proxy.ContainerConfigGenerate(&flags.ConnectionDetails, flags.ProxyName, flags.ProxyPort, flags.Server, flags.MaxCache, flags.Email, rootCA, proxyCrt, proxyKey, intermediateCAs, flags.Output)

	if err != nil {
		return utils.Errorf(err, L("failed to connect to the server"))
	}

	return nil
}
