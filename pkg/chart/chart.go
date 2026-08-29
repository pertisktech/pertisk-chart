package chart

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Chart represents a Helm chart
type Chart struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	Description  string            `yaml:"description"`
	Home         string            `yaml:"home"`
	Sources      []string          `yaml:"sources"`
	Maintainers  []Maintainer      `yaml:"maintainers"`
	Icon         string            `yaml:"icon"`
	AppVersion   string            `yaml:"appVersion"`
	Deprecated   bool              `yaml:"deprecated"`
	Keywords     []string          `yaml:"keywords"`
	KubeVersion  string            `yaml:"kubeVersion"`
	Type         string            `yaml:"type"`
	Annotations  map[string]string `yaml:"annotations"`
	Dependencies []Dependency      `yaml:"dependencies"`
}

// Maintainer represents a chart maintainer
type Maintainer struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
	URL   string `yaml:"url"`
}

// Dependency represents a chart dependency
type Dependency struct {
	Name       string   `yaml:"name"`
	Version    string   `yaml:"version"`
	Repository string   `yaml:"repository"`
	Condition  string   `yaml:"condition"`
	Tags       []string `yaml:"tags"`
	Enabled    bool     `yaml:"enabled"`
}

// ChartVersion represents a chart version entry in index.yaml
type ChartVersion struct {
	Name        string            `yaml:"name"`
	Version     string            `yaml:"version"`
	Description string            `yaml:"description,omitempty"`
	Home        string            `yaml:"home,omitempty"`
	Sources     []string          `yaml:"sources,omitempty"`
	Maintainers []Maintainer      `yaml:"maintainers,omitempty"`
	Icon        string            `yaml:"icon,omitempty"`
	AppVersion  string            `yaml:"appVersion,omitempty"`
	Deprecated  bool              `yaml:"deprecated,omitempty"`
	Keywords    []string          `yaml:"keywords,omitempty"`
	KubeVersion string            `yaml:"kubeVersion,omitempty"`
	Type        string            `yaml:"type,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
	URLs        []string          `yaml:"urls"`
	Digest      string            `yaml:"digest,omitempty"`
	Created     string            `yaml:"created,omitempty"`
}

// Index represents the Helm repository index
type Index struct {
	APIVersion string                    `yaml:"apiVersion"`
	Generated  string                    `yaml:"generated"`
	Entries    map[string][]ChartVersion `yaml:"entries"`
}

// ParseChartFromTarball extracts and parses Chart.yaml from a tarball
func ParseChartFromTarball(reader io.Reader) (*Chart, error) {
	gzr, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar: %w", err)
		}

		// Look for Chart.yaml
		if filepath.Base(header.Name) == "Chart.yaml" {
			var chart Chart
			if err := yaml.NewDecoder(tr).Decode(&chart); err != nil {
				return nil, fmt.Errorf("failed to decode Chart.yaml: %w", err)
			}
			return &chart, nil
		}
	}

	return nil, fmt.Errorf("Chart.yaml not found in tarball")
}

// ParseChartName extracts chart name from filename
// Format: name-version.tgz
func ParseChartName(filename string) (name, version string, err error) {
	base := filepath.Base(filename)
	if !strings.HasSuffix(base, ".tgz") {
		return "", "", fmt.Errorf("invalid chart filename: %s", filename)
	}

	base = strings.TrimSuffix(base, ".tgz")
	parts := strings.Split(base, "-")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid chart filename format: %s", filename)
	}

	version = parts[len(parts)-1]
	name = strings.Join(parts[:len(parts)-1], "-")
	return name, version, nil
}

// ExtractValuesYAML extracts values.yaml from a chart tarball
func ExtractValuesYAML(reader io.Reader) ([]byte, error) {
	gzr, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar: %w", err)
		}

		// Look for values.yaml in the chart directory
		// Format: chart-name/values.yaml
		if strings.HasSuffix(header.Name, "/values.yaml") || filepath.Base(header.Name) == "values.yaml" {
			valuesData, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("failed to read values.yaml: %w", err)
			}
			return valuesData, nil
		}
	}

	return nil, fmt.Errorf("values.yaml not found in chart tarball")
}

// ExtractREADME extracts README.md from a chart tarball
func ExtractREADME(reader io.Reader) ([]byte, error) {
	gzr, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar: %w", err)
		}

		// Look for README.md in the chart directory
		// Format: chart-name/README.md or README.md
		baseName := filepath.Base(header.Name)
		if baseName == "README.md" || strings.HasSuffix(header.Name, "/README.md") {
			readmeData, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("failed to read README.md: %w", err)
			}
			return readmeData, nil
		}
	}

	return nil, fmt.Errorf("README.md not found in chart tarball")
}

// ExtractNOTES extracts NOTES.txt from a chart tarball
func ExtractNOTES(reader io.Reader) ([]byte, error) {
	gzr, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar: %w", err)
		}

		// Look for NOTES.txt in the chart directory
		baseName := filepath.Base(header.Name)
		if baseName == "NOTES.txt" || strings.HasSuffix(header.Name, "/NOTES.txt") {
			notesData, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("failed to read NOTES.txt: %w", err)
			}
			return notesData, nil
		}
	}

	return nil, fmt.Errorf("NOTES.txt not found in chart tarball")
}

// ResourceInfo represents a Kubernetes resource extracted from manifests
type ResourceInfo struct {
	Kind       string                 `json:"kind"`
	Name       string                 `json:"name"`
	Namespace  string                 `json:"namespace,omitempty"`
	APIVersion string                 `json:"apiVersion"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Manifest   string                 `json:"manifest,omitempty"`
}

