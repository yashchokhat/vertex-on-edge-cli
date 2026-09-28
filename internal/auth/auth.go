package auth

type Provider interface {
	IsAuthenticated() bool
	Authenticate() error
	GetIdentifier() string
}
