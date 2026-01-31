#!/bin/bash
set -e

REPO="12458/exa-cli"
BINARY_NAME="exa"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *) error "Unsupported operating system: $(uname -s)" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        armv7l|armv6l) echo "armv7" ;;
        *) error "Unsupported architecture: $(uname -m)" ;;
    esac
}

# Get latest release version from GitHub
get_latest_version() {
    local version
    if command -v curl &> /dev/null; then
        version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget &> /dev/null; then
        version=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        error "Neither curl nor wget found. Please install one of them."
    fi

    if [ -z "$version" ]; then
        error "Failed to get latest version from GitHub"
    fi
    echo "$version"
}

# Download file
download() {
    local url="$1"
    local output="$2"

    if command -v curl &> /dev/null; then
        curl -fsSL "$url" -o "$output"
    elif command -v wget &> /dev/null; then
        wget -q "$url" -O "$output"
    else
        error "Neither curl nor wget found. Please install one of them."
    fi
}

# Determine install directory
get_install_dir() {
    if [ -w "/usr/local/bin" ]; then
        echo "/usr/local/bin"
    elif [ -d "$HOME/.local/bin" ]; then
        echo "$HOME/.local/bin"
    else
        mkdir -p "$HOME/.local/bin"
        echo "$HOME/.local/bin"
    fi
}

# Install shell completions
install_completions() {
    local binary="$1"
    local installed=0

    # Bash completion
    if command -v bash &> /dev/null; then
        local bash_completion_dir=""
        if [ -d "/etc/bash_completion.d" ] && [ -w "/etc/bash_completion.d" ]; then
            bash_completion_dir="/etc/bash_completion.d"
        elif [ -d "/usr/local/etc/bash_completion.d" ] && [ -w "/usr/local/etc/bash_completion.d" ]; then
            bash_completion_dir="/usr/local/etc/bash_completion.d"
        elif [ -d "$HOME/.local/share/bash-completion/completions" ]; then
            bash_completion_dir="$HOME/.local/share/bash-completion/completions"
        elif [ -d "$HOME/.bash_completion.d" ]; then
            bash_completion_dir="$HOME/.bash_completion.d"
        else
            mkdir -p "$HOME/.local/share/bash-completion/completions"
            bash_completion_dir="$HOME/.local/share/bash-completion/completions"
        fi

        if [ -n "$bash_completion_dir" ]; then
            "$binary" completion bash > "$bash_completion_dir/exa"
            info "Bash completion installed to $bash_completion_dir/exa"
            installed=1
        fi
    fi

    # Zsh completion
    if command -v zsh &> /dev/null; then
        local zsh_completion_dir=""
        if [ -d "/usr/local/share/zsh/site-functions" ] && [ -w "/usr/local/share/zsh/site-functions" ]; then
            zsh_completion_dir="/usr/local/share/zsh/site-functions"
        elif [ -d "$HOME/.zsh/completions" ]; then
            zsh_completion_dir="$HOME/.zsh/completions"
        else
            mkdir -p "$HOME/.zsh/completions"
            zsh_completion_dir="$HOME/.zsh/completions"
        fi

        if [ -n "$zsh_completion_dir" ]; then
            "$binary" completion zsh > "$zsh_completion_dir/_exa"
            info "Zsh completion installed to $zsh_completion_dir/_exa"
            if [ "$zsh_completion_dir" = "$HOME/.zsh/completions" ]; then
                warn "Add to your ~/.zshrc: fpath=(~/.zsh/completions \$fpath)"
            fi
            installed=1
        fi
    fi

    # Fish completion
    if command -v fish &> /dev/null; then
        local fish_completion_dir=""
        if [ -d "$HOME/.config/fish/completions" ]; then
            fish_completion_dir="$HOME/.config/fish/completions"
        elif [ -d "/usr/local/share/fish/vendor_completions.d" ] && [ -w "/usr/local/share/fish/vendor_completions.d" ]; then
            fish_completion_dir="/usr/local/share/fish/vendor_completions.d"
        else
            mkdir -p "$HOME/.config/fish/completions"
            fish_completion_dir="$HOME/.config/fish/completions"
        fi

        if [ -n "$fish_completion_dir" ]; then
            "$binary" completion fish > "$fish_completion_dir/exa.fish"
            info "Fish completion installed to $fish_completion_dir/exa.fish"
            installed=1
        fi
    fi

    if [ "$installed" -eq 0 ]; then
        warn "No supported shells (bash, zsh, fish) found. Skipping completion installation."
    fi
}

main() {
    info "Installing exa CLI..."

    OS=$(detect_os)
    ARCH=$(detect_arch)
    VERSION=$(get_latest_version)

    info "Detected OS: $OS, Arch: $ARCH"
    info "Latest version: $VERSION"

    # Construct download URL
    FILENAME="${BINARY_NAME}_${OS}_${ARCH}"
    if [ "$OS" = "windows" ]; then
        FILENAME="${FILENAME}.exe"
    fi
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

    info "Downloading from: $DOWNLOAD_URL"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    # Download binary
    TMP_FILE="$TMP_DIR/$BINARY_NAME"
    download "$DOWNLOAD_URL" "$TMP_FILE"

    # Make executable
    chmod +x "$TMP_FILE"

    # Determine install location
    INSTALL_DIR=$(get_install_dir)
    INSTALL_PATH="$INSTALL_DIR/$BINARY_NAME"

    # Install
    if [ -w "$INSTALL_DIR" ]; then
        mv "$TMP_FILE" "$INSTALL_PATH"
    else
        info "Requires sudo to install to $INSTALL_DIR"
        sudo mv "$TMP_FILE" "$INSTALL_PATH"
    fi

    info "Installed $BINARY_NAME to $INSTALL_PATH"

    # Install shell completions
    info "Installing shell completions..."
    install_completions "$INSTALL_PATH"

    # Verify installation
    if command -v "$BINARY_NAME" &> /dev/null; then
        info "Installation successful!"
        echo ""
        info "Run 'exa --help' to get started"
    else
        warn "Installation complete, but '$BINARY_NAME' is not in PATH."
        warn "Add $INSTALL_DIR to your PATH:"
        echo ""
        echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
        echo ""
        info "Or run directly: $INSTALL_PATH --help"
    fi

    echo ""
    info "Restart your shell or source your shell config to enable completions."
}

main "$@"
