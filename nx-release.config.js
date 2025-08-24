module.exports = {
  projects: {
    cli: {
      projectRoot: 'cli',
      tagPrefix: 'cli-v',
      releaseTagPattern: 'cli-v{version}',
      changelog: {
        workspaceChangelog: false,
        projectChangelogs: {
          cli: {
            createRelease: 'github',
            entryWhenNoChanges: false,
            renderOptions: {
              authors: false,
              commitReferences: true,
              versionTitleDate: true
            }
          }
        }
      },
      version: {
        conventionalCommits: true,
        preVersionCommand: 'npx nx run cli:build',
        postVersionCommand: 'npx nx run cli:publish'
      }
    }
  }
};
