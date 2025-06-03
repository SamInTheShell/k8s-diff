/*
Package main implements k8s-diff, a semantic Kubernetes manifest diff tool.

k8s-diff compares Kubernetes YAML manifests by parsing them into structured objects
rather than performing line-by-line text comparison. This approach provides more
meaningful diffs that understand the hierarchical nature of Kubernetes resources.

Architecture Overview:
1. CLI argument parsing and validation
2. YAML parsing into K8sObject structs
3. Object identification and mapping by kind/name
4. Recursive semantic comparison
5. Color-coded visual diff output

Key Features:
- Handles multi-document YAML files (separated by ---)
- Identifies objects by kind and metadata.name
- Recursive comparison of nested maps and arrays
- ANSI color-coded output for different change types
- Cross-platform terminal compatibility

Author: Created with AI assistance
License: MIT
*/
package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// ANSI color codes optimized for both light and dark terminal backgrounds.
// These bright variants ensure good contrast and readability across different themes.
const (
	ColorRed    = "\033[91m" // Bright red - for removals/deletions
	ColorGreen  = "\033[92m" // Bright green - for additions
	ColorYellow = "\033[93m" // Bright yellow - for modifications
	ColorWhite  = "\033[97m" // Bright white - for context/unchanged
	ColorReset  = "\033[0m"  // Reset to terminal default color
)

// helpText contains the CLI usage documentation displayed when users
// run the tool with -h, --help, or with incorrect arguments.
const helpText = `k8s-diff - A semantic Kubernetes manifest diff tool

USAGE:
    k8s-diff [OPTIONS] <file1> <file2>

ARGUMENTS:
    <file1>    First Kubernetes manifest file
    <file2>    Second Kubernetes manifest file

OPTIONS:
    -h, --help    Show this help message

EXAMPLES:
    k8s-diff manifest1.yaml manifest2.yaml
    k8s-diff old-deployment.yaml new-deployment.yaml

DESCRIPTION:
    k8s-diff compares Kubernetes manifest files semantically, understanding
    the structure of YAML objects rather than doing line-by-line comparison.

    Output uses color coding:
    - Red: Removals and taint indicators (!)
    - Green: Additions
    - Yellow: Modifications (shown as ~~ old_value and ~> new_value)
    - White: Unchanged elements

    The taint indicator (!) appears with container additions/removals to
    highlight structural changes to container arrays.
`

// K8sObject represents a Kubernetes resource with the most common fields.
// This struct captures the essential structure of most Kubernetes objects
// while using interface{} for flexible handling of varying content.
//
// Fields:
//   - APIVersion: Kubernetes API version (e.g., "v1", "apps/v1")
//   - Kind: Resource type (e.g., "Pod", "Deployment", "ConfigMap")
//   - Metadata: Object metadata including name, namespace, labels, etc.
//   - Data: Used primarily by ConfigMaps and Secrets
//   - Spec: Resource specification used by most workload resources
//
// The omitempty tags ensure that nil fields don't appear in YAML output.
type K8sObject struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   map[string]interface{} `yaml:"metadata"`
	Data       map[string]interface{} `yaml:"data,omitempty"`
	Spec       map[string]interface{} `yaml:"spec,omitempty"`
}

// main orchestrates the entire diff process:
// 1. Parse and validate CLI arguments
// 2. Check file existence
// 3. Parse YAML files into K8sObject structs
// 4. Perform semantic comparison and output results
//
// Error handling: All errors are printed to stderr with appropriate exit codes.
func main() {
	args := os.Args[1:] // Skip program name

	// Handle help flags - show help and exit gracefully
	if len(args) == 0 || contains(args, "-h") || contains(args, "--help") {
		fmt.Print(helpText)
		if len(args) == 0 {
			os.Exit(1) // Error exit for no args
		}
		os.Exit(0) // Success exit for explicit help request
	}

	// Validate argument count - exactly 2 file paths required
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Error: Expected exactly 2 file arguments, got %d\n\n", len(args))
		fmt.Print(helpText)
		os.Exit(1)
	}

	file1 := args[0]
	file2 := args[1]

	// Verify both files exist before attempting to parse them
	if err := checkFileExists(file1); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := checkFileExists(file2); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Parse YAML files into structured objects
	objects1, err := parseK8sObjects(file1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", file1, err)
		os.Exit(1)
	}

	objects2, err := parseK8sObjects(file2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", file2, err)
		os.Exit(1)
	}

	// Perform semantic diff and output results
	diffK8sObjects(objects1, objects2)
}

