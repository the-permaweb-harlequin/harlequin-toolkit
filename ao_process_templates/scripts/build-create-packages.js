#!/usr/bin/env node

import fs from 'fs-extra';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const LANGUAGES = {
  assemblyscript: {
    name: 'AssemblyScript',
    emoji: '🎭',
    dependencies: {
      '@permaweb/ao-loader': '^0.0.49',
      '@types/node': '^20.0.0',
      'assemblyscript': '^0.27.0',
      'tsx': '^4.20.5',
      'typescript': '^5.3.0'
    },
    peerDependencies: {
      'assemblyscript-json': '^1.1.0'
    }
  },
  go: {
    name: 'Go',
    emoji: '🐹',
    dependencies: {
      '@permaweb/ao-loader': '^0.0.49'
    }
  }
};

async function buildCreatePackage(language) {
  console.log(`📦 Building create-ao-${language} package...`);

  const languageConfig = LANGUAGES[language];
  if (!languageConfig) {
    throw new Error(`Unknown language: ${language}`);
  }

  const templateDir = path.join(__dirname, '..', 'languages', language, 'template');
  const outputDir = path.join(__dirname, '..', 'create-packages', `create-ao-${language}`);

  // Clean output directory
  await fs.remove(outputDir);
  await fs.ensureDir(outputDir);

  // Create bin directory and script
  await fs.ensureDir(path.join(outputDir, 'bin'));

  const binScript = `#!/usr/bin/env node

import fs from 'fs-extra';
import path from 'path';
import { fileURLToPath } from 'url';
import prompts from 'prompts';
import chalk from 'chalk';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

async function main() {
  console.log(chalk.blue('${languageConfig.emoji} Creating a new AO ${languageConfig.name} process...\\n'));

  // Get project name from command line or prompt
  const projectName = process.argv[2] || (await prompts({
    type: 'text',
    name: 'name',
    message: 'What is your project name?',
    initial: 'my-ao-process'
  })).name;

  if (!projectName) {
    console.log(chalk.red('❌ Project name is required'));
    process.exit(1);
  }

  const targetDir = path.resolve(process.cwd(), projectName);
  const templateDir = path.join(__dirname, '..', 'template');

  // Check if directory already exists
  if (fs.existsSync(targetDir)) {
    const { overwrite } = await prompts({
      type: 'confirm',
      name: 'overwrite',
      message: \`Directory \${projectName} already exists. Overwrite?\`,
      initial: false
    });

    if (!overwrite) {
      console.log(chalk.yellow('⚠️ Operation cancelled'));
      process.exit(0);
    }

    fs.removeSync(targetDir);
  }

  // Copy template
  console.log(chalk.green(\`📁 Creating project in \${targetDir}...\`));
  fs.copySync(templateDir, targetDir);

  // Replace template variables
  await replaceTemplateVariables(targetDir, projectName);

  console.log(chalk.green('✅ Project created successfully!\\n'));
  console.log(chalk.cyan('Next steps:'));
  console.log(chalk.white(\`  cd \${projectName}\`));
  ${language === 'assemblyscript' ?
    `console.log(chalk.white('  pnpm install'));
  console.log(chalk.white('  pnpm run build'));
  console.log(chalk.white('  pnpm run test'));` :
    `console.log(chalk.white('  go mod tidy'));
  console.log(chalk.white('  make build'));
  console.log(chalk.white('  make test'));`
  }
  console.log(chalk.gray('\\nHappy coding! 🚀'));
}

async function replaceTemplateVariables(targetDir, projectName) {
  const files = await fs.readdir(targetDir, { recursive: true });

  for (const file of files) {
    const filePath = path.join(targetDir, file);
    const stat = await fs.stat(filePath);

    if (stat.isFile()) {
      const content = await fs.readFile(filePath, 'utf8');
      const newContent = content.replace(/{{PROJECT_NAME}}/g, projectName);

      if (content !== newContent) {
        await fs.writeFile(filePath, newContent);
      }
    }
  }
}

main().catch(console.error);`;

  await fs.writeFile(path.join(outputDir, 'bin', `create-ao-${language}.js`), binScript);
  await fs.chmod(path.join(outputDir, 'bin', `create-ao-${language}.js`), 0o755);

  // Copy template directory
  await fs.copy(templateDir, path.join(outputDir, 'template'));

  // Create package.json
  const packageJson = {
    name: `create-ao-${language}`,
    version: '1.0.0',
    description: `Create a new AO process using ${languageConfig.name}`,
    type: 'module',
    bin: {
      [`create-ao-${language}`]: `./bin/create-ao-${language}.js`
    },
    files: [
      'bin/',
      'template/',
      'README.md'
    ],
    scripts: {
      build: 'echo "No build needed"',
      test: 'echo "No tests yet"',
      prepublishOnly: 'echo "Ready to publish"'
    },
    keywords: [
      'ao',
      'arweave',
      language,
      'webassembly',
      'template',
      'create',
      'scaffolding'
    ],
    author: 'The Permaweb Harlequin',
    license: 'MIT',
    dependencies: {
      'fs-extra': '^11.1.1',
      'prompts': '^2.4.2',
      'chalk': '^5.3.0'
    }
  };

  await fs.writeJSON(path.join(outputDir, 'package.json'), packageJson, { spaces: 2 });

  // Create README
  const readme = `# create-ao-${language}

Create a new AO process using ${languageConfig.name}.

## Usage

\`\`\`bash
npx create-ao-${language} my-project
\`\`\`

This will create a new directory with a complete AO ${languageConfig.name} project ready for development.

## What's Included

- ${languageConfig.name} AO process template
- Build configuration
- Test suite using @permaweb/ao-loader
- Documentation and examples
- Development workflow

## Requirements

${language === 'assemblyscript' ? '- Node.js 18+' : '- Node.js 18+\n- Go 1.21+'}

## Learn More

- [AO Documentation](https://ao.arweave.dev/)
- [Harlequin Toolkit](https://github.com/the-permaweb-harlequin/harlequin-toolkit)
`;

  await fs.writeFile(path.join(outputDir, 'README.md'), readme);

  console.log(`✅ Built create-ao-${language} package`);
}

async function main() {
  const language = process.argv[2];

  if (!language) {
    console.log('Usage: node build-create-packages.js <language>');
    console.log('Available languages:', Object.keys(LANGUAGES).join(', '));
    process.exit(1);
  }

  if (language === 'all') {
    for (const lang of Object.keys(LANGUAGES)) {
      await buildCreatePackage(lang);
    }
  } else {
    await buildCreatePackage(language);
  }
}

// Run if this script is executed directly
if (import.meta.url.startsWith('file:') && process.argv[1] && import.meta.url.includes(process.argv[1])) {
  main().catch(console.error);
}
