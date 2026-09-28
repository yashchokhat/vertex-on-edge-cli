# Vertex-on-Edge

Infrastructure without the DevOps overhead.

Vertex-on-Edge is an open-source CLI designed to allow developers to deploy existing applications directly to their own cloud infrastructure. It removes the need to manually learn Docker, CI/CD pipelines, or cloud provisioning.

Currently, the project is in its first milestone: providing an interactive, cross-platform CLI interface and an intelligent technology stack detection engine.

## Operating System Support

The CLI is fully cross-platform and natively supported on:
- macOS (Intel and Apple Silicon)
- Windows (10 and 11)
- Linux (Ubuntu, Debian, Fedora, Arch, and other major distributions)

It includes native integrations for file system navigation, utilizing Finder on macOS, File Explorer on Windows, and standard dialogs on Linux.

## Supported Technology Stacks

The detection engine uses heuristic analysis of your project's directory structure, configuration files, and package manifests to accurately identify your stack. 

Currently supported frameworks and environments include:

**Frontend and Fullstack Web**
- Next.js
- React
- Vue
- Angular
- Svelte and SvelteKit
- Nuxt
- Remix
- NestJS
- Express.js
- Generic Node.js

**Backend and API**
- Django, FastAPI, Flask, and Generic Python
- Spring Boot and Generic Java
- Go (Standard, Gin, Echo)
- Laravel and Generic PHP
- Ruby on Rails and Generic Ruby
- Rust

**Mobile and Hybrid**
- Flutter
- React Native
- Capacitor
- Android (Native Java and Kotlin)
- iOS (Native Swift and Objective-C)

**Containers**
- Generic Docker Containers

## Roadmap

The long-term vision of Vertex-on-Edge is to completely automate the journey from local source code to public application:

1. Detect Tech Stack (Completed)
2. Generate Dockerfile
3. Build Container
4. Test and Security Scan
5. Create GitHub Repository
6. Configure GitHub Actions
7. Push to Container Registry
8. Provision AWS EC2
9. Deploy Public Application

## Installation

### Prerequisites

Ensure you have Go 1.24 or higher installed on your system.

### Build from Source

Clone the repository and run the build command. This will generate a standalone binary for your operating system.

```bash
git clone https://github.com/yashchokhat/vertex-on-edge.git
cd vertex-on-edge
make build
```

The compiled binary will be located at `bin/vertex-on-edge`.

## Usage

### Analyze a Project

To begin the interactive startup experience and detect a project's technology stack, run the CLI without any arguments:

```bash
cd your-project
vertex-on-edge
```

The CLI will prompt you to accept the terms of service, ask you to select a target project directory (either the current directory, a manual path, or via a native file browser window), and then analyze the source code.

### Direct Detection

To run the technology detection immediately on the current directory without the interactive setup menus:

```bash
vertex-on-edge detect
```

## Setup Guide

To start using Vertex-on-Edge locally across your projects, you can install the binary globally on your system.

### Option 1: Install via Go (Recommended)

If you have Go installed, you can compile and install it directly to your `GOPATH`:

```bash
go install github.com/yashchokhat/vertex-on-edge/cmd/vertex-on-edge@latest
```

Ensure `~/go/bin` is in your system's `PATH`.

### Option 2: Build and Link Manually

```bash
git clone https://github.com/yashchokhat/vertex-on-edge.git
cd vertex-on-edge
make build

# Move the binary to a global bin directory
sudo mv bin/vertex-on-edge /usr/local/bin/
```

Once installed, simply run `vertex-on-edge` from any project directory.

## Development

The project uses a standard Makefile for common tasks.

```bash
make build    # Compiles the binary
make run      # Compiles and immediately executes the CLI
make test     # Runs the test suite
make lint     # Formats the code and runs static analysis
make clean    # Removes compiled binaries
```

### Architecture

The codebase is structured to be extensible, making it easy to add subsequent modules as the project grows.

- `cmd/vertex-on-edge/` - Contains the main entry point.
- `internal/cli/` - Defines the Cobra commands and execution flow.
- `internal/config/` - Houses configuration constants and supported stack definitions.
- `internal/detector/` - Contains the heuristic scanning engine and language-specific rules.
- `internal/ui/` - Provides terminal formatting, interactive prompts, and cross-platform native dialog wrappers.
- `pkg/models/` - Exposes public data models used across the application.

## Legal

### License
This project is licensed under the MIT License. See the `LICENSE` file for details.

### Terms and Conditions
By using Vertex-on-Edge, you agree that you are responsible for your own deployments and cloud provider costs. Read the full [Terms and Conditions](https://github.com/yashchokhat/vertex-on-edge/blob/main/TERMS.md).