// contains checks if a string slice contains a specific string.
// Used for CLI argument parsing to detect help flags.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// checkFileExists verifies that a file exists and is accessible.
// Returns a descriptive error if the file doesn't exist or can't be accessed.
func checkFileExists(filename string) error {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return fmt.Errorf("file '%s' does not exist", filename)
	}
	return err
}

// parseK8sObjects reads a YAML file and parses it into a slice of K8sObject structs.
// Handles multi-document YAML files by splitting on "---" separators.
// Validates that each object has the required Kubernetes fields.
//
// The function:
// 1. Reads the entire file content
// 2. Splits by "---" to handle multiple Kubernetes objects
// 3. Parses each document as a separate K8sObject
// 4. Validates each object for required Kubernetes fields
// 5. Skips empty documents
//
// Returns: slice of parsed and validated objects and any parsing/validation error
func parseK8sObjects(filename string) ([]K8sObject, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Split by --- to handle multiple objects in a single YAML file
	docs := strings.Split(string(content), "---")
	var objects []K8sObject

	for i, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue // Skip empty documents
		}

		var obj K8sObject
		if err := yaml.Unmarshal([]byte(doc), &obj); err != nil {
			return nil, fmt.Errorf("failed to parse object %d: %v", i+1, err)
		}

		// Validate the parsed object
		if err := validateK8sObject(obj, i+1); err != nil {
			return nil, err
		}

		objects = append(objects, obj)
	}

	return objects, nil
}

// validateK8sObject checks that a parsed object has the required Kubernetes fields.
// All Kubernetes objects must have: apiVersion, kind, and metadata.name.
// The metadata.namespace field is optional (defaults to "default" when not specified).
//
// Parameters:
//   - obj: The parsed K8sObject to validate
//   - objNum: Object number for error reporting (1-based)
//
// Returns: error if validation fails, nil if object is valid
func validateK8sObject(obj K8sObject, objNum int) error {
	// Check required apiVersion field
	if obj.APIVersion == "" {
		return fmt.Errorf("object %d: missing required field 'apiVersion'", objNum)
	}

	// Check required kind field
	if obj.Kind == "" {
		return fmt.Errorf("object %d: missing required field 'kind'", objNum)
	}

	// Check that metadata exists
	if obj.Metadata == nil {
		return fmt.Errorf("object %d (%s): missing required field 'metadata'", objNum, obj.Kind)
	}

	// Check required metadata.name field
	name, hasName := obj.Metadata["name"]
	if !hasName {
		return fmt.Errorf("object %d (%s): missing required field 'metadata.name'", objNum, obj.Kind)
	}

	// Validate that name is a non-empty string
	if nameStr, ok := name.(string); !ok || nameStr == "" {
		return fmt.Errorf("object %d (%s): 'metadata.name' must be a non-empty string, got %T", objNum, obj.Kind, name)
	}

	// Validate namespace if present (must be a string)
	if namespace, hasNamespace := obj.Metadata["namespace"]; hasNamespace {
		if _, ok := namespace.(string); !ok {
			return fmt.Errorf("object %d (%s/%s): 'metadata.namespace' must be a string, got %T", objNum, obj.Kind, name, namespace)
		}
	}
	// Note: namespace is optional - if not specified, Kubernetes defaults to "default"

	return nil
}

