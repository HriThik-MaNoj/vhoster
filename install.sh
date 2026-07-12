#!/bin/sh
set -e

REPO="HriThik-MaNoj/vhoster"
BIN="vhoster"
INSTALL_DIR="/usr/local/bin"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

# Resolve latest version from GitHub tags
VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name"' | head -1 | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
if [ -z "$VERSION" ]; then
	echo "${RED}Failed to determine latest version${NC}"
	exit 1
fi

if command -v go >/dev/null 2>&1; then
	echo "${GREEN}Go detected — installing via go install${NC}"
	go install "github.com/${REPO}@${VERSION}"
	GOPATH=$(go env GOPATH 2>/dev/null || echo "$HOME/go")
	if [ -f "${GOPATH}/bin/${BIN}" ]; then
		sudo cp "${GOPATH}/bin/${BIN}" "${INSTALL_DIR}/${BIN}"
		sudo chmod +x "${INSTALL_DIR}/${BIN}"
	else
		echo "${RED}Could not find binary after go install${NC}"
		exit 1
	fi
	echo "${GREEN}Installed ${BIN} ${VERSION} to ${INSTALL_DIR}/${BIN}${NC}"
	echo "Run it with: sudo ${BIN}"
	exit 0
fi

echo "Go not found — downloading prebuilt binary for ${VERSION}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
	x86_64|amd64)  ARCH="amd64" ;;
	aarch64|arm64) ARCH="arm64" ;;
	*)
		echo "${RED}Unsupported architecture: ${ARCH}${NC}"
		exit 1
		;;
esac

V=$(echo "$VERSION" | sed 's/^v//')
URL="https://github.com/${REPO}/releases/download/${VERSION}/${BIN}_${V}_${OS}_${ARCH}.tar.gz"

echo "Downloading ${URL} ..."
curl -fsSL "$URL" -o "/tmp/${BIN}.tar.gz"
tar xzf "/tmp/${BIN}.tar.gz" -C /tmp
sudo mv "/tmp/${BIN}" "${INSTALL_DIR}/${BIN}"
sudo chmod +x "${INSTALL_DIR}/${BIN}"
rm -f "/tmp/${BIN}.tar.gz"

echo "${GREEN}Installed ${BIN} ${VERSION} to ${INSTALL_DIR}/${BIN}${NC}"
echo "Run it with: sudo ${BIN}"
