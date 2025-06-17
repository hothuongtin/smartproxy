#!/bin/bash

# SmartProxy Release Script
# This script builds releases for all platforms and optionally creates a GitHub release

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get version from git
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
if [[ $VERSION == *"dirty"* ]]; then
    echo -e "${YELLOW}Warning: Working directory has uncommitted changes${NC}"
fi

echo -e "${GREEN}Building SmartProxy releases for version: ${VERSION}${NC}"

# Build all releases
echo "Building releases for all platforms..."
make release

# List generated files
echo -e "\n${GREEN}Generated release files:${NC}"
ls -lh release/

# Calculate checksums
echo -e "\n${GREEN}Calculating checksums...${NC}"
cd release
shasum -a 256 *.tar.gz *.zip > checksums.txt 2>/dev/null || sha256sum *.tar.gz *.zip > checksums.txt
cat checksums.txt
cd ..

# Check if we should create a GitHub release
if [[ "$1" == "--github" ]]; then
    # Check if gh is installed
    if ! command -v gh &> /dev/null; then
        echo -e "${RED}GitHub CLI (gh) is not installed. Please install it first:${NC}"
        echo "brew install gh"
        exit 1
    fi

    # Check if we're on a tag
    if ! git describe --exact-match --tags HEAD 2>/dev/null; then
        echo -e "${YELLOW}Not on a tagged commit. Create a tag first:${NC}"
        echo "git tag -a v1.0.0 -m 'Release v1.0.0'"
        echo "git push origin v1.0.0"
        exit 1
    fi

    TAG=$(git describe --exact-match --tags HEAD)
    
    echo -e "\n${GREEN}Creating GitHub release for tag: ${TAG}${NC}"
    
    # Create release notes
    NOTES="## SmartProxy ${TAG}

### Release Assets
- Linux AMD64: \`smartproxy-${VERSION}-linux-amd64.tar.gz\`
- Linux ARM64: \`smartproxy-${VERSION}-linux-arm64.tar.gz\`
- macOS AMD64: \`smartproxy-${VERSION}-darwin-amd64.tar.gz\`
- macOS ARM64: \`smartproxy-${VERSION}-darwin-arm64.tar.gz\`
- Windows AMD64: \`smartproxy-${VERSION}-windows-amd64.zip\`

### Installation
\`\`\`bash
# Download and extract for your platform
tar xzf smartproxy-${VERSION}-<platform>.tar.gz

# Run the proxy
./smartproxy
\`\`\`

### Checksums
\`\`\`
$(cat release/checksums.txt)
\`\`\`
"

    # Create GitHub release
    gh release create "${TAG}" \
        --title "SmartProxy ${TAG}" \
        --notes "${NOTES}" \
        release/*.tar.gz \
        release/*.zip \
        release/checksums.txt
        
    echo -e "${GREEN}GitHub release created successfully!${NC}"
else
    echo -e "\n${YELLOW}To create a GitHub release, run:${NC}"
    echo "./scripts/release.sh --github"
fi