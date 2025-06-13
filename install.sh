#!/bin/sh
set -e

# Colors
GREEN='\\033[0;32m'
YELLOW='\\033[1;33m'
CYAN='\\033[0;36m'
NC='\\033[0m' # No Color

# This script handles the installation of 'dops' by detecting the OS
# and architecture and downloading the appropriate binary.

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
BINARY_NAME=""

REPO="Mikescher/better-docker-ps"
INSTALL_DIR="/usr/local/bin"
BINARY_PATH="${INSTALL_DIR}/dops"

echo "${GREEN}Detecting platform: ${OS}/${ARCH}...${NC}"

if [ "$OS" = "linux" ]; then
    if [ "$ARCH" = "x86_64" ]; then
        BINARY_NAME="dops_linux-amd64-static"
    elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
        BINARY_NAME="dops_linux-arm64-static"
    fi
elif [ "$OS" = "darwin" ]; then
    if [ "$ARCH" = "arm64" ]; then
        BINARY_NAME="dops_macos-arm64"
    elif [ "$ARCH" = "x86_64" ]; then
        echo "${YELLOW}Error: Intel-based Macs are not supported.${NC}"
        exit 1
    fi
fi

if [ -z "$BINARY_NAME" ]; then
    echo "${YELLOW}Error: Unsupported OS or Architecture: ${OS}/${ARCH}${NC}"
    exit 1
fi

DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${BINARY_NAME}"

echo "${GREEN}Downloading 'dops' from ${DOWNLOAD_URL}...${NC}"

# Use curl or wget to download the binary to the installation path
# The script should be run with sudo, so we don't need sudo here.
if command -v curl >/dev/null 2>&1; then
    curl -sSL "${DOWNLOAD_URL}" -o "${BINARY_PATH}"
else
    wget -qO "${BINARY_PATH}" "${DOWNLOAD_URL}"
fi

echo "${GREEN}Setting execute permissions on ${BINARY_PATH}...${NC}"
chmod +x "${BINARY_PATH}"

echo "${GREEN}'dops' installed successfully to ${BINARY_PATH}${NC}"
echo "${GREEN}You can now run 'dops' from your terminal.${NC}"

echo ""
echo "${YELLOW}Optional: To use 'dops' as a drop-in replacement for 'docker ps',${NC}"
echo "${YELLOW}add the following function to your shell configuration file (e.g., ~/.zshrc, ~/.bashrc):${NC}"
echo ""
cat << EOF
${CYAN}docker() {
  case \\$1 in
    ps)
      shift
      command dops "\\$@"
      ;;
    *)
      command docker "\\$@";;
  esac
}${NC}
EOF
echo "" 