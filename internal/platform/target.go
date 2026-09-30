package platform

type TargetID string

const (
	TargetEC2           TargetID = "ec2"
	TargetECS           TargetID = "ecs"
	TargetEKS           TargetID = "eks"
	TargetAppRunner     TargetID = "app-runner"
	TargetCloudRun      TargetID = "cloud-run"
	TargetComputeEngine TargetID = "compute-engine"
	TargetGKE           TargetID = "gke"
	TargetContainerApps TargetID = "container-apps"
	TargetAppService    TargetID = "app-service"
	TargetAKS           TargetID = "aks"
)

type Target struct {
	ID        TargetID
	Name      string
	Provider  ProviderID
	Available bool
}

func GetTargetsForProvider(provider ProviderID) []Target {
	switch provider {
	case ProviderAWS:
		return []Target{
			{ID: TargetEC2, Name: "EC2", Provider: ProviderAWS, Available: true},
			{ID: TargetECS, Name: "ECS", Provider: ProviderAWS, Available: false},
			{ID: TargetEKS, Name: "EKS", Provider: ProviderAWS, Available: false},
			{ID: TargetAppRunner, Name: "App Runner", Provider: ProviderAWS, Available: false},
		}
	case ProviderGCP:
		return []Target{
			{ID: TargetCloudRun, Name: "Cloud Run", Provider: ProviderGCP, Available: true},
			{ID: TargetComputeEngine, Name: "Compute Engine", Provider: ProviderGCP, Available: false},
			{ID: TargetGKE, Name: "GKE", Provider: ProviderGCP, Available: false},
		}
	case ProviderAzure:
		return []Target{
			{ID: TargetContainerApps, Name: "Container Apps", Provider: ProviderAzure, Available: true},
			{ID: TargetAppService, Name: "App Service", Provider: ProviderAzure, Available: false},
			{ID: TargetAKS, Name: "AKS", Provider: ProviderAzure, Available: false},
		}
	}
	return nil
}
