package types

// Base type for the Kubernetes Deployer config json
type K8sDeployerConfig struct {
	DockerImagePrefix       string            `json:"DockerImagePrefix"`
	DockerContainerRegistry DockerRegistry    `json:"DockerContainerRegistry"`
	BuildOutputDirectory    string            `json:"BuildOutputDirectory"`
	KubernetesConfig        KubernetesConfig  `json:"KbernetesConfig"`
	ServicesDirectory       ServicesDirectory `json:"ServicesDirectory"`
}

// Struct for Docker container registry settings
type DockerRegistry struct {
	Dev  string `json:"Dev"`
	Prod string `json:"Prod"`
}

// Struct for Kubernetes configuration
type KubernetesConfig struct {
	Directory DirectoryConfig `json:"Directory"`
	Files     FileConfig      `json:"Files"`
}

// Struct for directory configuration
type DirectoryConfig struct {
	Go     string `json:"Go"`
	Dotnet string `json:"Dotnet"`
}

// Struct for file configuration
type FileConfig struct {
	Dev  EnvironmentFiles `json:"Dev"`
	Prod EnvironmentFiles `json:"Prod"`
}

// Struct for environment-specific file configurations
type EnvironmentFiles struct {
	Deployment string `json:"Deployment"`
	Service    string `json:"Service"`
}

// Struct for services directory
type ServicesDirectory struct {
	Root ServicesDirectoryRoot `json:"Root"`
	All  AllServices           `json:"All"`
}

// Struct for ServicesDirectory.Root
type ServicesDirectoryRoot struct {
	Go     string `json:"Go"`
	Dotnet string `json:"Dotnet"`
}

// Struct for all services
type AllServices struct {
	Go     map[string]string `json:"Go"`
	Dotnet map[string]string `json:"Dotnet"`
}
