# Pertisk Chart Repository

A modern Helm chart repository server built with Go, featuring both a web UI and REST API. Similar to ChartMuseum, this server provides a complete solution for hosting and managing Helm charts.

## Features

- 📦 **Chart Management**: Upload, browse, download, and delete Helm charts
- 🎨 **Modern Web UI**: Beautiful, responsive interface for managing charts
- 🔌 **REST API**: Full-featured API for programmatic access
- 📋 **Index Generation**: Automatic `index.yaml` generation for Helm compatibility
- 💾 **Local Storage**: File-based storage backend (easily extensible to cloud storage)
- 🔍 **Chart Discovery**: Browse all charts and versions through the UI
- ✅ **Helm Compatible**: Works seamlessly with Helm CLI

## Quick Start

### Prerequisites

- Go 1.21 or later
- Helm (for testing)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd pertisk-chart
```

2. Install dependencies:
```bash
go mod download
```

3. Build the server:
```bash
go build -o pertisk-chart ./cmd/server
```

4. Run the server:
```bash
./pertisk-chart --port=8080 --storage-local-rootdir=./chartstorage
```

Or run directly with Go:
```bash
go run ./cmd/server --port=8080 --storage-local-rootdir=./chartstorage
```

### Usage

1. **Access the Web UI**: Open http://localhost:8080 in your browser

2. **Add repository to Helm**:
```bash
helm repo add pertisk http://localhost:8080
helm repo update
```

3. **Upload a chart**:
```bash
# Via UI: Navigate to /charts and use the upload form
# Via API:
curl -X POST http://localhost:8080/api/charts \
  -F "chart=@mychart-1.0.0.tgz"
```

4. **Browse charts**: Visit http://localhost:8080/charts

5. **Search and install charts**:
```bash
helm search repo pertisk
helm install myapp pertisk/mychart
```

## API Endpoints

### Health Check
- `GET /api/health` - Server health status

### Charts
- `GET /api/charts` - List all charts
- `GET /api/charts/:name` - Get chart information
- `GET /api/charts/:name/:version` - Get specific chart version
- `POST /api/charts` - Upload a chart (multipart/form-data, field: `chart`)
- `DELETE /api/charts/:name/:version` - Delete a chart version

### Helm Repository
- `GET /index.yaml` - Helm repository index (used by `helm repo update`)

### Chart Download
- `GET /charts/:name/:version/:filename` - Download chart package

## Configuration

Command-line flags:

- `--port` - Port to listen on (default: "8080")
- `--storage` - Storage backend (default: "local")
- `--storage-local-rootdir` - Local storage root directory (default: "./chartstorage")
- `--enable-metrics` - Enable Prometheus metrics (default: false)
- `--debug` - Enable debug mode (default: false)

## Project Structure

```
pertisk-chart/
├── cmd/
│   └── server/
│       └── main.go          # Main entry point
├── pkg/
│   ├── api/
│   │   └── server.go        # API handlers and routes
│   ├── chart/
│   │   └── chart.go         # Chart parsing and utilities
│   └── storage/
│       └── storage.go       # Storage backend interface
├── web/
│   ├── templates/
│   │   ├── index.html      # Home page
│   │   └── charts.html     # Charts listing page
│   └── static/
│       ├── css/
│       │   └── main.css    # Styles
│       └── js/
│           └── main.js     # JavaScript
├── go.mod
└── README.md
```

## Development

### Running in Development Mode

```bash
go run ./cmd/server --debug --port=8080
```

### Building for Production

```bash
go build -ldflags="-s -w" -o pertisk-chart ./cmd/server
```

### Testing

Create a test chart:
```bash
helm create test-chart
helm package test-chart
```

Upload it:
```bash
curl -X POST http://localhost:8080/api/charts -F "chart=@test-chart-0.1.0.tgz"
```

## Features Similar to ChartMuseum

- ✅ Helm-compatible repository index
- ✅ Chart upload via API
- ✅ Chart download
- ✅ Web UI for chart management
- ✅ REST API
- ✅ Local storage backend
- ✅ Chart metadata parsing
- ✅ Multiple chart versions support

## Future Enhancements

- [ ] Cloud storage backends (S3, GCS, Azure)
- [ ] Authentication and authorization
- [ ] Multi-tenancy support
- [ ] Chart signing and verification
- [ ] Prometheus metrics
- [ ] Redis caching
- [ ] Chart search functionality
- [ ] API rate limiting

## License

Apache 2.0

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

