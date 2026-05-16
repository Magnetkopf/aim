#!/bin/bash
set -e

REPO="Magnetkopf/aim"
INSTALL_DIR="/usr/local/bin"
INSTALL_PATH="${INSTALL_DIR}/aim"

# Detect architecture (x86_64 comes before arm patterns)
detect_arch() {
    local arch=$(uname -m)
    case "$arch" in
        x86_64)     echo "amd64" ;;
        i386|i686)  echo "386" ;;
        aarch64)    echo "arm64" ;;
        armv7l|armv8l) echo "arm" ;;
        arm*)       echo "arm" ;;
        riscv64)    echo "riscv64" ;;
        *)
            echo "Error: Unsupported architecture: $arch" >&2
            exit 1
            ;;
    esac
}

# Detect OS
detect_os() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    if [ "$os" != "linux" ]; then
        echo "Error: This script only supports Linux" >&2
        exit 1
    fi
    echo "linux"
}

# Get latest nightly release info
get_latest_release() {
    local os=$(detect_os)
    local arch=$(detect_arch)

    curl -sL "https://api.github.com/repos/${REPO}/releases/tags/nightly" \
        | grep -o '"browser_download_url": "[^"]*aim-'${os}'-'${arch}'\.tar\.gz"' \
        | sed -n 's/.*"\([^"]*\)".*/\1/p'
}

# Check if running as root for system-wide install
check_permissions() {
    if [ "$(id -u)" -ne 0 ]; then
        echo "Warning: Installing to ${INSTALL_DIR} requires root privileges."
    fi
}

main() {
    local os=$(detect_os)
    local arch=$(detect_arch)
    local target="aim-${os}-${arch}.tar.gz"

    echo "Detected platform: ${os}-${arch}"
    echo "Downloading latest nightly release..."

    # Create temp directory
    local tmpdir=$(mktemp -d)
    trap "rm -rf $tmpdir" EXIT

    # Download the release
    local download_url=$(get_latest_release)
    if [ -z "$download_url" ]; then
        echo "Error: Could not find release for ${os}-${arch}" >&2
        exit 1
    fi

    echo "Downloading from: ${download_url}"
    curl -sL "$download_url" -o "${tmpdir}/${target}"

    # Extract
    echo "Extracting..."
    tar -xzf "${tmpdir}/${target}" -C "$tmpdir"

    # Install
    echo "Installing to ${INSTALL_PATH}..."
    mkdir -p "$INSTALL_DIR"
    mv "${tmpdir}/aim" "$INSTALL_PATH"
    chmod +x "$INSTALL_PATH"

    echo "Successfully installed aim to ${INSTALL_PATH}"

}

check_permissions
main "$@"
