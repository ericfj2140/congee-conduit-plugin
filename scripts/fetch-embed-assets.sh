#!/usr/bin/env bash
# Download sha256-pinned MiniLM ONNX + onnxruntime shared library into the plugin package.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
  x86_64) goarch=amd64 ;;
  arm64|aarch64) goarch=arm64 ;;
  *) echo "unsupported arch: $arch" >&2; exit 1 ;;
esac
case "$os" in
  darwin) goos=darwin ;;
  linux) goos=linux ;;
  *) echo "unsupported os: $os" >&2; exit 1 ;;
esac

# sentence-transformers/all-MiniLM-L6-v2 ONNX (384-d). Update sha when bumping the pin.
MODEL_URL="${MODEL_URL:-https://huggingface.co/sentence-transformers/all-MiniLM-L6-v2/resolve/main/onnx/model.onnx}"
MODEL_SHA256="${MODEL_SHA256:-}"

# onnxruntime 1.19.2 GPU-less. Override ORT_VERSION / ORT_SHA256 when pinning a new release.
ORT_VERSION="${ORT_VERSION:-1.19.2}"
ORT_SHA256="${ORT_SHA256:-}"

models_dir="$root/models"
lib_dir="$root/lib/${goos}_${goarch}"
mkdir -p "$models_dir" "$lib_dir"

download() {
  local url="$1" dest="$2" sha="$3"
  echo "fetch $url"
  curl -fsSL -o "$dest" "$url"
  local got
  got="$(shasum -a 256 "$dest" | awk '{print $1}')"
  echo "sha256 $got  $(basename "$dest")"
  if [[ -n "$sha" && "$got" != "$sha" ]]; then
    echo "sha256 mismatch for $(basename "$dest"): got $got want $sha" >&2
    rm -f "$dest"
    exit 1
  fi
}

download "$MODEL_URL" "$models_dir/minilm.onnx" "$MODEL_SHA256"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
if [[ "$goos" == "darwin" ]]; then
  ort_name="onnxruntime-osx-arm64-${ORT_VERSION}"
  if [[ "$goarch" == "amd64" ]]; then
    ort_name="onnxruntime-osx-x86_64-${ORT_VERSION}"
  fi
  url="https://github.com/microsoft/onnxruntime/releases/download/v${ORT_VERSION}/${ort_name}.tgz"
  download "$url" "$tmp/ort.tgz" "$ORT_SHA256"
  tar -xzf "$tmp/ort.tgz" -C "$tmp"
  cp "$tmp/${ort_name}/lib/libonnxruntime."*.dylib "$lib_dir/" 2>/dev/null || cp "$tmp/${ort_name}/lib/libonnxruntime.dylib" "$lib_dir/"
else
  ort_name="onnxruntime-linux-x64-${ORT_VERSION}"
  if [[ "$goarch" == "arm64" ]]; then
    ort_name="onnxruntime-linux-aarch64-${ORT_VERSION}"
  fi
  url="https://github.com/microsoft/onnxruntime/releases/download/v${ORT_VERSION}/${ort_name}.tgz"
  download "$url" "$tmp/ort.tgz" "$ORT_SHA256"
  tar -xzf "$tmp/ort.tgz" -C "$tmp"
  cp "$tmp/${ort_name}/lib/libonnxruntime.so"* "$lib_dir/"
fi

echo "wrote $models_dir/minilm.onnx"
echo "wrote $lib_dir"
echo "Set MODEL_SHA256 and ORT_SHA256 in CI once to pin."