// ExtractManifests extracts all template files from a chart tarball
// Note: This extracts raw templates (may contain Helm template syntax)
// For fully rendered manifests, you would need to use Helm's template engine
func ExtractManifests(reader io.Reader) ([]byte, error) {
	// Read all data first since we may need to read multiple times
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read chart data: %w", err)
	}

	gzr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	var manifests []string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar: %w", err)
		}

		// Look for template files (YAML files in templates directory)
		if strings.Contains(header.Name, "/templates/") {
			// Skip non-YAML files and helpers
			ext := filepath.Ext(header.Name)
			if ext != ".yaml" && ext != ".yml" && ext != ".tpl" {
				continue
			}

			// Skip NOTES.txt and helpers
			baseName := filepath.Base(header.Name)
			if baseName == "NOTES.txt" || baseName == "_helpers.tpl" {
				continue
			}

			// Read template content
			templateData, err := io.ReadAll(tr)
			if err != nil {
				continue // Skip files we can't read
			}

			// Add separator and filename comment
			manifest := fmt.Sprintf("---\n# Source: %s\n%s", header.Name, string(templateData))
			manifests = append(manifests, manifest)
		}
	}

	if len(manifests) == 0 {
		return nil, fmt.Errorf("no template files found in chart tarball")
	}

	return []byte(strings.Join(manifests, "\n")), nil
}

