package chart

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"path/filepath"
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
	Annotations  map[string]string  `yaml:"annotations"`
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
	Name       string `yaml:"name"`
	Version    string `yaml:"version"`
	Repository string `yaml:"repository"`
	Condition  string `yaml:"condition"`
	Tags       []string `yaml:"tags"`
	Enabled    bool   `yaml:"enabled"`
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
	APIVersion string                  `yaml:"apiVersion"`
	Generated  string                  `yaml:"generated"`
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

