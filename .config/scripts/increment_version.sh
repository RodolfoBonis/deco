#!/bin/bash

# Script robusto para incremento de versão seguindo Semantic Versioning (SemVer)
# Detecta automaticamente o tipo de mudança baseado nos commits e incrementa a versão apropriada

set -euo pipefail  # Exit on error, undefined vars, pipe failures

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to validate version format
validate_version() {
    local version=$1
    if [[ ! $version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        log_error "Invalid version format: $version. Expected format: X.Y.Z"
        exit 1
    fi
}

# Function to parse version into components
parse_version() {
    local version=$1
    IFS='.' read -ra parts <<< "$version"
    echo "${parts[0]} ${parts[1]} ${parts[2]}"
}

# Function to detect version bump type from commit messages
detect_bump_type() {
    local commit_range=${1:-"HEAD~10..HEAD"}
    local bump_type="patch"  # Default to patch
    
    # Get commit messages from the specified range
    local commit_messages=$(git log --pretty=format:"%s" "$commit_range" 2>/dev/null || echo "")
    
    if [[ -z "$commit_messages" ]]; then
        echo "patch"
        return
    fi
    
    # Check for breaking changes (major version bump)
    if echo "$commit_messages" | grep -qE "(BREAKING CHANGE|!:|!\(|breaking change)"; then
        echo "major"
        return
    fi
    
    # Check for features (minor version bump)
    if echo "$commit_messages" | grep -qE "^(feat|feature):"; then
        echo "minor"
        return
    fi
    
    # Check for other conventional commit types that might indicate minor bump
    if echo "$commit_messages" | grep -qE "^(perf|refactor|style|docs):"; then
        echo "minor"
        return
    fi
    
    # Check for fixes (patch version bump)
    if echo "$commit_messages" | grep -qE "^(fix|bugfix|hotfix):"; then
        echo "patch"
        return
    fi
    
    # Check for chore, build, ci, test (patch version bump)
    if echo "$commit_messages" | grep -qE "^(chore|build|ci|test):"; then
        echo "patch"
        return
    fi
    
    echo "patch"
}

# Function to increment version based on bump type
increment_version() {
    local major=$1
    local minor=$2
    local patch=$3
    local bump_type=$4
    
    case $bump_type in
        "major")
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        "minor")
            minor=$((minor + 1))
            patch=0
            ;;
        "patch")
            patch=$((patch + 1))
            ;;
        *)
            log_error "Invalid bump type: $bump_type"
            exit 1
            ;;
    esac
    
    echo "$major.$minor.$patch"
}

# Function to check if version file exists and create if needed
ensure_version_file() {
    if [[ ! -f "version.txt" ]]; then
        log_warning "version.txt not found, creating with initial version 0.1.0"
        echo "0.1.0" > version.txt
    fi
}

# Function to get current git status
get_git_status() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log_error "Not in a git repository"
        exit 1
    fi
    
    # Check if working directory is clean
    if [[ -n $(git status --porcelain) ]]; then
        log_warning "Working directory is not clean. Uncommitted changes detected."
        log_info "Current git status:"
        git status --short
    fi
}

# Function to validate git configuration
validate_git_config() {
    local user_name=$(git config user.name 2>/dev/null || echo "")
    local user_email=$(git config user.email 2>/dev/null || echo "")
    
    if [[ -z "$user_name" || -z "$user_email" ]]; then
        log_warning "Git user configuration incomplete:"
        log_warning "  user.name: $user_name"
        log_warning "  user.email: $user_email"
        log_info "Consider setting git config user.name and user.email"
    fi
}

# Main execution
main() {
    log_info "Starting version increment process..."
    
    # Validate git repository
    get_git_status
    validate_git_config
    
    # Ensure version file exists
    ensure_version_file
    
    # Read current version
    local current_version=$(cat version.txt)
    log_info "Current version: $current_version"
    
    # Validate current version format
    validate_version "$current_version"
    
    # Parse version components
    read -r major minor patch <<< "$(parse_version "$current_version")"
    log_info "Version components: major=$major, minor=$minor, patch=$patch"
    
    # Detect bump type from recent commits
    local bump_type=$(detect_bump_type)
    log_info "Detected bump type: $bump_type"
    
    # Increment version
    local new_version=$(increment_version "$major" "$minor" "$patch" "$bump_type")
    
    # Validate new version
    validate_version "$new_version"
    
    # Write new version to file
    echo "$new_version" > version.txt
    
    log_success "Version incremented successfully!"
    log_info "Previous version: $current_version"
    log_info "New version: $new_version"
    log_info "Bump type: $bump_type"
    
    # Show git diff for version file
    if git diff version.txt > /dev/null 2>&1; then
        log_info "Version file changes:"
        git diff version.txt
    fi
}

# Handle script arguments
case "${1:-}" in
    "--help"|"-h")
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --major        Force major version bump"
        echo "  --minor        Force minor version bump"
        echo "  --patch        Force patch version bump"
        echo "  --range RANGE  Specify git commit range for analysis (default: HEAD~10..HEAD)"
        echo ""
        echo "Examples:"
        echo "  $0                    # Auto-detect bump type from recent commits"
        echo "  $0 --major            # Force major version bump"
        echo "  $0 --range HEAD~5..HEAD  # Analyze last 5 commits"
        echo ""
        echo "The script follows Semantic Versioning (SemVer) and analyzes commit messages"
        echo "to automatically determine the appropriate version bump:"
        echo "  - major: breaking changes (BREAKING CHANGE, !:)"
        echo "  - minor: new features (feat:, perf:, refactor:)"
        echo "  - patch: bug fixes (fix:, chore:, build:)"
        exit 0
        ;;
    "--major")
        # Force major bump
        ensure_version_file
        current_version=$(cat version.txt)
        validate_version "$current_version"
        read -r major minor patch <<< "$(parse_version "$current_version")"
        new_version=$(increment_version "$major" "$minor" "$patch" "major")
        echo "$new_version" > version.txt
        log_success "Forced major version bump: $current_version -> $new_version"
        exit 0
        ;;
    "--minor")
        # Force minor bump
        ensure_version_file
        current_version=$(cat version.txt)
        validate_version "$current_version"
        read -r major minor patch <<< "$(parse_version "$current_version")"
        new_version=$(increment_version "$major" "$minor" "$patch" "minor")
        echo "$new_version" > version.txt
        log_success "Forced minor version bump: $current_version -> $new_version"
        exit 0
        ;;
    "--patch")
        # Force patch bump
        ensure_version_file
        current_version=$(cat version.txt)
        validate_version "$current_version"
        read -r major minor patch <<< "$(parse_version "$current_version")"
        new_version=$(increment_version "$major" "$minor" "$patch" "patch")
        echo "$new_version" > version.txt
        log_success "Forced patch version bump: $current_version -> $new_version"
        exit 0
        ;;
    "--range")
        if [[ -z "${2:-}" ]]; then
            log_error "Range argument requires a value"
            exit 1
        fi
        # Override commit range for analysis
        COMMIT_RANGE="$2"
        main
        ;;
    "")
        # Default behavior - auto-detect
        main
        ;;
    *)
        log_error "Unknown option: $1"
        echo "Use --help for usage information"
        exit 1
        ;;
esac