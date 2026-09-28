package platform

type ProviderID string

const (
	ProviderAWS   ProviderID = "aws"
	ProviderGCP   ProviderID = "gcp"
	ProviderAzure ProviderID = "azure"
)

type Provider struct {
	ID   ProviderID
	Name string
}

func GetProviders() []Provider {
	return []Provider{
		{ID: ProviderAWS, Name: "AWS"},
		{ID: ProviderGCP, Name: "Google Cloud"},
		{ID: ProviderAzure, Name: "Microsoft Azure"},
	}
}
