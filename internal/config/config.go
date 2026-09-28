package config

// Version constants for the CLI. These can be overridden at build time
// using ldflags: -ldflags "-X github.com/yashchokhat/vertex-on-edge/internal/config.Version=1.0.0"
var (
	Version   = "0.1.0"
	Commit    = "dev"
	BuildDate = "unknown"
)

const (
	// AppName is the display name of the application.
	AppName = "Vertex-on-Edge"

	// AppTagline is the application tagline.
	AppTagline = "Infrastructure without the DevOps overhead."

	// GitHubURL is the project's GitHub URL.
	GitHubURL = "github.com/yashchokhat/vertex-on-edge"

	// CLIBinary is the expected binary name.
	CLIBinary = "vertex-on-edge"
)

// Config holds runtime configuration for the CLI.
// This will be expanded as new modules are added.
type Config struct {
	// ProjectPath is the path to the project being analyzed.
	ProjectPath string

	// Verbose enables verbose output.
	Verbose bool

	// NoAnimation disables startup animations.
	NoAnimation bool

	// NoColor disables colored output.
	NoColor bool
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{}
}

// SupportedStacks returns the list of technologies that Vertex-on-Edge can detect or configure.
func SupportedStacks() []string {
	return []string{
		"Next.js",
		"React",
		"Vue",
		"Angular",
		"Svelte / SvelteKit",
		"Nuxt",
		"Remix",
		"NestJS",
		"Express.js",
		"Node.js (Generic)",
		"Django",
		"FastAPI",
		"Flask",
		"Python (Generic)",
		"Go (Standard / Gin / Echo)",
		"Spring Boot",
		"Java (Generic)",
		"Rust",
		"Laravel",
		"PHP (Generic)",
		"Ruby on Rails",
		"Ruby (Generic)",
		"Flutter",
		"Android (Native)",
		"iOS (Native)",
		"React Native",
		"Capacitor",
		"Docker (Generic Container)",
	}
}
