package version

// Version is the application version. Overridden at build time via:
//
//	-ldflags "-X github.com/pertisk-tech/pertisk-chart/pkg/version.Version=1.2.3"
var Version = "dev"