// diffK8sObjects performs the high-level comparison between two sets of Kubernetes objects.
//
// Algorithm:
// 1. Create lookup maps keyed by "Kind/Name" for O(1) object identification
// 2. Find objects that exist only in file1 (removals)
// 3. Find objects that exist only in file2 (additions)
// 4. Compare objects that exist in both files (modifications)
//
// This approach handles:
// - Objects added or removed between files
// - Objects that exist in both but have different content
// - Maintains object identity across comparisons
func diffK8sObjects(objects1, objects2 []K8sObject) {
	// Create maps for O(1) lookup by kind/name combination
	map1 := make(map[string]K8sObject)
	map2 := make(map[string]K8sObject)

	// Build lookup map for first file's objects
	for _, obj := range objects1 {
		key := getObjectKey(obj)
		map1[key] = obj
	}

	// Build lookup map for second file's objects
	for _, obj := range objects2 {
		key := getObjectKey(obj)
		map2[key] = obj
	}

	// Identify objects removed (exist in file1 but not file2)
	for key, obj := range map1 {
		if _, exists := map2[key]; !exists {
			fmt.Printf("%s- %s %s (removed)%s\n", ColorRed, obj.Kind, getObjectName(obj), ColorReset)
		}
	}

	// Identify objects added (exist in file2 but not file1)
	for key, obj := range map2 {
		if _, exists := map1[key]; !exists {
			fmt.Printf("%s+ %s %s (added)%s\n", ColorGreen, obj.Kind, getObjectName(obj), ColorReset)
		}
	}

	// Compare objects that exist in both files for modifications
	for key, obj1 := range map1 {
		if obj2, exists := map2[key]; exists {
			diffObject(obj1, obj2)
		}
	}
}

// getObjectKey creates a unique identifier for a Kubernetes object.
// Format: "Kind/Name" or "Kind/Namespace/Name" if namespace is specified
// This key is used for object lookup and comparison between files.
//
// Examples:
//   - "Pod/nginx" (default namespace)
//   - "Pod/kube-system/nginx" (explicit namespace)
//   - "ConfigMap/app-config" (cluster-scoped or default namespace)
func getObjectKey(obj K8sObject) string {
	name := getObjectName(obj)
	namespace := getObjectNamespace(obj)

	if namespace != "" && namespace != "default" {
		return fmt.Sprintf("%s/%s/%s", obj.Kind, namespace, name)
	}
	return fmt.Sprintf("%s/%s", obj.Kind, name)
}

// getObjectName extracts the name from a Kubernetes object's metadata.
// Since validation ensures the name exists and is a string, this should not fail
// for valid objects. Returns "unknown" only as a defensive fallback.
func getObjectName(obj K8sObject) string {
	if name, ok := obj.Metadata["name"].(string); ok {
		return name
	}
	return "unknown" // Should not happen after validation
}

// getObjectNamespace extracts the namespace from a Kubernetes object's metadata.
// Returns empty string if namespace is not specified (indicating default namespace).
func getObjectNamespace(obj K8sObject) string {
	if namespace, ok := obj.Metadata["namespace"].(string); ok {
		return namespace
	}
	return "" // No namespace specified - defaults to "default"
}

