package main

import (
	"flag"
	"fmt"
	"os"

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

	flag.StringVar(&mode, "mode", "", "Set the mode (dev/prod)")
	flag.StringVar(&serviceType, "type", "", "Set the service type (go/dotnet)")
	flag.StringVar(&serviceName, "service", "", "Set the service name")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("[!] Not enough arguments provided. Usage: go run main.go <operation>")
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
		if err := utils.Build(cfg, cwd, mode, serviceType, serviceName); err != nil {
			os.Exit(1)
		}
	case "deploy":
		if err := utils.Deploy(serviceType); err != nil {
			os.Exit(1)
		}
	case "bnd":
		if err := utils.Build(cfg, cwd, mode, serviceType, serviceName); err != nil {
			os.Exit(1)
		}
		if err := utils.Deploy(serviceType); err != nil {
			os.Exit(1)
		}
	default:
		fmt.Printf("[!] Unknown operation: %s\n", operation)
		os.Exit(1)
	}

	fmt.Printf("[+] %s operation completed successfully for %s.\n", operation, serviceName)
}
