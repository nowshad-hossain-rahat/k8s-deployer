package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/nowshad-hossain-rahat/k8s-deployer/constants"
	"github.com/nowshad-hossain-rahat/k8s-deployer/types"
)

type BuildInfo struct {
	ServiceDirectoryRoot string
	DeploymentYamlPath   string
	ServiceYamlPath      string
	NewDockerImagePath   string
	NextVersion          string
}

func Build(
	cfg *types.K8sDeployerConfig,
	cwd, mode, serviceType, serviceName string,
) (*BuildInfo, error) {
	fmt.Println("[+] Build process started...")

	serviceDirectoryRoot := GetServiceDirectoryRoot(cfg, cwd, serviceType, serviceName)

	output, err := buildMicroserviceBinary(cfg, serviceDirectoryRoot, serviceType, serviceName)

	if err != nil {
		return nil, err
	}

	fmt.Println("[+] Build process completed...")
	fmt.Println(output)

	deploymentYamlPath, serviceYamlPath := GetDeploymentAndServiceYamlPaths(cfg, serviceDirectoryRoot, mode, serviceType, serviceName)

	fmt.Printf("[+] Parsing deployment YAML file: %s\n", deploymentYamlPath)

	deployment, err := ParseYaml(deploymentYamlPath)

	if err != nil {
		return nil, err
	} else if deployment == nil {
		return nil, fmt.Errorf("[!] Failed to parse deployment YAML file")
	}

	if len(deployment.Spec.Template.Spec.Containers) == 0 {
		return nil, fmt.Errorf("[!] Failed to find container in deployment YAML file")
	}

	dockerImagePath := deployment.Spec.Template.Spec.Containers[0].Image

	fmt.Println("[+] Extracting current version of the Docker image and generating the next verison...")
	_, nextVersion := ParseVersion(dockerImagePath)

	if mode != "deploy" {
		fmt.Printf("[+] Next version: %s\n", nextVersion)
	}

	fmt.Println("[+] Building docker image...")

	dockerImagePath = ParseDockerImagePath(cfg, mode, serviceName, nextVersion)
	output, err = buildDockerImage(serviceDirectoryRoot, dockerImagePath)

	if err != nil {
		fmt.Println(output)

		return nil, err
	}

	fmt.Println("[+] Building Docker image completed...")
	fmt.Println(output)

	fmt.Printf("[+] Updating deployment YAML file: %s\n", deploymentYamlPath)

	deployment.Spec.Template.Spec.Containers[0].Image = dockerImagePath

	if err := UpdateYaml(deploymentYamlPath, deployment); err != nil {
		return nil, err
	}

	return &BuildInfo{
		ServiceDirectoryRoot: serviceDirectoryRoot,
		DeploymentYamlPath:   deploymentYamlPath,
		ServiceYamlPath:      serviceYamlPath,
		NewDockerImagePath:   dockerImagePath,
		NextVersion:          nextVersion,
	}, nil
}

func buildMicroserviceBinary(
	cfg *types.K8sDeployerConfig,
	cwd, serviceType, serviceName string,
) (string, error) {
	fullServiceName := ParseServiceName(cfg.DockerImagePrefix, serviceName)

	if serviceType == constants.Go {

		fmt.Printf("[+] Building Go binary for the %s...\n", fullServiceName)

		buildFileName := fmt.Sprintf("%s_%s", cfg.DockerImagePrefix, serviceName)

		if cfg.DockerImagePrefix == "" {
			buildFileName = serviceName
		}

		buildOutputPath := path.Join(cwd, fmt.Sprintf("/build/%s", buildFileName))

		os.Remove(buildOutputPath)

		cmd := exec.Command("go", "build", "-o", buildOutputPath)

		cmd.Dir = cwd
		cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0")

		var output, errOutput bytes.Buffer
		cmd.Stdout = &output
		cmd.Stderr = &errOutput

		if err := cmd.Run(); err != nil {
			return errOutput.String(), fmt.Errorf("[!] Failed to build service")
		}

		return output.String(), nil

	} else if serviceType == constants.Dotnet {

		fmt.Printf("[+] Building .NET binary for the '%s'...\n", fullServiceName)

		buildOutputPath := path.Join(cwd, "/build")

		os.Remove(buildOutputPath)

		cmd := exec.Command("dotnet", "publish", "-c", "Release", "-o", buildOutputPath)
		cmd.Dir = cwd
		cmd.Env = os.Environ()

		var output, errOutput bytes.Buffer
		cmd.Stdout = &output
		cmd.Stderr = &errOutput

		if err := cmd.Run(); err != nil {
			return errOutput.String(), fmt.Errorf("[!] Failed to build service: %v", err.Error())
		}

		return output.String(), nil
	} else {
		return "", fmt.Errorf("[!] Unknown service type: %s", serviceType)
	}
}

func buildDockerImage(cwd, dockerImage string) (string, error) {
	cmd := exec.Command("docker", "build", "-t", dockerImage, ".")
	cmd.Dir = cwd

	var output, errOutput bytes.Buffer

	cmd.Stdout = &output
	cmd.Stderr = &errOutput

	if err := cmd.Run(); err != nil {
		return errOutput.String(), fmt.Errorf("[!] Failed to build Docker image: %v", err)
	}

	return output.String(), nil
}
