%global debug_package %{nil}

Name:           pertisk-chart
Version:        %{?package_version}%{!?package_version:0.1.2}
Release:        %{?package_release}%{!?package_release:1}%{?dist}
Summary:        Helm chart repository server with web UI
License:        Apache-2.0
URL:            https://github.com/pertisk-tech/pertisk-chart
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  gcc
BuildRequires:  make
BuildRequires:  systemd-rpm-macros
Requires:       libcap
Requires:       shadow-utils
Requires(pre):  shadow-utils
Requires(post): systemd
Requires(preun): systemd
Requires(postun): systemd

%description
A Helm chart repository server with a web UI and REST API.
Provides chart upload, browse, download, and automatic index.yaml
generation for Helm compatibility.

%prep
%autosetup -n %{name}-%{version}

%build
export PATH=/usr/local/go/bin:/usr/bin:$PATH
export CGO_ENABLED=1
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org,direct
export GOFLAGS="-buildvcs=false"

go build -a -ldflags="-s -w" -tags "osusergo netgo" -o pertisk-chart ./cmd/server
go build -a -ldflags="-s -w" -tags "osusergo netgo" -o pertisk-chart-create-admin ./cmd/create-admin

%install
install -d %{buildroot}%{_bindir}
install -d %{buildroot}%{_libexecdir}
install -d %{buildroot}%{_sysconfdir}/%{name}
install -d %{buildroot}%{_unitdir}
install -d %{buildroot}%{_sharedstatedir}/%{name}/chartstorage
install -d %{buildroot}%{_localstatedir}/log/%{name}
install -d %{buildroot}%{_datadir}/%{name}

install -m 0755 pertisk-chart %{buildroot}%{_bindir}/%{name}
install -m 0755 pertisk-chart-create-admin %{buildroot}%{_bindir}/%{name}-create-admin
install -m 0755 packaging/pertisk-chart.sh %{buildroot}%{_libexecdir}/%{name}
install -m 0644 packaging/config.conf %{buildroot}%{_sysconfdir}/%{name}/config.conf
install -m 0644 packaging/pertisk-chart.service %{buildroot}%{_unitdir}/%{name}.service
cp -a web %{buildroot}%{_datadir}/%{name}/

%pre
if ! getent group pertisk-chart >/dev/null 2>&1; then
    groupadd --system pertisk-chart
fi
if ! getent passwd pertisk-chart >/dev/null 2>&1; then
    useradd --system --gid pertisk-chart --home-dir %{_sharedstatedir}/%{name} \
        --shell /sbin/nologin --comment "Pertisk Helm Chart Repository" pertisk-chart
fi

%post
%systemd_post pertisk-chart.service
if [ $1 -eq 1 ]; then
    chown -R pertisk-chart:pertisk-chart %{_sharedstatedir}/%{name} || true
    chown -R pertisk-chart:pertisk-chart %{_localstatedir}/log/%{name} || true
    chmod 750 %{_sharedstatedir}/%{name} || true
    chmod 750 %{_localstatedir}/log/%{name} || true
    mkdir -p %{_sharedstatedir}/%{name}/chartstorage
    chown pertisk-chart:pertisk-chart %{_sharedstatedir}/%{name}/chartstorage || true
    chmod 750 %{_sharedstatedir}/%{name}/chartstorage || true
fi
if command -v setcap >/dev/null 2>&1; then
    setcap 'cap_net_bind_service=+ep' %{_bindir}/%{name} || true
fi
if [ $1 -eq 1 ]; then
    echo ""
    echo "Pertisk Chart Repository has been installed successfully!"
    echo ""
    echo "IMPORTANT: Create an admin user before starting the service:"
    echo "  sudo pertisk-chart-create-admin \\"
    echo "    -username admin \\"
    echo "    -email admin@example.com \\"
    echo "    -password your-secure-password \\"
    echo "    -data-dir /var/lib/pertisk-chart \\"
    echo "    -db-type sqlite"
    echo ""
    echo "Then start the service:"
    echo "  sudo systemctl start pertisk-chart"
    echo "  sudo systemctl enable pertisk-chart"
    echo ""
    echo "Configuration: /etc/pertisk-chart/config.conf"
fi

%preun
%systemd_preun pertisk-chart.service

%postun
%systemd_postun pertisk-chart.service

%files
%doc README.md
%config(noreplace) %{_sysconfdir}/%{name}/config.conf
%{_bindir}/%{name}
%{_bindir}/%{name}-create-admin
%{_libexecdir}/%{name}
%{_unitdir}/%{name}.service
%{_datadir}/%{name}/
%dir %attr(0750,pertisk-chart,pertisk-chart) %{_sharedstatedir}/%{name}
%dir %attr(0750,pertisk-chart,pertisk-chart) %{_sharedstatedir}/%{name}/chartstorage
%dir %attr(0750,pertisk-chart,pertisk-chart) %{_localstatedir}/log/%{name}

%changelog
* Thu Aug 13 2026 Pertisk Team <dev@pertisk.tech> - 0.1.2-1
- Add AlmaLinux RPM packaging with systemd service