// diffObject performs detailed comparison between two K8sObject instances.
// Only outputs diff information if the objects are actually different.
//
// The function compares each major section:
// - apiVersion and kind (basic object identity)
// - metadata (name, namespace, labels, annotations, etc.)
// - data (for ConfigMaps and Secrets)
// - spec (for workload resources like Pods, Deployments)
//
// Output format mimics YAML structure with "---" separators and proper indentation.
// Uses color coding to distinguish between unchanged and modified sections.
func diffObject(obj1, obj2 K8sObject) {
	// Skip output if objects are identical
	if !reflect.DeepEqual(obj1, obj2) {
		fmt.Printf("\n---\n") // YAML document separator

		// Compare apiVersion field
		if obj1.APIVersion == obj2.APIVersion {
			fmt.Printf("apiVersion: %s\n", obj1.APIVersion)
		} else {
			fmt.Printf("%s~~ apiVersion: %s%s\n", ColorYellow, obj1.APIVersion, ColorReset)
			fmt.Printf("%s~> apiVersion: %s%s\n", ColorYellow, obj2.APIVersion, ColorReset)
		}

		// Compare kind field
		if obj1.Kind == obj2.Kind {
			fmt.Printf("kind: %s\n", obj1.Kind)
		} else {
			fmt.Printf("%s~~ kind: %s%s\n", ColorYellow, obj1.Kind, ColorReset)
			fmt.Printf("%s~> kind: %s%s\n", ColorYellow, obj2.Kind, ColorReset)
		}

		// Compare metadata section
		if reflect.DeepEqual(obj1.Metadata, obj2.Metadata) {
			fmt.Printf("metadata:\n")
			printYAMLValue("  ", obj1.Metadata, false)
		} else {
			fmt.Printf("%smetadata:%s\n", ColorYellow, ColorReset)
			diffAnyValue("  ", obj1.Metadata, obj2.Metadata)
		}

		// Compare data section (ConfigMaps, Secrets)
		if obj1.Data != nil || obj2.Data != nil {
			if reflect.DeepEqual(obj1.Data, obj2.Data) {
				if obj1.Data != nil {
					fmt.Printf("data:\n")
					printYAMLValue("  ", obj1.Data, false)
				}
			} else {
				fmt.Printf("%sdata:%s\n", ColorYellow, ColorReset)
				diffAnyValue("  ", obj1.Data, obj2.Data)
			}
		}

		// Compare spec section (Pods, Deployments, Services, etc.)
		if obj1.Spec != nil || obj2.Spec != nil {
			if reflect.DeepEqual(obj1.Spec, obj2.Spec) {
				if obj1.Spec != nil {
					fmt.Printf("spec:\n")
					printYAMLValue("  ", obj1.Spec, false)
				}
			} else {
				fmt.Printf("%sspec:%s\n", ColorYellow, ColorReset)
				diffAnyValue("  ", obj1.Spec, obj2.Spec)
			}
		}
	}
}

// printYAMLValue recursively prints a YAML value with proper indentation and structure.
// Handles maps, slices, and scalar values while maintaining YAML formatting.
//
// Parameters:
//   - indent: Current indentation level (grows with nesting depth)
//   - value: The value to print (map, slice, or scalar)
//   - isChanged: Whether to apply color highlighting for changes
//
// This function recreates YAML structure for consistent output formatting.
func printYAMLValue(indent string, value interface{}, isChanged bool) {
	color := ""
	reset := ""
	if isChanged {
		color = ColorYellow
		reset = ColorReset
	}

	switch v := value.(type) {
	case map[string]interface{}:
		// Handle nested maps (e.g., metadata.labels, spec.containers)
		for key, val := range v {
			switch val.(type) {
			case map[string]interface{}, []interface{}:
				// Complex values get their own line with increased indentation
				fmt.Printf("%s%s%s:%s\n", indent, color, key, reset)
				printYAMLValue(indent+"  ", val, isChanged)
			default:
				// Simple key-value pairs on one line
				fmt.Printf("%s%s%s: %v%s\n", indent, color, key, val, reset)
			}
		}
	case []interface{}:
		// Handle arrays (e.g., containers, volumes, env variables)
		for _, item := range v {
			fmt.Printf("%s%s-%s\n", indent, color, reset)
			printYAMLValue(indent+"  ", item, isChanged)
		}
	default:
		// Handle scalar values (strings, numbers, booleans)
		fmt.Printf("%s%s%v%s\n", indent, color, value, reset)
	}
}

