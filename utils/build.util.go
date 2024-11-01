package utils

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/nowshad-hossain-rahat/k8s-deployer/constants"
	"github.com/nowshad-hossain-rahat/k8s-deployer/types"
)

func Build(cfg *types.K8sDeployerConfig, cwd, mode, serviceType, serviceName string) error {
	fmt.Println("[+] Build process started...")

	serviceDirectoryRoot := GetServiceDirectoryRoot(cfg, cwd, serviceType, serviceName)

	output, err := buildMicroserviceBinary(cfg, serviceDirectoryRoot, mode, serviceType, serviceName)

	if err != nil {
		return err
	}

	fmt.Println("[+] Build process completed...")
	fmt.Println(output)

	deploymentYamlPath, serviceYamlPath := GetDeploymentAndServiceYamlPaths(cfg, mode, serviceType, serviceName)

	fmt.Printf("[+] Parsing deployment YAML file: %s\n", deploymentYamlPath)

	deployment, err := ParseYaml(deploymentYamlPath)

	if err != nil {
		return err
	} else if deployment == nil {
		return fmt.Errorf("[!] Failed to parse deployment YAML file")
	}

	if len(deployment.Spec.Template.Spec.Containers) == 0 {
		return fmt.Errorf("[!] Failed to find container in deployment YAML file")
	}

	dockerImage := deployment.Spec.Template.Spec.Containers[0].Image

	fmt.Println("[+] Extracting current version of the Docker image and generating the next verison...")
	currentVersion, nextVersion := ParseVersion(dockerImage)

	fmt.Printf("[+] Updating deployment YAML file: %s\n", deploymentYamlPath)
	if err := UpdateYaml(deploymentYamlPath, deployment); err != nil {
		return err
	}

	return nil
}

func buildMicroserviceBinary(
	cfg *types.K8sDeployerConfig,
	cwd, mode, serviceType, serviceName string,
) (string, error) {
	var output bytes.Buffer
	var cmd *exec.Cmd
	cmd.Dir = cwd
	cmd.Stdout = &output

	if serviceType == constants.Go {

		fmt.Printf("[+] Building Go binary for the %s-%s-service...\n", cfg.DockerImagePrefix, serviceName)

		cmd.Env = append(cmd.Env, "GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0")

		cmd = exec.Command("go", "build", "-o", fmt.Sprintf("./build/%s_%s", cfg.DockerImagePrefix, serviceName))

	} else if serviceType == constants.Dotnet {

		fmt.Printf("[+] Building .NET binary for the %s-%s-service...\n", cfg.DockerImagePrefix, serviceName)

		cmd = exec.Command("dotnet", "publish", "-c", "Release", "-o", "./build")

	} else {
		return "", fmt.Errorf("[!] Unknown service type: %s", serviceType)
	}

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("[!] Failed to build service: %v", err)
	}

	return output.String(), nil
}
