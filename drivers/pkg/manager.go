package pkg

import (
	"context"
	"github.com/settlectl/settle-core/common"
	"github.com/settlectl/settle-core/inventory"
)


type PackageManager interface {
	Install(ctx context.Context, runtimeCtx *inventory.Context, packages []common.Package) error
	Remove(ctx context.Context, runtimeCtx *inventory.Context, packages []common.Package) error
	DoesExist(ctx context.Context, runtimeCtx *inventory.Context, packages []common.Package) (bool, error)
}