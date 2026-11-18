package nixpin

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/quiqr/chatcms-app/internal/store"
)

type Postgres struct {
	queries *store.Queries
}

func NewPostgres(queries *store.Queries) *Postgres {
	return &Postgres{queries: queries}
}

func (p *Postgres) Get(ctx context.Context, name, version, system string) (Package, bool, error) {
	row, err := p.queries.GetNixPackage(ctx, store.GetNixPackageParams{Name: name, Version: version, System: system})
	if errors.Is(err, pgx.ErrNoRows) {
		return Package{}, false, nil
	}
	if err != nil {
		return Package{}, false, err
	}
	return Package{
		Name: row.Name, Version: row.Version, System: row.System,
		Revision: row.Revision, AttrPath: row.AttrPath, StorePath: row.StorePath,
	}, true, nil
}

func (p *Postgres) Put(ctx context.Context, pkg Package) error {
	return p.queries.PutNixPackage(ctx, store.PutNixPackageParams{
		Name: pkg.Name, Version: pkg.Version, System: pkg.System,
		Revision: pkg.Revision, AttrPath: pkg.AttrPath, StorePath: pkg.StorePath,
	})
}

var _ Store = (*Postgres)(nil)
