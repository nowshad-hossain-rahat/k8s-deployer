package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/nowshad-hossain-rahat/k8s-deployer/constants"
	"github.com/nowshad-hossain-rahat/k8s-deployer/types"
)

func deploy(
	cfg *types.K8sDeployerConfig,
	cwd,
	mode,
	serviceName,
	dockerImagePath,
	deploymentFilePath,
	serviceFilePath string,
) error {
	fullServiceName := ParseServiceName(cfg.DockerImagePrefix, serviceName)
	deleteExistingDeployment(fullServiceName)

	fmt.Println("[+] Deployment process started...")

	var cmd *exec.Cmd
	var output bytes.Buffer

	if mode == constants.Dev {
		imagePushOutput, err := loadDockerImageToMinikube(cwd, dockerImagePath)
		if err != nil {
			return err
		}

		fmt.Println(imagePushOutput)
	} else {
		imagePushOutput, err := pushDockerImageToLive(cwd, dockerImagePath)
		if err != nil {
			return err
		}

		fmt.Println(imagePushOutput)
	}

	fmt.Println("[+] Applying deployment YAML file: " + deploymentFilePath)
	cmd = exec.Command("kubectl", "apply", "-f", deploymentFilePath)
	cmd.Dir = cwd
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("[!] Failed to apply deployment YAML file for '%s': %v", fullServiceName, err)
	}

	fmt.Println(output.String())
	output.Reset()

	fmt.Println("[+] Applying service YAML file: " + serviceFilePath)
	cmd = exec.Command("kubectl", "apply", "-f", serviceFilePath)
	cmd.Dir = cwd
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("[!] Failed to apply service YAML file for '%s': %v", fullServiceName, err)
	}

	fmt.Println(output.String())

	fmt.Println("[+] Deployment process completed...")

	return nil
}

func DeployAlone(
	cfg *types.K8sDeployerConfig,
	cwd,
	mode,
	serviceType,
	serviceName string,
) error {
	serviceDirectoryRoot := GetServiceDirectoryRoot(cfg, cwd, serviceType, serviceName)
	deploymentYamlPath, serviceYamlPath := GetDeploymentAndServiceYamlPaths(cfg, serviceDirectoryRoot, mode, serviceType, serviceName)

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

	dockerImagePath := deployment.Spec.Template.Spec.Containers[0].Image

	fmt.Println("[+] Extracting current version of the Docker image and generating the next verison...")
	currentVersion, _ := ParseVersion(dockerImagePath)

	fmt.Printf("[+] Current version: %s\n", currentVersion)

	dockerImagePath = ParseDockerImagePath(cfg, mode, serviceName, currentVersion)

	return deploy(
		cfg,
		serviceDirectoryRoot,
		mode,
		serviceName,
		dockerImagePath,
		deploymentYamlPath,
		serviceYamlPath,
	)
}

func DeployAfterBuild(
	cfg *types.K8sDeployerConfig,
	buildInfo *BuildInfo,
	mode,
	serviceName string,
) error {
	return deploy(
		cfg,
		buildInfo.ServiceDirectoryRoot,
		mode,
		serviceName,
		buildInfo.NewDockerImagePath,
		buildInfo.DeploymentYamlPath,
		buildInfo.ServiceYamlPath,
	)
}

// func removeMinikubeImage(dockerImagePath string) error {
// 	cmd := exec.Command("minikube", "ssh", "--", "docker", "rmi", dockerImagePath, "--force")

// 	if err := cmd.Run(); err != nil {
// 		return fmt.Errorf("[!] Failed to remove existing Docker image from Minikube: %v", err)
// 	}

// 	fmt.Printf("[+] Removed `%s` Docker image from Minikube\n", dockerImagePath)
// 	return nil
// }

func loadDockerImageToMinikube(cwd, dockerImagePath string) (string, error) {
	fmt.Printf("[->] Loading docker image (%s) to minikube...\n", dockerImagePath)

	var output bytes.Buffer
	cmd := exec.Command("minikube", "image", "load", dockerImagePath)

	cmd.Dir = cwd
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("[!] Failed to load Docker image into Minikube: %v", err)
	}

	return output.String(), nil
}

func pushDockerImageToLive(cwd, dockerImagePath string) (string, error) {
	fmt.Printf("[->] Loading docker image (%s) to minikube...\n", dockerImagePath)

	var output bytes.Buffer
	cmd := exec.Command("docker", "push", dockerImagePath)

	cmd.Dir = cwd
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("[!] Failed to push the image into the Docker registry: %v", err)
	}

	return output.String(), nil
}

func deleteExistingDeployment(fullServiceName string) {
	fmt.Println("[+] Deleting existing deployment...")

	cmd := exec.Command("kubectl", "delete", fullServiceName+"-deployment")
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		fmt.Printf("[!] Failed to delete existing deployment or maybe there wasn't any deployment yet: %v\n", err)
		return
	}

	fmt.Println("[+] Deleted existing deployment if any")
}
