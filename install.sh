#!/bin/bash
set -e

INSTALL_DIR="/usr/local/bin"
BINARY_NAME="base"
DOWNLOAD_BASE_URL="${BASE_DOWNLOAD_URL:-https://github.com/igorkalen/base/releases/latest/download}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo ""
echo -e "${CYAN}╔════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║ B.A.S.E. Installer                         ║${NC}"
echo -e "${CYAN}║ Backend Automation & Scripting Environment ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════╝${NC}"
echo ""

detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux)   OS="linux" ;;
        darwin)  OS="darwin" ;;
        *)
            echo -e "${RED}Error: Unsupported operating system: $OS${NC}"
            exit 1
            ;;
    esac

    case "$ARCH" in
        x86_64)  ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        arm64)   ARCH="arm64" ;;
        *)
            echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac

    echo -e "${GREEN}Detected:${NC} $OS/$ARCH"
}

download_binary() {
    DOWNLOAD_URL="${DOWNLOAD_BASE_URL}/base-${OS}-${ARCH}"
    TMP_FILE=$(mktemp)

    echo -e "${YELLOW}Downloading B.A.S.E. from ${DOWNLOAD_URL}...${NC}"

    if command -v curl &> /dev/null; then
        curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE"
    elif command -v wget &> /dev/null; then
        wget -q "$DOWNLOAD_URL" -O "$TMP_FILE"
    else
        echo -e "${RED}Error: curl or wget is required${NC}"
        exit 1
    fi

    chmod +x "$TMP_FILE"
    echo -e "${GREEN}Download complete.${NC}"
}

install_binary() {
    echo -e "${YELLOW}Installing to ${INSTALL_DIR}/${BINARY_NAME}...${NC}"

    if [ -w "$INSTALL_DIR" ]; then
        mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        echo -e "${YELLOW}Requesting sudo access to install to ${INSTALL_DIR}...${NC}"
        sudo mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
        sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    echo -e "${GREEN}Installed successfully.${NC}"
}

verify_install() {
    if command -v base &> /dev/null; then
        VERSION=$(base --version 2>/dev/null || echo "unknown")
        echo ""
        echo -e "${GREEN}✓ B.A.S.E. installed successfully!${NC}"
        echo -e "  ${VERSION}"
        echo ""
        echo -e "  Get started:"
        echo -e "    ${CYAN}base help${NC}              Show all commands"
        echo -e "    ${CYAN}base new my-app${NC}        Create a new project"
        echo -e "    ${CYAN}base script.base${NC}       Run a script"
        echo -e "    ${CYAN}base${NC}                   Start interactive REPL"
        echo ""
        echo -e "  To uninstall:"
        echo -e "    ${CYAN}base uninstall${NC}"
        echo ""
    else
        echo -e "${RED}Warning: 'base' not found in PATH.${NC}"
        echo -e "Make sure ${INSTALL_DIR} is in your PATH."
        echo -e "Add this to your shell profile:"
        echo -e "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    fi
}

detect_platform
download_binary
install_binary
verify_install
