package utils

import (
	"fmt"
	"os/exec"
)

func Deploy(serviceType string) error {
	fmt.Printf("[+] Deploying service of type: %s...\n", serviceType)

	var cmd *exec.Cmd
	if serviceType == "go" {
		cmd = exec.Command("kubectl", "apply", "-f", "k8s/deployment.dev.yaml")
	} else if serviceType == "dotnet" {
		cmd = exec.Command("kubectl", "apply", "-f", "k8s/deployment.prod.yaml")
	} else {
		return fmt.Errorf("[!] Unknown service type: %s", serviceType)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("[!] Failed to deploy service: %v", err)
	}
	return nil
}
