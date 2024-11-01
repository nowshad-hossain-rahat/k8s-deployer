package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/nowshad-hossain-rahat/k8s-deployer/constants"
	"github.com/nowshad-hossain-rahat/k8s-deployer/types"
	"gopkg.in/yaml.v3"
)

func GetDeploymentAndServiceYamlPaths(
	cfg *types.K8sDeployerConfig,
	serviceDirectoryRoot string,
	mode, serviceType, serviceName string,
) (string, string) {
	var deploymentYamlPath, serviceYamlPath string

	if mode == constants.Dev {
		if serviceType == constants.Go {
			deploymentYamlPath = path.Join(
				serviceDirectoryRoot,
				cfg.KubernetesConfig.Directory.Go,
				cfg.KubernetesConfig.Files.Dev.Deployment,
			)

			serviceYamlPath = path.Join(
				serviceDirectoryRoot,
				cfg.KubernetesConfig.Directory.Go,
				cfg.KubernetesConfig.Files.Dev.Service,
			)
		} else {
			deploymentYamlPath = path.Join(
				serviceDirectoryRoot,
				cfg.KubernetesConfig.Directory.Dotnet,
				cfg.KubernetesConfig.Files.Dev.Deployment,
			)

			serviceYamlPath = path.Join(
				serviceDirectoryRoot,
				cfg.KubernetesConfig.Directory.Dotnet,
				cfg.KubernetesConfig.Files.Dev.Service,
			)
		}
	} else {
		if serviceType == constants.Go {
			deploymentYamlPath = path.Join(
				serviceDirectoryRoot,
				cfg.KubernetesConfig.Directory.Go,
				cfg.KubernetesConfig.Files.Prod.Deployment,
			)

			serviceYamlPath = path.Join(
				serviceDirectoryRoot,
				cfg.KubernetesConfig.Directory.Go,
				cfg.KubernetesConfig.Files.Prod.Service,
			)
		} else {
			deploymentYamlPath = path.Join(
				serviceDirectoryRoot,
				cfg.KubernetesConfig.Directory.Dotnet,
				cfg.KubernetesConfig.Files.Prod.Deployment,
			)

			serviceYamlPath = path.Join(
				serviceDirectoryRoot,
				cfg.KubernetesConfig.Directory.Dotnet,
				cfg.KubernetesConfig.Files.Prod.Service,
			)
		}
	}

	return deploymentYamlPath, serviceYamlPath
}

func ParseYaml(yamlPath string) (*types.Deployment, error) {
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("[!] Error reading YAML file: %v\n", err)
	}

	var deployment types.Deployment
	err = yaml.Unmarshal(data, &deployment)
	if err != nil {
		return nil, fmt.Errorf("[!] Error parsing YAML file: %v\n", err)
	}

	return &deployment, nil
}

func UpdateYaml(yamlPath string, deployment *types.Deployment) error {
	// Marshal modified struct back to YAML
	modifiedData, err := yaml.Marshal(deployment)

	if err != nil {
		return err
	}

	// Write back to the file
	err = os.WriteFile(yamlPath, modifiedData, 0644)
	if err != nil {
		return err
	}

	return nil
}

// ParseVersion takes a Docker image string and returns the current version and the next version to be used.
// If the image does not contain a version, it uses the default version 1.0.0.
// It increments the patch version by 1 and ensures that the version is valid.
// If the patch version is over 99, it resets the patch version to 0 and increments the minor version by 1.
// If the minor version is over 99, it resets the minor version to 0 and increments the major version by 1.
func ParseVersion(dockerImage string) (string, string) {
	versionStr := dockerImage[strings.LastIndex(dockerImage, ":")+1:]

	if versionStr == "" {
		versionStr = "1.0.0"

		fmt.Println("Image: ", dockerImage)
		fmt.Println("No version was set in YAML for the Docker image. Using default version: 1.0.0")
	} else {
		fmt.Println("Image: ", dockerImage)
		fmt.Println("Current Version: ", versionStr)
	}

	majorStr := versionStr[:1]
	minorStr := versionStr[2:3]
	patchStr := versionStr[4:]

	if majorStr != "" || minorStr != "" || patchStr != "" {
		majorStr = "1"
		minorStr = "0"
		patchStr = "0"
	}

	major, _ := strconv.Atoi(majorStr)
	minor, _ := strconv.Atoi(minorStr)
	patch, _ := strconv.Atoi(patchStr)

	if major < 0 {
		major = 0
	}

	if minor < 0 {
		minor = 0
	}

	if patch < 0 {
		patch = 0
	}

	patch = patch + 1

	if patch > 99 {
		patch = 0
		minor = minor + 1

		if minor > 99 {
			minor = 0
			major = major + 1
		}
	}

	newVersion := fmt.Sprintf("%d.%d.%d", major, minor, patch)

	return versionStr, newVersion
}