// ExtractResources parses Kubernetes resources from chart templates
// Note: This attempts to parse templates, but templates may contain Helm syntax
// For fully rendered resources, you would need to use Helm's template engine
// This function tries to extract resource info even from templates with syntax
func ExtractResources(reader io.Reader) ([]ResourceInfo, error) {
	// Read all data first
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read chart data: %w", err)
	}

	manifests, err := ExtractManifests(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	var resources []ResourceInfo

	// Split manifests by --- separator
	manifestParts := strings.Split(string(manifests), "---")

	for _, part := range manifestParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Remove source comments, but remember the template filename for fallback naming
		lines := strings.Split(part, "\n")
		var yamlLines []string
		var sourceFile string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "# Source:") {
				sourceFile = strings.TrimSpace(strings.TrimPrefix(trimmed, "# Source:"))
				continue
			}
			if trimmed != "" {
				yamlLines = append(yamlLines, line)
			}
		}
		cleanPart := strings.Join(yamlLines, "\n")

		if cleanPart == "" {
			continue
		}

		// Human-readable fallback when the name can't be resolved from template syntax
		templateLabel := filepath.Base(sourceFile)
		templateLabel = strings.TrimSuffix(templateLabel, filepath.Ext(templateLabel))

		// Try to parse as YAML first (works for simple templates)
		var resource map[string]interface{}
		var kind, apiVersion, name, namespace string
		var metadata map[string]interface{}

		if err := yaml.Unmarshal([]byte(cleanPart), &resource); err == nil {
			// Successfully parsed as YAML
			if k, ok := resource["kind"].(string); ok {
				kind = k
			}
			if av, ok := resource["apiVersion"].(string); ok {
				apiVersion = av
			}
			if md, ok := resource["metadata"].(map[string]interface{}); ok {
				metadata = md
				if n, ok := md["name"].(string); ok {
					name = n
				}
				if ns, ok := md["namespace"].(string); ok {
					namespace = ns
				}
			}
		} else {
			// Failed to parse as YAML - try regex extraction for templates with syntax
			kind = extractField(cleanPart, "kind")
			apiVersion = extractField(cleanPart, "apiVersion")
			name = extractField(cleanPart, "name")
			namespace = extractField(cleanPart, "namespace")
			metadata = nil // No metadata available from template syntax
		}

		// If we found kind and name, create a resource entry
		// Also create entry if we found kind (name might be in template syntax)
		if kind != "" {
			if name == "" {
				// Try to extract name from metadata section
				name = extractField(cleanPart, "name")
				if name == "" {
					name = fmt.Sprintf("<templated: %s>", templateLabel)
				}
			}
			// If metadata wasn't set from YAML parsing, try to parse again
			if metadata == nil {
				if err := yaml.Unmarshal([]byte(cleanPart), &resource); err == nil {
					if md, ok := resource["metadata"].(map[string]interface{}); ok {
						metadata = md
						// Override with parsed values if available
						if k, ok := resource["kind"].(string); ok && k != "" {
							kind = k
						}
						if av, ok := resource["apiVersion"].(string); ok && av != "" {
							apiVersion = av
						}
						if n, ok := md["name"].(string); ok && n != "" {
							name = n
						}
						if ns, ok := md["namespace"].(string); ok {
							namespace = ns
						}
					}
				}
			}

			resources = append(resources, ResourceInfo{
				Kind:       kind,
				Name:       name,
				Namespace:  namespace,
				APIVersion: apiVersion,
				Metadata:   metadata,
				Manifest:   cleanPart,
			})
		} else if strings.Contains(cleanPart, "kind:") || strings.Contains(cleanPart, "apiVersion:") {
			// Found YAML-like structure but couldn't extract - might be a resource template
			// Extract what we can
			kind = extractField(cleanPart, "kind")
			if kind == "" {
				kind = "Unknown"
			}
			name = extractField(cleanPart, "name")
			if name == "" {
				name = fmt.Sprintf("<templated: %s>", templateLabel)
			}
			apiVersion = extractField(cleanPart, "apiVersion")
			if apiVersion == "" {
				apiVersion = "v1"
			}

			resources = append(resources, ResourceInfo{
				Kind:       kind,
				Name:       name,
				Namespace:  namespace,
				APIVersion: apiVersion,
				Metadata:   nil,
				Manifest:   cleanPart,
			})
		}
	}

	return resources, nil
}

// extractField tries to extract a field value from YAML/template content
// It looks for patterns like "kind: Deployment" or "kind: {{ .Values.kind }}"
func extractField(content, fieldName string) string {
	// Pattern 1: Simple YAML format: "kind: Deployment"
	pattern1 := fmt.Sprintf(`(?m)^\s*%s\s*:\s*(.+)$`, fieldName)
	re1 := regexp.MustCompile(pattern1)
	matches := re1.FindStringSubmatch(content)
	if len(matches) > 1 {
		value := strings.TrimSpace(matches[1])
		// Remove quotes if present
		value = strings.Trim(value, `"'`)
		// Remove template syntax markers but keep the value
		if strings.Contains(value, "{{") {
			// Try to extract default value from template
			reDefault := regexp.MustCompile(`\|\s*default\s+["']?([^"']+)["']?`)
			if defaultMatch := reDefault.FindStringSubmatch(value); len(defaultMatch) > 1 {
				return strings.TrimSpace(defaultMatch[1])
			}
			// If no default, try to extract from .Values or .Release
			reValues := regexp.MustCompile(`\.Values\.(\w+)`)
			if valMatch := reValues.FindStringSubmatch(value); len(valMatch) > 1 {
				return valMatch[1] // Return the value key name
			}
			reRelease := regexp.MustCompile(`\.Release\.(\w+)`)
			if relMatch := reRelease.FindStringSubmatch(value); len(relMatch) > 1 {
				return relMatch[1]
			}
			// Try to extract from .Chart
			reChart := regexp.MustCompile(`\.Chart\.(\w+)`)
			if chartMatch := reChart.FindStringSubmatch(value); len(chartMatch) > 1 {
				return chartMatch[1]
			}
			// Common helper pattern: {{ include "chart.fullname" . }} - surface the helper name
			reInclude := regexp.MustCompile(`include\s+"([^"]+)"`)
			if includeMatch := reInclude.FindStringSubmatch(value); len(includeMatch) > 1 {
				parts := strings.Split(includeMatch[1], ".")
				return parts[len(parts)-1]
			}
			return "" // Template syntax without extractable value
		}
		return value
	}
	return ""
}
