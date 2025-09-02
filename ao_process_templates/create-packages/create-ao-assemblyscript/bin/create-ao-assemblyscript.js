#!/usr/bin/env node

import fs from 'fs-extra';
import path from 'path';
import { fileURLToPath } from 'url';
import prompts from 'prompts';
import chalk from 'chalk';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

async function main() {
  console.log(chalk.blue('🎭 Creating a new AO AssemblyScript process...\n'));

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
      message: `Directory ${projectName} already exists. Overwrite?`,
      initial: false
    });

    if (!overwrite) {
      console.log(chalk.yellow('⚠️ Operation cancelled'));
      process.exit(0);
    }

    fs.removeSync(targetDir);
  }

  // Copy template
  console.log(chalk.green(`📁 Creating project in ${targetDir}...`));
  fs.copySync(templateDir, targetDir);

  // Replace template variables
  await replaceTemplateVariables(targetDir, projectName);

  console.log(chalk.green('✅ Project created successfully!\n'));
  console.log(chalk.cyan('Next steps:'));
  console.log(chalk.white(`  cd ${projectName}`));
  console.log(chalk.white('  pnpm install'));
  console.log(chalk.white('  pnpm run build'));
  console.log(chalk.white('  pnpm run test'));
  console.log(chalk.gray('\nHappy coding! 🚀'));
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

main().catch(console.error);