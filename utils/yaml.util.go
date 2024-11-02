package utils

import (
	"fmt"
	"os"
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
		return nil, fmt.Errorf("error reading YAML file: %v", err)
	}

	var deployment types.Deployment
	err = yaml.Unmarshal(data, &deployment)
	if err != nil {
		return nil, fmt.Errorf("error parsing YAML file: %v", err)
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

		fmt.Println("[+] Image: ", dockerImage)
		fmt.Println("[!] No version was set in YAML for the Docker image. Using default version: 1.0.0")
	} else {
		fmt.Println("[+] Image: ", dockerImage)
		fmt.Println("[+] Current Version: ", versionStr)
	}

	majorStr := versionStr[:1]
	minorStr := versionStr[2:3]
	patchStr := versionStr[4:]

	if majorStr == "" || minorStr == "" || patchStr == "" {
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

func ParseDockerImagePath(cfg *types.K8sDeployerConfig, mode, serviceName, version string) string {
	containerRegistry := cfg.DockerContainerRegistry.Dev

	if mode == constants.Prod {
		containerRegistry = cfg.DockerContainerRegistry.Prod
	}

	dockerTag := fmt.Sprintf("%s_%s", cfg.DockerImagePrefix, serviceName)

	if cfg.DockerImagePrefix == "" {
		dockerTag = serviceName
	}

	dockerImage := path.Join(containerRegistry, dockerTag)

	if version == "" {
		version = "1.0.0"
	}

	return fmt.Sprintf("%s:%s", dockerImage, version)
}

func ParseServiceName(prefix, serviceName string) string {
	if prefix == "" {
		return serviceName
	} else {
		return fmt.Sprintf("%s-%s-service", prefix, serviceName)
	}
}

// func cleanDockerTagFile(dockerTag string) error {
// 	if _, err := os.Stat(dockerTag); err == nil {
// 		fmt.Printf("[+] Found existing file %s. Deleting it...\n", dockerTag)
// 		if err := os.Remove(dockerTag); err != nil {
// 			return fmt.Errorf("[!] Failed to delete old file %s: %v", dockerTag, err)
// 		}
// 		fmt.Printf("[+] Deleted old file %s.\n", dockerTag)
// 	} else {
// 		fmt.Printf("[+] No previous file %s found.\n", dockerTag)
// 	}
// 	return nil
// }

// func deleteK8sPod(serviceName string) error {
// 	cmd := exec.Command("kubectl", "delete", "pod", "-l", fmt.Sprintf("app=%s-service", serviceName))
// 	if err := cmd.Run(); err != nil {
// 		return fmt.Errorf("[!] Failed to delete existing pod: %v", err)
// 	}
// 	return nil
// }

// func deployToKubernetes(projectRoot, serviceName string) error {
// 	fmt.Printf("[+] Deploying %s-service to Kubernetes...\n", serviceName)

// 	deploymentFile := fmt.Sprintf("%s/k8s/deployment.dev.yaml", projectRoot)
// 	serviceFile := fmt.Sprintf("%s/k8s/service.yaml", projectRoot)

// 	if err := checkK8sManifestExists(deploymentFile, serviceFile); err != nil {
// 		return err
// 	}

// 	if err := applyK8sManifest(deploymentFile); err != nil {
// 		return err
// 	}

// 	if err := applyK8sManifest(serviceFile); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func checkK8sManifestExists(deploymentFile, serviceFile string) error {
// 	if _, err := os.Stat(deploymentFile); os.IsNotExist(err) {
// 		return fmt.Errorf("[!] Kubernetes manifest for %s-service not found", deploymentFile)
// 	}
// 	if _, err := os.Stat(serviceFile); os.IsNotExist(err) {
// 		return fmt.Errorf("[!] Kubernetes service file for %s-service not found", serviceFile)
// 	}
// 	return nil
// }

// func applyK8sManifest(file string) error {
// 	cmd := exec.Command("kubectl", "apply", "-f", file)
// 	if err := cmd.Run(); err != nil {
// 		return fmt.Errorf("[!] Failed to apply manifest %s: %v", file, err)
// 	}
// 	return nil
// }
