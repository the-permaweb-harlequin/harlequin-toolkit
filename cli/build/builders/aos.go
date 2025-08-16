package builders

import (
	"context"

	harlequinBuild "github.com/the-permaweb-harlequin/harlequin-toolkit/cli/build"
	harlequinConfig "github.com/the-permaweb-harlequin/harlequin-toolkit/cli/config"
)

/*
AOS builder is a builder for the vanilla AOS module
It uses the AO build container to build the project
Requires a config file for the build container (stack memory, target, etc)
Requires the AOS git hash to clone the repo
Requires the project path to bundle the lua code


Steps:
1. Clone the AOS repo and clean the imports
2. Create the build directory
3. Bundle the lua code and write the bundle to the build directory
4. Inject the bundle into the AOS code
5. Call the container to build the project
6. Write the process.wasm and the bundled lua code to the output directory
7. Clean up the build container and build directory

*/

type AOSBuilder struct {
	entrypoint string
	outputDir  string
	config *harlequinConfig.Config
	runner *harlequinBuild.BuildRunner
}

func NewAOSBuilder(config *harlequinConfig.Config) *AOSBuilder {
	runner, err := harlequinBuild.NewAOBuildRunner(config, "")
	if err != nil {
		panic(err)
	}
	return &AOSBuilder{config: config, runner: runner}
}

func (b *AOSBuilder) Build(ctx context.Context, projectPath string) error {
	return b.runner.BuildProject(ctx, projectPath)
}