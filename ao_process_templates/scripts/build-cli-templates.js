#!/usr/bin/env node

import fs from 'fs-extra';
import path from 'path';
import { fileURLToPath } from 'url';
import { execSync } from 'child_process';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const LANGUAGES = {
  assemblyscript: {
    name: 'AssemblyScript',
    description: 'AssemblyScript AO process with WASM compilation',
    instructions: [
      'pnpm install',
      'pnpm run build',
      'pnpm run test'
    ]
  },
  go: {
    name: 'Go',
    description: 'Go AO process with WASM compilation',
    instructions: [
      'go mod tidy',
      'make build',
      'make test'
    ]
  }
};

async function buildCliTemplate(language) {
  console.log(`📦 Building CLI template for ${language}...`);

  const languageConfig = LANGUAGES[language];
  if (!languageConfig) {
    throw new Error(`Unknown language: ${language}`);
  }

  const templateDir = path.join(__dirname, '..', 'languages', language, 'template');
  const outputDir = path.join(__dirname, '..', 'cli-templates');

  await fs.ensureDir(outputDir);

  // Create a temporary directory for CLI template
  const tempDir = path.join(outputDir, `temp-${language}`);
  await fs.remove(tempDir);
  await fs.copy(templateDir, tempDir);

  // Create CLI-specific metadata
  const cliMetadata = {
    language,
    name: languageConfig.name,
    description: languageConfig.description,
    instructions: languageConfig.instructions,
    version: '1.0.0',
    created: new Date().toISOString()
  };

  await fs.writeJSON(path.join(tempDir, '.harlequin-template.json'), cliMetadata, { spaces: 2 });

  // Create CLI installation script
  const installScript = `#!/bin/bash
# Harlequin CLI Template Installation Script for ${languageConfig.name}

set -e

PROJECT_NAME="$1"
if [ -z "$PROJECT_NAME" ]; then
    echo "Usage: $0 <project-name>"
    exit 1
fi

echo "🎭 Creating ${languageConfig.name} AO process: $PROJECT_NAME"

# Replace template variables
find . -type f -name "*.md" -o -name "*.json" -o -name "*.go" -o -name "go.mod" | xargs sed -i.bak "s/{{PROJECT_NAME}}/$PROJECT_NAME/g"
find . -name "*.bak" -delete

echo "✅ Template prepared successfully!"
echo ""
echo "Next steps:"
${languageConfig.instructions.map(cmd => `echo "  ${cmd}"`).join('\n')}
echo ""
echo "Happy coding! 🚀"
`;

  await fs.writeFile(path.join(tempDir, 'install.sh'), installScript);
  await fs.chmod(path.join(tempDir, 'install.sh'), 0o755);

  // Create tarball
  const tarballPath = path.join(outputDir, `${language}.tar.gz`);

  try {
    execSync(`tar -czf "${tarballPath}" -C "${tempDir}" .`, { stdio: 'inherit' });
    console.log(`✅ Created CLI template: ${tarballPath}`);
  } catch (error) {
    console.error(`Failed to create tarball: ${error.message}`);
    throw error;
  } finally {
    // Clean up temp directory
    await fs.remove(tempDir);
  }

  // Create template manifest for CLI
  const manifestPath = path.join(outputDir, `${language}.json`);
  const manifest = {
    ...cliMetadata,
    tarball: `${language}.tar.gz`,
    size: (await fs.stat(tarballPath)).size
  };

  await fs.writeJSON(manifestPath, manifest, { spaces: 2 });

  console.log(`✅ Built CLI template for ${language}`);
}

async function buildAllTemplates() {
  const outputDir = path.join(__dirname, '..', 'cli-templates');

  // Create master manifest
  const templates = {};

  for (const language of Object.keys(LANGUAGES)) {
    await buildCliTemplate(language);

    const manifestPath = path.join(outputDir, `${language}.json`);
    const manifest = await fs.readJSON(manifestPath);
    templates[language] = manifest;
  }

  // Write master manifest
  const masterManifest = {
    version: '1.0.0',
    generated: new Date().toISOString(),
    templates
  };

  await fs.writeJSON(path.join(outputDir, 'templates.json'), masterManifest, { spaces: 2 });

  console.log('✅ Built all CLI templates and master manifest');
}

async function main() {
  const language = process.argv[2];

  if (!language) {
    console.log('Usage: node build-cli-templates.js <language|all>');
    console.log('Available languages:', Object.keys(LANGUAGES).join(', '));
    process.exit(1);
  }

  if (language === 'all') {
    await buildAllTemplates();
  } else {
    await buildCliTemplate(language);
  }
}

// Run if this script is executed directly
if (import.meta.url.startsWith('file:') && process.argv[1] && import.meta.url.includes(process.argv[1])) {
  main().catch(console.error);
}
