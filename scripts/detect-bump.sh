#!/usr/bin/env bash
# Print version=X.Y.Z, bump=major|minor|patch|none, skip=true|false for GITHUB_OUTPUT.
set -euo pipefail

override="${BUMP_OVERRIDE:-}"
last="$(git describe --tags --abbrev=0 2>/dev/null || true)"
if [[ -z "$last" ]]; then
	# First release: ship plugin.json's 0.1.0 as-is (do not invent 0.2.0).
	echo "version=0.1.0"
	echo "bump=none"
	echo "skip=false"
	exit 0
fi
if [[ "$override" == "none" ]]; then
	echo "version=${last#v}"
	echo "bump=none"
	echo "skip=true"
	exit 0
fi

range="${last}..HEAD"
if [[ -z "$(git log --oneline "$range" 2>/dev/null || true)" ]]; then
	echo "version=${last#v}"
	echo "bump=none"
	echo "skip=true"
	exit 0
fi

subjects="$(git log --format=%s "$range")"
bodies="$(git log --format=%B "$range")"
bump=patch
if [[ "$override" == "major" || "$override" == "minor" || "$override" == "patch" ]]; then
	bump="$override"
elif echo "$bodies" | grep -q 'BREAKING CHANGE:' || echo "$subjects" | grep -qE '^[a-z]+(\([^)]+\))?!:'; then
	bump=major
elif echo "$subjects" | grep -qE '^feat(\(|:|!)'; then
	bump=minor
fi

ver="${last#v}"
IFS=. read -r maj min pat <<<"$ver"
maj=${maj:-0}
min=${min:-0}
pat=${pat:-0}
case "$bump" in
major) maj=$((maj + 1)); min=0; pat=0 ;;
minor) min=$((min + 1)); pat=0 ;;
patch) pat=$((pat + 1)) ;;
esac

echo "version=${maj}.${min}.${pat}"
echo "bump=${bump}"
echo "skip=false"
