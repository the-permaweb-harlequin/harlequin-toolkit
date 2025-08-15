package builders

import (
	"context"

	harlequinBuild "github.com/the-permaweb-harlequin/harlequin-toolkit/cli/build"
	harlequinConfig "github.com/the-permaweb-harlequin/harlequin-toolkit/cli/config"
)

type AOSBuilder struct {
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