// diffAnyValue is the core recursive comparison function that handles any Go value type.
// It dispatches to specialized diff functions based on the value type.
//
// Type handling:
//   - map[string]interface{}: Calls diffMaps for key-by-key comparison
//   - []interface{}: Calls diffSlices for element-by-element comparison
//   - Other types: Direct value comparison with ~~/~> format for changes
//
// This function is the heart of the semantic diff algorithm.
func diffAnyValue(indent string, val1, val2 interface{}) {
	switch v1 := val1.(type) {
	case map[string]interface{}:
		if v2, ok := val2.(map[string]interface{}); ok {
			// Both values are maps - compare them structurally
			diffMaps(indent, v1, v2)
		} else {
			// Type mismatch - show as complete replacement
			fmt.Printf("%s%s~~ %s%s\n", indent, ColorYellow, formatValue(val1), ColorReset)
			fmt.Printf("%s%s~> %s%s\n", indent, ColorYellow, formatValue(val2), ColorReset)
		}
	case []interface{}:
		if v2, ok := val2.([]interface{}); ok {
			// Both values are arrays - compare them element-wise
			if isContainerArray(v1) && isContainerArray(v2) {
				diffContainerArrays(indent, v1, v2)
			} else {
				diffSlices(indent, v1, v2)
			}
		} else {
			// Type mismatch - show as complete replacement
			fmt.Printf("%s%s~~ %s%s\n", indent, ColorYellow, formatValue(val1), ColorReset)
			fmt.Printf("%s%s~> %s%s\n", indent, ColorYellow, formatValue(val2), ColorReset)
		}
	default:
		// Scalar values (strings, numbers, booleans) - direct comparison
		if !reflect.DeepEqual(val1, val2) {
			fmt.Printf("%s%s~~ %s%s\n", indent, ColorYellow, formatValue(val1), ColorReset)
			fmt.Printf("%s%s~> %s%s\n", indent, ColorYellow, formatValue(val2), ColorReset)
		}
	}
}

// diffMaps compares two maps key by key, identifying additions, removals, and modifications.
//
// Algorithm:
// 1. Create a union of all keys from both maps
// 2. For each key, determine if it was added, removed, or modified
// 3. Recursively compare modified values
//
// Output symbols:
//   - "+": Key exists only in map2 (addition)
//   - "-": Key exists only in map1 (removal)
//   - "~": Key exists in both but values differ (modification)
//
// This handles nested structures like metadata.labels, spec.containers, etc.
func diffMaps(indent string, map1, map2 map[string]interface{}) {
	// Build union of all keys from both maps
	allKeys := make(map[string]bool)
	for key := range map1 {
		allKeys[key] = true
	}
	for key := range map2 {
		allKeys[key] = true
	}

	// Compare each key's presence and value
	for key := range allKeys {
		val1, exists1 := map1[key]
		val2, exists2 := map2[key]

		if !exists1 {
			// Key was added in map2
			fmt.Printf("%s%s+ %s: %s%s\n", indent, ColorGreen, key, formatValue(val2), ColorReset)
		} else if !exists2 {
			// Key was removed from map1
			fmt.Printf("%s%s- %s: %s%s\n", indent, ColorRed, key, formatValue(val1), ColorReset)
		} else if !reflect.DeepEqual(val1, val2) {
			// Key exists in both but values differ
			fmt.Printf("%s%s~ %s:%s\n", indent, ColorYellow, key, ColorReset)
			diffAnyValue(indent+"  ", val1, val2)
		}
	}
}

// diffSlices compares two slices element by element.
// For arrays of different lengths or complex nested changes, shows complete replacement.
// For arrays of same length, compares each index position recursively.
//
// Kubernetes arrays this handles:
//   - spec.containers (container definitions)
//   - spec.volumes (volume mounts)
//   - metadata.labels (when stored as arrays)
//   - env variables, ports, etc.
//
// Limitation: Currently optimized for simple cases. Could be enhanced with
// LCS (Longest Common Subsequence) algorithm for better array diff visualization.
func diffSlices(indent string, slice1, slice2 []interface{}) {
	// For arrays of different lengths, show complete replacement
	// This handles cases where containers are added/removed
	if len(slice1) != len(slice2) {
		fmt.Printf("%s%s~~ %s%s\n", indent, ColorYellow, formatValue(slice1), ColorReset)
		fmt.Printf("%s%s~> %s%s\n", indent, ColorYellow, formatValue(slice2), ColorReset)
		return
	}

	// For same-length arrays, compare element by element
	for i := 0; i < len(slice1); i++ {
		if !reflect.DeepEqual(slice1[i], slice2[i]) {
			fmt.Printf("%s%s[%d]:%s\n", indent, ColorYellow, i, ColorReset)
			diffAnyValue(indent+"  ", slice1[i], slice2[i])
		}
	}
}

