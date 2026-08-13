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
./pertisk-chart --port=7080 --storage-local-rootdir=./chartstorage
```

Or run directly with Go:
```bash
go run ./cmd/server --port=7080 --storage-local-rootdir=./chartstorage
```

### Usage

1. **Access the Web UI**: Open http://localhost:7080 in your browser

2. **Add repository to Helm**:
```bash
helm repo add pertisk http://localhost:7080
helm repo update
```

3. **Upload a chart**:
```bash
# Via UI: Navigate to /charts and use the upload form
# Via API:
curl -X POST http://localhost:7080/api/charts \
  -F "chart=@mychart-1.0.0.tgz"
```

4. **Browse charts**: Visit http://localhost:7080/charts

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

- `--port` - Port to listen on (default: "7080")
- `--enable-http3` - Enable HTTP/3 support (requires TLS certificates)
- `--tls-cert` - Path to TLS certificate file (required for HTTP/3)
- `--tls-key` - Path to TLS private key file (required for HTTP/3)
- `--enable-zstd` - Enable zstd compression (default: true)
- `--storage` - Storage backend (default: "local")
- `--storage-local-rootdir` - Local storage root directory (default: "./chartstorage")
- `--data-dir` - Data directory for user storage (default: "./data")
- `--db-type` - Database type: sqlite, postgres, or file (default: "sqlite")
- `--db-dsn` - Database connection string (DSN). For SQLite: file path (default: ./data/users.db). For PostgreSQL: connection string
- `--jwt-secret` - JWT secret key (or set JWT_SECRET env var)
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
go run ./cmd/server --debug --port=7080
```

### Hot Reload with Air

For a better development experience with automatic reloading on file changes:

1. **Install Air** (if not already installed):
```bash
make install-air
# or
go install github.com/air-verse/air@latest
```

2. **Run with hot reload**:
```bash
make run-dev
# or
air
```

Air will automatically rebuild and restart the server when you modify any `.go` or `.html` files. The configuration is in `.air.toml`.

### Building for Production

```bash
go build -ldflags="-s -w" -o pertisk-chart ./cmd/server
```

## AlmaLinux RPM

Build an installable RPM for AlmaLinux 9 (el9) using Docker:

```bash
make rpm
```

Optional overrides:

```bash
make rpm VERSION=0.1.2 RELEASE=1 ALMA_VERSION=9
# AlmaLinux 8 or 10:
make rpm ALMA_VERSION=8
make rpm ALMA_VERSION=10
```

On an AlmaLinux/RHEL host with Go 1.21+ and `rpm-build` installed, you can build without Docker:

```bash
sudo dnf install -y rpm-build rpmdevtools gcc make git tar gzip rsync systemd-rpm-macros
make rpm-native
```

Packages are written to `dist/`:

- `pertisk-chart-<version>-1.el9.x86_64.rpm` — binary package
- `pertisk-chart-<version>-1.el9.src.rpm` — source package

### Install on AlmaLinux

```bash
sudo dnf install -y dist/pertisk-chart-*.el9.x86_64.rpm
# or
sudo rpm -Uvh dist/pertisk-chart-*.el9.x86_64.rpm
```

From this repo on an AlmaLinux host after `make rpm` or `make rpm-native`:

```bash
make rpm-install
```

Create an admin user, then start the service:

```bash
sudo pertisk-chart-create-admin \
  -username admin \
  -email admin@example.com \
  -password your-secure-password \
  -data-dir /var/lib/pertisk-chart \
  -db-type sqlite

sudo systemctl enable --now pertisk-chart
sudo systemctl status pertisk-chart
```

The service listens on port 7080. Edit `/etc/pertisk-chart/config.conf` and run `sudo systemctl restart pertisk-chart` to apply changes.

| Path | Purpose |
| --- | --- |
| `/usr/bin/pertisk-chart` | Server binary |
| `/usr/bin/pertisk-chart-create-admin` | Admin bootstrap CLI |
| `/etc/pertisk-chart/config.conf` | Service configuration |
| `/var/lib/pertisk-chart` | Data and SQLite database |
| `/var/lib/pertisk-chart/chartstorage` | Uploaded charts |
| `/usr/share/pertisk-chart/web` | Web UI assets |

### Testing

Create a test chart:
```bash
helm create test-chart
helm package test-chart
```

Upload it:
```bash
curl -X POST http://localhost:7080/api/charts -F "chart=@test-chart-0.1.0.tgz"
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

## HTTP/3 and Compression Support

### HTTP/3 Support

The server supports HTTP/3 (QUIC) for improved performance and lower latency. To enable HTTP/3:

```bash
./pertisk-chart \
  --enable-http3 \
  --tls-cert=/path/to/cert.pem \
  --tls-key=/path/to/key.pem \
  --port=7080
```

**Note:** HTTP/3 requires TLS certificates. The server will fall back to HTTP/1.1 and HTTP/2 if HTTP/3 is not enabled or certificates are not provided.

### Zstandard (zstd) Compression

Zstd compression is enabled by default and automatically compresses responses when clients support it. The server supports multiple compression algorithms in order of preference:

1. zstd (Zstandard) - Best compression ratio and speed
2. brotli - Good compression ratio
3. gzip - Widely supported
4. deflate - Fallback option

Compression is automatically negotiated based on the client's `Accept-Encoding` header.

To disable compression:
```bash
./pertisk-chart --enable-zstd=false
```

## Authentication and Admin Setup

### Creating an Admin User

To access admin features (domain configuration, user management), you need to create an admin user. You can do this using the provided CLI tool:

**Using Make:**
```bash
make create-admin USERNAME=admin EMAIL=admin@example.com PASSWORD=your-secure-password
```

**Or directly with Go:**
```bash
go run cmd/create-admin/main.go \
  -username admin \
  -email admin@example.com \
  -password your-secure-password \
  -db-type sqlite \
  -data-dir ./data
```

**For PostgreSQL:**
```bash
go run cmd/create-admin/main.go \
  -username admin \
  -email admin@example.com \
  -password your-secure-password \
  -db-type postgres \
  -db-dsn "host=localhost user=postgres password=secret dbname=pertisk port=5432 sslmode=disable"
```

**Note:** If the user already exists, the command will promote them to admin instead of creating a new user.

### User Registration

Regular users can register through the web UI by clicking "Register" in the navigation bar. Regular users can:
- Upload charts
- View and manage their own charts
- Download charts

Only admin users can:
- Access admin configuration panel
- Configure domain settings
- Manage all users
- Set admin privileges

## Database Configuration

The server supports multiple database backends for user management:

### SQLite (Default)

SQLite is the default database and requires no additional setup:

```bash
./pertisk-chart --db-type=sqlite --db-dsn=./data/users.db
```

Or use the default location:
```bash
./pertisk-chart  # Uses ./data/users.db by default
```

### PostgreSQL

For PostgreSQL, provide a connection string:

```bash
./pertisk-chart \
  --db-type=postgres \
  --db-dsn="host=localhost user=postgres password=secret dbname=pertisk port=5432 sslmode=disable"
```

Or use the `DATABASE_URL` environment variable:
```bash
export DATABASE_URL="postgres://user:password@localhost:5432/pertisk?sslmode=disable"
./pertisk-chart --db-type=postgres
```

### File-based (Legacy)

For backward compatibility, file-based storage is still supported:

```bash
./pertisk-chart --db-type=file --data-dir=./data
```

## Future Enhancements

- [ ] Cloud storage backends (S3, GCS, Azure)
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

