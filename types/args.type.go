package types

// Command line arguments
type Args struct {
	DeployTo         string // "dev" or "prod"
	MicroserviceType string // "go" or "dotnet"
}
