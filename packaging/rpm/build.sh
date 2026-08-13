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

tar -C "$TMPDIR" -czf "$TOPDIR/SOURCES/${TARNAME}.tar.gz" "$TARNAME"
cp "$ROOT/packaging/rpm/pertisk-chart.spec" "$TOPDIR/SPECS/"

rpmbuild -ba \
    --define "_topdir $TOPDIR" \
    --define "package_version $VERSION" \
    --define "package_release $RELEASE" \
    "$TOPDIR/SPECS/pertisk-chart.spec"

mkdir -p "$ROOT/dist"
find "$TOPDIR/RPMS" "$TOPDIR/SRPMS" -name '*.rpm' -exec cp {} "$ROOT/dist/" \;

echo ""
echo "RPM packages written to $ROOT/dist:"
ls -lh "$ROOT/dist"/*.rpm
