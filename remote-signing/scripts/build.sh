#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🔨 Building Harlequin Remote Signing Library${NC}"

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo -e "${YELLOW}📁 Project root: $PROJECT_ROOT${NC}"

# Check if frontend directory exists
if [ ! -d "$PROJECT_ROOT/frontend" ]; then
    echo -e "${RED}❌ Frontend directory not found at $PROJECT_ROOT/frontend${NC}"
    exit 1
fi

# Build frontend
echo -e "${YELLOW}🏗️  Building frontend...${NC}"
cd "$PROJECT_ROOT/frontend"

# Check if node_modules exists, if not install dependencies
if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}📦 Installing frontend dependencies...${NC}"
    yarn install
fi

# Build the frontend
echo -e "${YELLOW}🔨 Building frontend assets...${NC}"
yarn build

# Check if build was successful
if [ ! -d "dist" ]; then
    echo -e "${RED}❌ Frontend build failed - dist directory not found${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Frontend built successfully${NC}"

# Go back to project root
cd "$PROJECT_ROOT"

# Build server package
echo -e "${YELLOW}🔨 Building server package...${NC}"
go build ./server/...

echo -e "${GREEN}✅ Server package built successfully${NC}"

# Build examples if requested
if [ "$1" = "--examples" ]; then
    echo -e "${YELLOW}🔨 Building examples...${NC}"
    cd example

    # Build each example with .build extension
    for example in *.go; do
        if [ -f "$example" ]; then
            name=$(basename "$example" .go)
            echo -e "${YELLOW}  Building $name.build...${NC}"
            go build -tags example -o "${name}.build" "$example"
        fi
    done

    echo -e "${GREEN}✅ Examples built successfully${NC}"
fi

echo -e "${GREEN}🎉 Build completed successfully!${NC}"
echo -e "${YELLOW}💡 To run examples: cd example && go run -tags example simple_upload.go${NC}"
