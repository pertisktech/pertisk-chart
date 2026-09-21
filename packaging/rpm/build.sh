#!/usr/bin/env bash
# Build an AlmaLinux/RHEL RPM with rpmbuild.
# Usage: VERSION=0.1.2 RELEASE=1 ./packaging/rpm/build.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
RELEASE="${RELEASE:-1}"
VERSION="${VERSION:-}"

if [ -z "$VERSION" ]; then
    if command -v git >/dev/null 2>&1 && [ -d "$ROOT/.git" ]; then
        VERSION="$(git -C "$ROOT" describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || true)"
    fi
    VERSION="${VERSION:-0.1.2}"
fi

# RPM Version cannot contain hyphens
VERSION="${VERSION//-/.}"

TOPDIR="${TOPDIR:-$ROOT/build/rpm}"
TARNAME="pertisk-chart-${VERSION}"

echo "Building pertisk-chart RPM ${VERSION}-${RELEASE} for $(uname -m)"

rm -rf "$TOPDIR"
mkdir -p "$TOPDIR"/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS}

TMPDIR="$(mktemp -d)"
cleanup() { rm -rf "$TMPDIR"; }
trap cleanup EXIT

mkdir -p "$TMPDIR/$TARNAME"
if command -v rsync >/dev/null 2>&1; then
    rsync -a \
        --exclude '.git' \
        --exclude 'dist' \
        --exclude 'build' \
        --exclude 'tmp' \
        --exclude 'data' \
        --exclude 'chartstorage' \
        --exclude 'vendor' \
        --exclude 'pertisk-chart' \
        --exclude '*.rpm' \
        "$ROOT/" "$TMPDIR/$TARNAME/"
else
    tar -C "$ROOT" \
        --exclude '.git' \
        --exclude 'dist' \
        --exclude 'build' \
        --exclude 'tmp' \
        --exclude 'data' \
        --exclude 'chartstorage' \
        --exclude 'vendor' \
        --exclude 'pertisk-chart' \
        --exclude '*.rpm' \
        -cf - . | tar -C "$TMPDIR/$TARNAME" -xf -
fi

# Ensure UI assets look "new" to Last-Modified caches after every package build.
# Alma/rpmbuild can normalize mtimes; browsers then keep serving a stale SPA shell.
stamp_web_assets() {
    local web_root="$1"
    local ver="$2"
    local cache_token="${ver}-$(date -u +%Y%m%d%H%M%S)"
    [ -d "$web_root" ] || return 0

    if [ -f "$web_root/index.html" ]; then
        # Bust CSS/JS query strings so upgraded RPMs pull fresh assets.
        sed -i.bak -E \
            -e "s|(design\\.css\\?v=)[^\"']+|\\1${cache_token}|g" \
            -e "s|(app\\.js\\?v=)[^\"']+|\\1${cache_token}|g" \
            "$web_root/index.html"
        rm -f "$web_root/index.html.bak"

        # Bake version into the sidebar label as an immediate fallback.
        python3 - "$web_root/index.html" "$ver" <<'PY'
import pathlib, sys
path = pathlib.Path(sys.argv[1])
ver = sys.argv[2].lstrip("v")
text = path.read_text()
old = '<span class="brand-version" id="appVersionLabel" aria-label="App version"></span>'
new = f'<span class="brand-version" id="appVersionLabel" aria-label="App version">v{ver}</span>'
if old in text:
    path.write_text(text.replace(old, new, 1))
PY
    fi

    find "$web_root" -type f -exec touch -m {} +
    mkdir -p "$web_root/static"
    printf '%s\n' "$ver" > "$web_root/static/version.txt"
}

stamp_web_assets "$TMPDIR/$TARNAME/web" "$VERSION"

tar -C "$TMPDIR" -czf "$TOPDIR/SOURCES/${TARNAME}.tar.gz" "$TARNAME"
cp "$ROOT/packaging/rpm/pertisk-chart.spec" "$TOPDIR/SPECS/"

# Avoid reproducible-build mtime clamping so upgraded UI files get fresh Last-Modified.
unset SOURCE_DATE_EPOCH || true

rpmbuild -ba \
    --define "_topdir $TOPDIR" \
    --define "package_version $VERSION" \
    --define "package_release $RELEASE" \
    --define "source_date_epoch_from_changelog 0" \
    "$TOPDIR/SPECS/pertisk-chart.spec"

mkdir -p "$ROOT/dist"
find "$TOPDIR/RPMS" "$TOPDIR/SRPMS" -name '*.rpm' -exec cp {} "$ROOT/dist/" \;

echo ""
echo "RPM packages written to $ROOT/dist:"
ls -lh "$ROOT/dist"/*.rpm
