# Kubernetes Manifest Diff Tool

A Go-based command-line tool for semantically comparing Kubernetes manifest files. Unlike traditional line-by-line diff tools, k8s-diff understands the structure of Kubernetes YAML objects and provides meaningful comparisons even when files are reordered or formatted differently.

## Features

- **Structural comparison**: Parses YAML objects and compares them semantically rather than line-by-line
- **Multi-object support**: Handles manifests with multiple Kubernetes objects separated by `---`
- **Kubernetes validation**: Validates that all objects have required fields (apiVersion, kind, metadata.name)
- **Clear output**: Shows additions, removals, and modifications in an easy-to-read format
- **Object-aware**: Groups changes by Kubernetes object (ConfigMap, Pod, etc.)
- **Container-aware diffing**: Identifies containers by name for semantic comparison, ignoring reordering
- **Taint indicators**: Red exclamation marks highlight structural changes to container arrays

## Usage

```bash
# Build the binary
go build -o k8s-diff diff.go

# Compare two manifest files
./k8s-diff test_data/scenario1/manifest1.yaml test_data/scenario1/manifest2.yaml

# Run all test scenarios
./test_runner.sh

# Test validation error handling
./test_validation.sh

# Show help
./k8s-diff --help
```

## Test Scenarios

The tool includes comprehensive test scenarios to demonstrate its capabilities:

### Scenario 1: Basic Changes
- **Location**: `test_data/scenario1/`
- **Tests**: ConfigMap data changes, Pod image updates, environment variable modifications

### Scenario 2: Container Reordering
- **Location**: `test_data/scenario2/`
- **Tests**: Two containers in different order between manifests
- **Current Behavior**: Shows no diff (containers compared semantically by name)
- **Note**: Container reordering is correctly ignored as Kubernetes treats it as functionally identical

### Scenario 3: Container Addition
- **Location**: `test_data/scenario3/`
- **Tests**: Pod with sidecar container added
- **Output**: Shows `+ ! container 'name'` with red exclamation mark indicating structural change

### Scenario 4: Container Removal
- **Location**: `test_data/scenario4/`
- **Tests**: Pod with monitoring container removed
- **Output**: Shows `- ! container 'name'` with red exclamation mark indicating structural change

### Validation Tests
- **Location**: `test_data/invalid/`
- **Purpose**: Test Kubernetes object validation with invalid manifests
- **Tests**: Missing apiVersion, kind, metadata, metadata.name; empty name; invalid namespace type
- **Script**: Run `./test_validation.sh` to test all validation scenarios

## Example Output

```
---
apiVersion: v1
kind: Pod
metadata:
  name: example-pod
spec:
  ~ containers:
    ~ container 'nginx':
      ~ image:
        ~~ nginx:1.21
        ~> nginx:1.22

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: example-config
data:
  ~ key1:
    ~~ value1
    ~> value1-changed
  - key2: value2
  + key3: value3
```

## Output Legend

- `+` Addition (Green)
- `-` Removal (Red)
- `~` Modification (Yellow)
  - `~~` Old value
  - `~>` New value
- `!` Taint indicator (Red) - Appears with container additions/removals to highlight structural changes

## Dependencies

- Go 1.19+
- gopkg.in/yaml.v3

## Validation

The tool validates that all Kubernetes objects have the required fields:

- **apiVersion** (required): Kubernetes API version
- **kind** (required): Resource type
- **metadata** (required): Object metadata
- **metadata.name** (required): Must be a non-empty string
- **metadata.namespace** (optional): If present, must be a string

Invalid manifests will produce clear error messages like:
```
Error parsing manifest.yaml: object 1 (ConfigMap): missing required field 'metadata.name'
```

## Installation

1. Clone this repository
2. Initialize Go module: `go mod init k8s-diff`
3. Install dependencies: `go get gopkg.in/yaml.v3`
4. Build the binary: `go build -o k8s-diff diff.go`

## Example Files

The repository includes two sample Kubernetes manifests (`manifest1.yaml` and `manifest2.yaml`) that demonstrate the tool's capabilities with ConfigMaps and Pods.

## Support & Maintenance

This project was created as a demonstration of AI-assisted development. While Claude (the AI assistant) designed and implemented the entire tool, ongoing maintenance and support will depend on community adoption and contributions.

**Current Status**: Functional proof-of-concept with enhanced container diffing capabilities
**Community Contributions**: Welcome! Feel free to fork, improve, and submit pull requests
**Bug Reports**: Please open GitHub issues for any problems you encounter
**Feature Requests**: Community-driven development is encouraged

Note: As an AI-created project, there is no guarantee of active maintenance by the original creator. The codebase is well-documented and designed to be easily understood and modified by other developers.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Development

### Building from Source
```bash
git clone <repository-url>
cd kubernetes-diffing
go build -o k8s-diff diff.go
```

### Project Structure
- `diff.go` - Main application source code (fully documented)
- `README.md` - Project documentation
- `LICENSE` - MIT license
- `.gitignore` - Git ignore patterns (excludes binaries and IDE files)
- `test_runner.sh` - Script to run all test scenarios
- `test_validation.sh` - Script to test validation error handling
- `test_data/` - Test scenarios for demonstrating diff capabilities
  - `scenario1/` - Basic changes (ConfigMap data, Pod image updates)
  - `scenario2/` - Container reordering (shows no changes with semantic diffing)
  - `scenario3/` - Container addition (shows taint indicator)
  - `scenario4/` - Container removal (shows taint indicator)
  - `invalid/` - Invalid manifests for testing validation (missing required fields)

### Git Ignore
The `.gitignore` file excludes:
- Compiled binaries (`k8s-diff`, `k8s-diff-documented`)
- Go build artifacts (`*.exe`, `*.test`, `*.out`)
- IDE files (`.vscode/`, `.idea/`, `*.swp`)
- OS files (`.DS_Store`, `Thumbs.db`)
- Temporary and log files

## Credits

This project was created entirely by Claude (Anthropic's AI assistant) in collaboration with a human user. The complete implementation including the Go code, CLI interface, documentation, and example files were generated by Claude.

Claude also selected the MIT License for this project because it's the most developer-friendly open-source license - it's simple, permissive, allows commercial use, and encourages adoption while requiring only minimal attribution. This makes it ideal for a developer tool like k8s-diff that should be widely accessible and useful to the Kubernetes community.

# Final Notes from the "Developer"
I just didn't want to solve this problem for a 4th time myself. This work will be validated over the coming months. If you make any PR's, I'll shovel them over to Gemini, Claude, or GPT. I can't believe how much easier this is to make this time. 🤣