// formatValue converts any Go value to a clean string representation suitable for diff output.
// Uses YAML marshaling for structured data to maintain consistency with input format.
//
// Handling strategy:
//   - Multi-line values: Shows first line + "..." for brevity
//   - Single-line values: Shows complete value
//   - Complex objects: YAML-formatted for readability
//   - Marshaling errors: Falls back to Go's default %v formatting
//
// This ensures diff output remains readable even for large nested structures.
func formatValue(val interface{}) string {
	yamlBytes, err := yaml.Marshal(val)
	if err != nil {
		// Fallback to Go's default string representation
		return fmt.Sprintf("%v", val)
	}

	// Clean up the YAML output for inline display
	yamlStr := strings.TrimSpace(string(yamlBytes))
	if strings.Contains(yamlStr, "\n") {
		// Multi-line: just show first line with ... for compactness
		firstLine := strings.Split(yamlStr, "\n")[0]
		return firstLine + " ..."
	}
	return yamlStr
}

// isContainerArray checks if we're dealing with a Kubernetes containers array
// by examining the structure for container-like objects with name and image fields.
func isContainerArray(slice []interface{}) bool {
	if len(slice) == 0 {
		return false
	}

	// Check if first element looks like a container (has name and image)
	if container, ok := slice[0].(map[string]interface{}); ok {
		_, hasName := container["name"]
		_, hasImage := container["image"]
		return hasName && hasImage
	}

	return false
}

// diffContainerArrays provides specialized diffing for Kubernetes container arrays.
// Containers are identified by name rather than array position, providing more
// semantic diff output for container additions, removals, and reordering.
//
// The red exclamation mark (!) indicator shows when the container array is "tainted"
// by additions or removals, helping users quickly identify structural changes.
func diffContainerArrays(indent string, slice1, slice2 []interface{}) {
	// Build maps keyed by container name for semantic comparison
	containers1 := make(map[string]interface{})
	containers2 := make(map[string]interface{})

	// Extract containers by name from first array
	for _, container := range slice1 {
		if c, ok := container.(map[string]interface{}); ok {
			if name, ok := c["name"].(string); ok {
				containers1[name] = container
			}
		}
	}

	// Extract containers by name from second array
	for _, container := range slice2 {
		if c, ok := container.(map[string]interface{}); ok {
			if name, ok := c["name"].(string); ok {
				containers2[name] = container
			}
		}
	}

	// Find all container names across both arrays
	allNames := make(map[string]bool)
	for name := range containers1 {
		allNames[name] = true
	}
	for name := range containers2 {
		allNames[name] = true
	}

	// Check if array is "tainted" by additions or removals
	hasTaint := len(containers1) != len(containers2)

	// Compare containers by name
	for name := range allNames {
		container1, exists1 := containers1[name]
		container2, exists2 := containers2[name]

		if !exists1 {
			// Container added - mark as tainted
			taintIndicator := ""
			if hasTaint {
				taintIndicator = fmt.Sprintf("%s! %s", ColorRed, ColorReset)
			}
			fmt.Printf("%s%s+ %scontainer '%s': %s%s\n", indent, ColorGreen, taintIndicator, name, formatValue(container2), ColorReset)
		} else if !exists2 {
			// Container removed - mark as tainted
			taintIndicator := ""
			if hasTaint {
				taintIndicator = fmt.Sprintf("%s! %s", ColorRed, ColorReset)
			}
			fmt.Printf("%s%s- %scontainer '%s': %s%s\n", indent, ColorRed, taintIndicator, name, formatValue(container1), ColorReset)
		} else if !reflect.DeepEqual(container1, container2) {
			// Container modified (no taint indicator for modifications)
			fmt.Printf("%s%s~ container '%s':%s\n", indent, ColorYellow, name, ColorReset)
			diffAnyValue(indent+"  ", container1, container2)
		}
	}
}
