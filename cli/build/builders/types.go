package builders

import "context"

type Builder interface {
	Build(ctx context.Context, projectPath string) error
	Clean(ctx context.Context, projectPath string) error
	Logs(ctx context.Context, projectPath string) error
	Status(ctx context.Context, projectPath string) error
	Stop(ctx context.Context, projectPath string) error
	Start(ctx context.Context, projectPath string) error
	Restart(ctx context.Context, projectPath string) error
}
