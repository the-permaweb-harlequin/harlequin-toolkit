#!/bin/bash
# Harlequin CLI Template Installation Script for Go

set -e

PROJECT_NAME="$1"
if [ -z "$PROJECT_NAME" ]; then
    echo "Usage: $0 <project-name>"
    exit 1
fi

echo "🎭 Creating Go AO process: $PROJECT_NAME"

# Replace template variables
find . -type f -name "*.md" -o -name "*.json" -o -name "*.go" -o -name "go.mod" | xargs sed -i.bak "s/r/$PROJECT_NAME/g"
find . -name "*.bak" -delete

echo "✅ Template prepared successfully!"
echo ""
echo "Next steps:"
echo "  go mod tidy"
echo "  make build"
echo "  make test"
echo ""
echo "Happy coding! 🚀"
