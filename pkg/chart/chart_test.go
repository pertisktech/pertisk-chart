package chart

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"testing"
)

func TestExtractResourcesParsesTemplatesWithSourceComment(t *testing.T) {
	chartData := buildTestChart(t, map[string]string{
		"demo/Chart.yaml": "apiVersion: v2\nname: demo\nversion: 0.1.0\n",
		"demo/templates/deployment.yaml": `apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo
  namespace: default
`,
	})

	resources, err := ExtractResources(bytes.NewReader(chartData))
	if err != nil {
		t.Fatalf("ExtractResources returned error: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}

	resource := resources[0]
	if resource.Kind != "Deployment" {
		t.Fatalf("expected kind Deployment, got %q", resource.Kind)
	}
	if resource.Name != "demo" {
		t.Fatalf("expected name demo, got %q", resource.Name)
	}
	if resource.Namespace != "default" {
		t.Fatalf("expected namespace default, got %q", resource.Namespace)
	}
	if resource.APIVersion != "apps/v1" {
		t.Fatalf("expected apiVersion apps/v1, got %q", resource.APIVersion)
	}
}

func buildTestChart(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzipWriter)

	for name, content := range files {
		data := []byte(content)
		if err := tarWriter.WriteHeader(&tar.Header{
			Name: name,
			Mode: 0o644,
			Size: int64(len(data)),
		}); err != nil {
			t.Fatalf("failed to write tar header: %v", err)
		}
		if _, err := io.Copy(tarWriter, bytes.NewReader(data)); err != nil {
			t.Fatalf("failed to write tar content: %v", err)
		}
	}

	if err := tarWriter.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}

	return buf.Bytes()
}
