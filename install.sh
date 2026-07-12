#!/bin/sh
set -e

REPO="HriThik-MaNoj/vhoster"
BIN="vhoster"
INSTALL_DIR="/usr/local/bin"

RED=$(printf '\033[0;31m')
GREEN=$(printf '\033[0;32m')
NC=$(printf '\033[0m')

VERSION=$(curl -sI "https://github.com/${REPO}/releases/latest" 2>/dev/null \
	| grep -i '^location:' \
	| sed -E 's|.*/tag/([^/[:space:]]+).*|\1|' | tr -d '\r')
if [ -z "$VERSION" ]; then
	printf "%sFailed to determine latest version%s\n" "$RED" "$NC"
	exit 1
fi

printf "%sInstalling vhoster %s ...%s\n" "$GREEN" "$VERSION" "$NC"

if command -v go >/dev/null 2>&1; then
	printf "%sGo detected — installing via go install%s\n" "$GREEN" "$NC"
	go install "github.com/${REPO}@${VERSION}"
	GOPATH=$(go env GOPATH 2>/dev/null || echo "$HOME/go")
	if [ -f "${GOPATH}/bin/${BIN}" ]; then
		sudo cp "${GOPATH}/bin/${BIN}" "${INSTALL_DIR}/${BIN}"
		sudo chmod +x "${INSTALL_DIR}/${BIN}"
	else
		printf "%sCould not find binary after go install%s\n" "$RED" "$NC"
		exit 1
	fi
	printf "%sInstalled to %s/%s%s\n" "$GREEN" "$INSTALL_DIR" "$BIN" "$NC"
	printf "Run it with: sudo %s\n" "$BIN"
	exit 0
fi

printf "%sGo not found — downloading prebuilt binary%s\n" "$GREEN" "$NC"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
	x86_64|amd64)  ARCH="amd64" ;;
	aarch64|arm64) ARCH="arm64" ;;
	*)
		printf "%sUnsupported architecture: %s%s\n" "$RED" "$ARCH" "$NC"
		exit 1
		;;
esac

V_NO_PREFIX=$(echo "$VERSION" | sed 's/^v//')
URL="https://github.com/${REPO}/releases/download/${VERSION}/${BIN}_${V_NO_PREFIX}_${OS}_${ARCH}.tar.gz"

printf "Downloading %s ...\n" "$URL"
curl -fsSL "$URL" -o "/tmp/${BIN}.tar.gz"
tar xzf "/tmp/${BIN}.tar.gz" -C /tmp
sudo mv "/tmp/${BIN}" "${INSTALL_DIR}/${BIN}"
sudo chmod +x "${INSTALL_DIR}/${BIN}"
rm -f "/tmp/${BIN}.tar.gz"

printf "%sInstalled to %s/%s%s\n" "$GREEN" "$INSTALL_DIR" "$BIN" "$NC"
printf "Run it with: sudo %s\n" "$BIN"
