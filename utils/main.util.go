package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/nowshad-hossain-rahat/k8s-deployer/constants"
	"github.com/nowshad-hossain-rahat/k8s-deployer/types"
)

func ParseConfig(cwd string) (*types.K8sDeployerConfig, error) {
	configFilePath := path.Join(cwd, "/", constants.ConfigFileName)

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("`%s` file not found", constants.ConfigFileName)
	}

	configJsonBytes, readErr := os.ReadFile(configFilePath)

	if readErr != nil {
		return nil, fmt.Errorf("failed to read the `%s`: %s", constants.ConfigFileName, readErr.Error())
	}

	var config types.K8sDeployerConfig
	err := json.Unmarshal(configJsonBytes, &config)

	if err != nil {
		return nil, fmt.Errorf("failed to parse the %s: %s", constants.ConfigFileName, err.Error())
	}

	return &config, nil
}

func GetServiceDirectoryRoot(cfg *types.K8sDeployerConfig, cwd, serviceType, serviceName string) string {
	serviceDirectoryRoot := ""

	if serviceType == constants.Go {
		exists := cfg.ServicesDirectory.All.Go[serviceName] != ""

		if !exists {
			fmt.Printf("[!] Service %s not found in the configured paths.\n", serviceName)
			os.Exit(1)
		}

		serviceDirectoryRoot = path.Join(
			cfg.ServicesDirectory.Root.Go,
			cfg.ServicesDirectory.All.Go[serviceName],
		)
	} else if serviceType == constants.Dotnet {
		exists := cfg.ServicesDirectory.All.Dotnet[serviceName] != ""

		if !exists {
			fmt.Printf("[!] Service %s not found in the configured paths.\n", serviceName)
			os.Exit(1)
		}

		serviceDirectoryRoot = path.Join(
			cfg.ServicesDirectory.Root.Dotnet,
			cfg.ServicesDirectory.All.Dotnet[serviceName],
		)
	} else {
		fmt.Printf("[!] Unknown service type: %s\n", serviceType)
		os.Exit(1)
	}

	info, err := os.Stat(serviceDirectoryRoot)
	if os.IsNotExist(err) {
		fmt.Printf("[!] Service directory not found: %s\n", serviceDirectoryRoot)
		os.Exit(1)
	} else if !info.IsDir() {
		fmt.Printf("[!] Service directory is not a directory: %s\n", serviceDirectoryRoot)
		os.Exit(1)
	}

	return serviceDirectoryRoot
}
