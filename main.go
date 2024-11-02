package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nowshad-hossain-rahat/k8s-deployer/constants"
	"github.com/nowshad-hossain-rahat/k8s-deployer/types"
	"github.com/nowshad-hossain-rahat/k8s-deployer/utils"
)

func main() {
	cwd, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	cfg, err := utils.ParseConfig(cwd)

	if err != nil {
		fmt.Println(err.Error())

		os.Exit(1)
	}

	runK8sDeployer(cwd, cfg)
}

func runK8sDeployer(cwd string, cfg *types.K8sDeployerConfig) {
	// Define command line flags
	var mode, serviceType, serviceName string
	var operation string

	flag.StringVar(&mode, "mode", "dev", "Set the mode (dev/prod)")
	flag.StringVar(&serviceType, "type", constants.Go, "Set the service type (go/dotnet)")
	flag.StringVar(&serviceName, "svc", "", "Set the service name from the list you've configured in the `"+constants.ConfigFileName+"` file")

	// Parse command line flags
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("[!] Not enough arguments provided. Usage: k8s-deployer <operation>")
		os.Exit(1)
	}

	// operation: build | deploy | bnd (build-and-deploy)
	operation = flag.Args()[0]

	if operation != "build" && operation != "deploy" && operation != "bnd" {
		fmt.Println("[!] Unknown operation: " + operation)

		os.Exit(1)
	}

	// Check if service name is provided and exists
	if serviceName == "" {
		fmt.Println("[!] No service name provided.")
		os.Exit(1)
	}

	switch operation {
	case "build":
		_, err := utils.Build(cfg, cwd, mode, serviceType, serviceName)

		if err != nil {
			fmt.Println(err.Error())

			os.Exit(1)
		}
	case "deploy":
		if err := utils.DeployAlone(cfg, cwd, mode, serviceType, serviceName); err != nil {
			os.Exit(1)
		}
	case "bnd":
		buildInfo, err := utils.Build(cfg, cwd, mode, serviceType, serviceName)

		if err != nil {
			os.Exit(1)
		}

		if err := utils.DeployAfterBuild(cfg, buildInfo, mode, serviceName); err != nil {
			os.Exit(1)
		}
	default:
		fmt.Printf("[!] Unknown operation: %s\n", operation)
		os.Exit(1)
	}

	fmt.Printf(
		"[+] '%s' operation completed successfully for %s.\n",
		operation,
		utils.ParseServiceName(cfg.DockerImagePrefix, serviceName),
	)
}
