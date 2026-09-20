#!/usr/bin/env bash
# Download onnxruntime 1.19.2 headers + shared library for GOOS/GOARCH (native CI).
set -euo pipefail

ver="${ORT_VERSION:-1.19.2}"
goos="${GOOS:-}"
goarch="${GOARCH:-}"
if [[ -z "$goos" || -z "$goarch" ]]; then
	goos="$(uname -s | tr '[:upper:]' '[:lower:]')"
	case "$goos" in
	darwin) goos=darwin ;;
	linux) goos=linux ;;
	esac
	case "$(uname -m)" in
	x86_64) goarch=amd64 ;;
	arm64 | aarch64) goarch=arm64 ;;
	*) echo "unsupported arch $(uname -m)" >&2; exit 1 ;;
	esac
fi

case "${goos}_${goarch}" in
darwin_arm64) name="onnxruntime-osx-arm64-${ver}" ;;
darwin_amd64) name="onnxruntime-osx-x86_64-${ver}" ;;
linux_amd64) name="onnxruntime-linux-x64-${ver}" ;;
linux_arm64) name="onnxruntime-linux-aarch64-${ver}" ;;
*) echo "unsupported ${goos}/${goarch}" >&2; exit 1 ;;
esac

dest="${ORT_PREFIX:-${RUNNER_TEMP:-/tmp}/ort}"
url="https://github.com/microsoft/onnxruntime/releases/download/v${ver}/${name}.tgz"
mkdir -p "$dest"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
echo "fetch $url"
curl -fsSL -o "$tmp/ort.tgz" "$url"
tar -xzf "$tmp/ort.tgz" -C "$tmp"
src="$tmp/$name"
mkdir -p "$dest/include" "$dest/lib"
cp -R "$src/include/." "$dest/include/"
# Shared libs (and versioned .so.N) for compile-time link / runtime smoke.
find "$src/lib" -maxdepth 1 \( -name 'libonnxruntime*' -o -name 'onnxruntime.dll' \) -exec cp {} "$dest/lib/" \;
echo "ORT_PREFIX=$dest"
echo "headers=$(ls "$dest/include" | wc -l) libs=$(ls "$dest/lib" | wc -l)"
