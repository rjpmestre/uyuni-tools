// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/shared/api"
	"github.com/uyuni-project/uyuni-tools/shared/api/proxy"
	. "github.com/uyuni-project/uyuni-tools/shared/l10n"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

// Flag names by key.
const (
	proxyName  = "proxyName"
	proxyPort  = "proxyPort"
	server     = "server"
	maxCache   = "maxCache"
	email      = "email"
	caCrt      = "caCrt"
	caKey      = "caKey"
	caPassword = "caPassword"
	output     = "output"
	cNames     = "cnames"
	country    = "country"
	state      = "state"
	city       = "city"
	org        = "org"
	orgUnit    = "orgUnit"
	sslEmail   = "sslEmail"
)

// Set of required fields for validation.
var ProxyCreateConfigGenerateRequiredFields = [5]string{
	proxyName,
	server,
	email,
	caCrt,
	caKey,
}

// Flags for proxy create config with generated certificates command.
type proxyCreateConfigGenerateFlags struct {
	proxyCreateConfigBaseFlags `mapstructure:",squash"`
	CaCrt                      string
	CaKey                      string
	CaPassword                 string
	CNames                     []string
	Country                    string
	State                      string
	City                       string
	Org                        string
	OrgUnit                    string
	SslEmail                   string
}

// createCmd entry command for managing cache.
// Setup for subcommand to clear (the cache).
func NewConfigGenerateCommand(globalFlags *types.GlobalFlags) *cobra.Command {
	var flags proxyCreateConfigGenerateFlags

	createConfigCmd := &cobra.Command{
		Use:   "generate",
		Short: L("Create a proxy configuration file with generated certificates"),
		Long:  L("Create a proxy configuration file with existing proxy certificate and key"),
		Example: `  Create a proxy configuration file providing only required parameters
	
	$ mgrctl proxy create config generate --proxyName="proxy.example.com" --server="server.example.com" --email="admin@org.com" --caCrt="ca.pem" --caKey="caKey.pem"
	
  Create a proxy configuration file providing only required parameters and ca password
	
	$ mgrctl proxy create config generate --proxyName="proxy.example.com" --server="server.example.com" --email="admin@org.com" --caCrt="ca.pem" --caKey="caKey.pem" --caPassword="pass.txt"

  Create a proxy configuration file providing all parameters
	
	$ mgrctl proxy create config generate --proxyName="proxy.example.com" --server="server.example.com" --email="admin@org.com" --caCrt="ca.pem" --caKey="caKey.pem" --caPassword="pass.txt" --cnames="proxy_a.example.com,proxy_b.example.com,proxy_c.example.com" --country="PT" --state="Lisbon" --city="Parede" --org="orgExample" --orgUnit="orgUnitExample" --sslEmail="sslEmail@example.com" -o="proxy-config"

  or an alternative format:
	
	$ mgrctl proxy create config generate --proxyName="proxy.example.com" --server="server.example.com" --email="admin@org.com" --caCrt="ca.pem" --caKey="caKey.pem" --caPassword="pass.txt" --cnames="proxy_a.example.com" --cnames="proxy_b.example.com" --cnames="proxy_c.example.com" --country="PT" --state="Lisbon" --city="Parede" --org="orgExample" --orgUnit="orgUnitExample" --sslEmail="sslEmail@example.com" -o="proxy-config"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return utils.CommandHelper(globalFlags, cmd, args, &flags, proxyCreateConfigGenerate)
		},
	}

	createConfigCmd.Flags().String(proxyName, "", L("Unique DNS-resolvable FQDN of this proxy. Mandatory."))
	createConfigCmd.Flags().Int(proxyPort, 8022, L("SSH port the proxy listens one."))
	createConfigCmd.Flags().String(server, "", L("FQDN of the server to connect to proxy to. Mandatory."))
	createConfigCmd.Flags().Int(maxCache, 102400, L("Maximum cache size in MB."))
	createConfigCmd.Flags().String(email, "", L("Email of the proxy administrator"))
	createConfigCmd.Flags().StringP(output, "o", "", L("Filename to write the configuration to (without extension)."))

	createConfigCmd.Flags().String(caCrt, "", L("Path to the certificate of the CA to use to generate a new proxy certificate. Mandatory."))
	createConfigCmd.Flags().String(caKey, "", L("Path to the private key of the CA to use to generate a new proxy certificate. Mandatory."))
	createConfigCmd.Flags().String(caPassword, "", L("Path to a file containing the password of the CA private key, will be prompted if not passed. Mandatory."))
	createConfigCmd.Flags().StringSlice(cNames, []string{}, L("Alternate name of the proxy to set in the certificate. May be provided multiple times or separated by commas."))
	createConfigCmd.Flags().String(country, "", L("Country code to set in the certificate."))
	createConfigCmd.Flags().String(state, "", L("State name to set in the certificate."))
	createConfigCmd.Flags().String(city, "", L("City name to set in the certificate."))
	createConfigCmd.Flags().String(org, "", L("Organization name to set in the certificate."))
	createConfigCmd.Flags().String(orgUnit, "", L("Organization unit name to set in the certificate."))
	createConfigCmd.Flags().String(sslEmail, "", L("Email to set in the certificate."))

	utils.ValidateMandatoryFlags(createConfigCmd, ProxyCreateConfigGenerateRequiredFields[:])

	api.AddAPIFlags(createConfigCmd)
	return createConfigCmd
}

func proxyCreateConfigGenerate(globalFlags *types.GlobalFlags, flags *proxyCreateConfigGenerateFlags, cmd *cobra.Command, args []string) error {
	// handle caPassword
	var caPasswordRead string
	if flags.CaPassword == "" {
		caPasswordRead = strings.TrimSpace(promptForPassword())
		if caPasswordRead == "" {
			return fmt.Errorf(L("%s is required"), caPassword)
		}
	} else {
		caPasswordRead = string(utils.ReadFile(flags.CaPassword))
	}

	// handle other file paths
	caCrt := string(utils.ReadFile(flags.CaCrt))
	caKey := string(utils.ReadFile(flags.CaKey))

	_, err := proxy.ContainerConfig(&flags.ConnectionDetails,
		flags.ProxyName, flags.ProxyPort, flags.Server, flags.MaxCache, flags.Email, caCrt, caKey, caPasswordRead,
		flags.CNames, flags.Country, flags.State, flags.City, flags.Org, flags.OrgUnit, flags.SslEmail,
		flags.Output)

	if err != nil {
		return utils.Errorf(err, L("failed to connect to the server"))
	}

	return nil
}

// Prompt for password.
func promptForPassword() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Please enter %s: ", caPassword)
	password, _ := reader.ReadString('\n')
	return password
}
