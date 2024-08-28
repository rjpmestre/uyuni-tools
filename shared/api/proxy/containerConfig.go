// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"errors"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/uyuni-project/uyuni-tools/shared/api"
	. "github.com/uyuni-project/uyuni-tools/shared/l10n"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

const containerConfigEndpoint = "proxy/containerConfig"

// Compute and download the configuration file for proxy containers with generated certificates.
func ContainerConfigGenerate(
	cnxDetails *api.ConnectionDetails,
	proxyName string,
	proxyPort int,
	server string,
	maxCache int,
	email string,
	rootCA string,
	proxyCrt string,
	proxyKey string,
	intermediateCAs []string,
	output string,
) (*[]int8, error) {
	client, err := api.Init(cnxDetails)

	if err == nil {
		err = client.Login()
	}

	if err != nil {
		return nil, utils.Errorf(err, L("failed to connect to the server"))
	}

	data := map[string]interface{}{
		"proxyName":       proxyName,
		"proxyPort":       proxyPort,
		"server":          server,
		"maxCache":        maxCache,
		"email":           email,
		"rootCA":          rootCA,
		"proxyCrt":        proxyCrt,
		"proxyKey":        proxyKey,
		"intermediateCAs": intermediateCAs,
	}

	log.Trace().Msgf("Creating proxy configuration file with generated certificates with data: %v...", data)

	res, err := api.Post[[]int8](client, containerConfigEndpoint, data)
	if err != nil {
		return nil, utils.Errorf(err, L("failed to create proxy configuration file with generated certificates"))
	}
	if !res.Success {
		return nil, errors.New(res.Message)
	}

	// Save the configuration file
	filename := getFilename(output, proxyName)
	if err := saveBinaryData(filename, res.Result); err != nil {
		return nil, utils.Errorf(err, L("Error saving binary data: %v"), err)
	}

	log.Debug().Msgf("Proxy configuration file with generated certificates saved as %s", filename)
	return &res.Result, nil
}

// Compute and download the configuration file for proxy containers.
func ContainerConfig(
	cnxDetails *api.ConnectionDetails,
	proxyName string,
	proxyPort int,
	server string,
	maxCache int,
	email string,
	caCertificate string,
	caKey string,
	caPassword string,
	cnames []string,
	country string,
	state string,
	city string,
	org string,
	orgUnit string,
	sslEmail string,
	output string,
) (*[]int8, error) {
	client, err := api.Init(cnxDetails)

	if err == nil {
		err = client.Login()
	}

	if err != nil {
		return nil, utils.Errorf(err, L("failed to connect to the server"))
	}

	data := map[string]interface{}{
		"proxyName":  proxyName,
		"proxyPort":  proxyPort,
		"server":     server,
		"maxCache":   maxCache,
		"email":      email,
		"caCrt":      caCertificate,
		"caKey":      caKey,
		"caPassword": caPassword,
		"cnames":     cnames,
		"country":    country,
		"state":      state,
		"city":       city,
		"org":        org,
		"orgUnit":    orgUnit,
		"sslEmail":   sslEmail,
	}

	log.Trace().Msgf("Creating proxy configuration file with data: %v...", data)

	res, err := api.Post[[]int8](client, containerConfigEndpoint, data)
	if err != nil {
		return nil, utils.Errorf(err, L("failed to create proxy configuration file"))
	}
	if !res.Success {
		return nil, errors.New(res.Message)
	}

	// Save the configuration file
	filename := getFilename(output, proxyName)
	if err := saveBinaryData(filename, res.Result); err != nil {
		return nil, utils.Errorf(err, L("Error saving binary data: %v"), err)
	}

	log.Debug().Msgf("Proxy configuration file saved as %s", filename)
	return &res.Result, nil
}

// getFilename returns the filename to save the configuration to.
func getFilename(output string, proxyName string) string {
	filename := output
	if filename == "" {
		filename = strings.Split(proxyName, ".")[0] + "-config"
	}
	return filename + ".tar.gz"
}

// SaveBinaryData saves binary data to a file.
func saveBinaryData(filename string, data []int8) error {
	// Need to convert the array of signed ints to unsigned/byte
	byteArray := make([]byte, len(data))
	for i, v := range data {
		byteArray[i] = byte(v)
	}

	file, err := os.Create(filename)
	if err != nil {
		return utils.Errorf(err, L("error creating file: %v"), filename)
	}
	defer file.Close()

	_, err = file.Write(byteArray)
	if err != nil {
		return utils.Errorf(err, L("error writing file: %v"), filename)
	}

	return nil
}