func BuildAndDeploy() {
	if len(os.Args) < 3 {
		fmt.Println("[!] Not enough arguments provided. Usage: go run main.go <SERVICE_NAME> <DOCKER_TAG> [IMAGE_VERSION]")
		os.Exit(1)
	}

	serviceName := os.Args[1]
	dockerTag := os.Args[2]
	imageVersion := "latest"
	if len(os.Args) > 3 {
		imageVersion = os.Args[3]
	}

	projectRoot, err := os.Getwd()
	if err != nil {
		fmt.Printf("[!] Failed to get project root: %v\n", err)
		os.Exit(1)
	}

	if err := cleanDockerTagFile(dockerTag); err != nil {
		os.Exit(1)
	}

	if err := buildBinary(serviceName, dockerTag); err != nil {
		os.Exit(1)
	}

	if err := buildDockerImage(dockerTag, imageVersion); err != nil {
		os.Exit(1)
	}

	if err := removeMinikubeImage(dockerTag, imageVersion); err != nil {
		os.Exit(1)
	}

	if err := loadImageToMinikube(dockerTag, imageVersion); err != nil {
		os.Exit(1)
	}

	if err := deleteK8sPod(serviceName); err != nil {
		os.Exit(1)
	}

	if err := deployToKubernetes(projectRoot, serviceName); err != nil {
		os.Exit(1)
	}

	fmt.Printf("[+] %s-service deployed successfully.\n", serviceName)
}

func cleanDockerTagFile(dockerTag string) error {
	if _, err := os.Stat(dockerTag); err == nil {
		fmt.Printf("[+] Found existing file %s. Deleting it...\n", dockerTag)
		if err := os.Remove(dockerTag); err != nil {
			return fmt.Errorf("[!] Failed to delete old file %s: %v", dockerTag, err)
		}
		fmt.Printf("[+] Deleted old file %s.\n", dockerTag)
	} else {
		fmt.Printf("[+] No previous file %s found.\n", dockerTag)
	}
	return nil
}

func buildBinary(serviceName, dockerTag string) error {
	fmt.Printf("[+] Building %s-service binary...\n", serviceName)
	cmd := exec.Command("go", "build", "-o", "./build/"+dockerTag)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("[!] Failed to build %s binary: %v", dockerTag, err)
	}
	return nil
}

func buildDockerImage(dockerTag, imageVersion string) error {
	fmt.Println("[+] Building docker image...")
	cmd := exec.Command("docker", "build", "-t", fmt.Sprintf("%s:%s", dockerTag, imageVersion), ".")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("[!] Failed to build Docker image for %s: %v", dockerTag, err)
	}
	return nil
}

func removeMinikubeImage(dockerTag, imageVersion string) error {
	cmd := exec.Command("minikube", "ssh", "--", "docker", "rmi", fmt.Sprintf("%s:%s", dockerTag, imageVersion), "--force")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("[!] Failed to remove existing Docker image from Minikube: %v", err)
	}
	return nil
}

func loadImageToMinikube(dockerTag, imageVersion string) error {
	fmt.Printf("[->] Loading docker image of %s-service to minikube...\n", dockerTag)
	cmd := exec.Command("minikube", "image", "load", fmt.Sprintf("%s:%s", dockerTag, imageVersion))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("[!] Failed to load Docker image into Minikube: %v", err)
	}
	return nil
}

func deleteK8sPod(serviceName string) error {
	cmd := exec.Command("kubectl", "delete", "pod", "-l", fmt.Sprintf("app=%s-service", serviceName))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("[!] Failed to delete existing pod: %v", err)
	}
	return nil
}

func deployToKubernetes(projectRoot, serviceName string) error {
	fmt.Printf("[+] Deploying %s-service to Kubernetes...\n", serviceName)

	deploymentFile := fmt.Sprintf("%s/k8s/deployment.dev.yaml", projectRoot)
	serviceFile := fmt.Sprintf("%s/k8s/service.yaml", projectRoot)

	if err := checkK8sManifestExists(deploymentFile, serviceFile); err != nil {
		return err
	}

	if err := applyK8sManifest(deploymentFile); err != nil {
		return err
	}

	if err := applyK8sManifest(serviceFile); err != nil {
		return err
	}

	return nil
}

func checkK8sManifestExists(deploymentFile, serviceFile string) error {
	if _, err := os.Stat(deploymentFile); os.IsNotExist(err) {
		return fmt.Errorf("[!] Kubernetes manifest for %s-service not found", deploymentFile)
	}
	if _, err := os.Stat(serviceFile); os.IsNotExist(err) {
		return fmt.Errorf("[!] Kubernetes service file for %s-service not found", serviceFile)
	}
	return nil
}

func applyK8sManifest(file string) error {
	cmd := exec.Command("kubectl", "apply", "-f", file)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("[!] Failed to apply manifest %s: %v", file, err)
	}
	return nil
}
