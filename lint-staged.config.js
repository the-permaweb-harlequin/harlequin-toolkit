module.exports = {
  // App TypeScript/JavaScript files
  'app/src/**/*.{ts,tsx,js,jsx}': [
    (filenames) =>
      `cd app && npx eslint ${filenames
        .map((f) => f.replace('app/src/', 'src/'))
        .join(' ')} --cache --fix`,
    'prettier --write'
  ],
  
  // Docs TypeScript/JavaScript files
  'docs/**/*.{ts,tsx,js,jsx,mjs}': [
    (filenames) =>
      `npx eslint ${filenames.join(' ')} --cache --fix`,
    'prettier --write'
  ],
  
  // App-specific styles - disable for now due to prettier compatibility issues
  // 'app/src/**/*.{less,css}': ['stylelint --fix'],
  
  // General formatting for config and documentation files
  '*.{json,md,yml,yaml}': ['prettier --write'],
  'app/*.{json,md,yml,yaml}': ['prettier --write'],
  'cli/*.{json,md,yml,yaml}': ['prettier --write'],
  'docs/**/*.{json,md,yml,yaml}': ['prettier --write'],
}
