#!/usr/bin/env bash

set -e -o pipefail

version=$(cat VERSION)
sed -i '' "s/const VERSION.*/const VERSION = \"${version}\"/" cmd/root.go &>/dev/null

git add cmd/root.go
