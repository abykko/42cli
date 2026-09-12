#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# --- Visual Styling & Colors ---
BOLD='\033[1;37m'
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

print_step() {
    echo -e "\n${BLUE}➔${NC} ${BOLD}$1${NC}"
}

print_success() {
    echo -e "   ${GREEN}✓${NC} $1"
}

print_info() {
    echo -e "   ${YELLOW}ℹ${NC} $1"
}

print_error() {
    echo -e "   ${RED}✗${NC} $1"
}

# --- Welcome Header ---
echo -e "${BLUE}===============================================${NC}"
echo -e "${BOLD}         42TUI ENVIRONMENT SETUP               ${NC}"
echo -e "${BLUE}===============================================${NC}"

# Detect OS
OS_TYPE="$(uname -s)"

# --- Dependency Management ---
print_step "Checking system environment..."

if [ "$OS_TYPE" = "Darwin" ]; then
    print_info "macOS system detected."
    
    # Check for Homebrew
    if ! command -v brew &> /dev/null; then
        print_error "Homebrew is required on macOS."
        echo -e "     Please install it first from: ${BLUE}https://brew.sh${NC}"
        exit 1
    fi

    # Install Podman
    if command -v podman &> /dev/null; then
        print_success "Podman is already installed: $(podman --version)"
    else
        print_info "Installing Podman via Homebrew..."
        brew install podman
        print_info "Initializing Podman machine..."
        podman machine init || true
        podman machine start || true
        print_success "Podman setup complete."
    fi

    # Install Go
    if command -v go &> /dev/null; then
        print_success "Go is already installed: $(go version | awk '{print $3}')"
    else
        print_info "Installing Go via Homebrew..."
        brew install go
        print_success "Go installation complete."
    fi

elif [ "$OS_TYPE" = "Linux" ]; then
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        VARIANT=${VARIANT_ID:-""}
    else
        print_error "Could not detect Linux distribution."
        exit 1
    fi
    print_info "Linux distribution detected: ${OS}"

    # Check Podman
    if command -v podman &> /dev/null; then
        print_success "Podman is already installed: $(podman --version)"
    else
        print_info "Installing Podman..."
        case $OS in
            ubuntu|debian)
                sudo apt update && sudo apt install -y podman
                ;;
            fedora)
                if [ "$VARIANT" = "silverblue" ]; then
                    print_error "Podman should be preinstalled on Silverblue."
                    exit 1
                else
                    sudo dnf install -y podman
                fi
                ;;
        esac
        print_success "Podman installation complete."
    fi

    # Check Go
    if command -v go &> /dev/null; then
        print_success "Go is already installed: $(go version | awk '{print $3}')"
    else
        print_info "Installing Go..."
        case $OS in
            ubuntu|debian)
                sudo apt update && sudo apt install -y golang
                ;;
            fedora)
                if [ "$VARIANT" = "silverblue" ]; then
                    print_info "Fedora Silverblue detected."
                    echo "     Use Toolbx to set up a dev environment:"
                    echo "     toolbox create && toolbox enter"
                    exit 0
                else
                    sudo dnf install -y golang
                fi
                ;;
            *)
                print_error "Distribution '$OS' is not supported by this automation."
                exit 1
                ;;
        esac
        print_success "Go installation complete."
    fi
else
    print_error "Unsupported Operating System: $OS_TYPE"
    exit 1
fi

echo -e "\n${GREEN}===============================================${NC}"
print_success "All dependencies are ready!"
echo -e "${GREEN}===============================================${NC}"

# --- Build and Run Steps ---

print_step "Downloading Go modules..."
go mod download
print_success "Modules downloaded successfully."

print_step "Building application (go build -o 42tui)..."
go build -o 42tui .
print_success "Build complete! Binary generated: ./42tui"

# --- Prompt to Run ---
echo -e "\n${BLUE}===============================================${NC}"
read -p " Do you want to run 42tui now? (Y/n): " response
echo -e "${BLUE}===============================================${NC}"

# Convert response to lowercase and target the first letter
response=$(echo "$response" | tr '[:upper:]' '[:lower:]')

# If response is empty (user hit Enter) or starts with 'y'
if [[ -z "$response" || "$response" == "y"* ]]; then
    echo -e "\n${GREEN}➔ Launching application...${NC}\n"
    ./42tui
else
    echo -e "\n${YELLOW}Setup finished!${NC} You can run the application later using: ${BOLD}./42tui${NC}\n"
fi