//go:build integration

package nixpin_test

import (
	"testing"

	"github.com/quiqr/chatcms-app/internal/nixpin"
	"github.com/quiqr/chatcms-app/internal/store"
	"github.com/quiqr/chatcms-app/internal/testsupport"
)

func TestPostgresRemembersAPackage(t *testing.T) {
	db, err := store.Open(t.Context(), testsupport.Postgres(t))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Migrate(t.Context()); err != nil {
		t.Fatal(err)
	}
	pins := nixpin.NewPostgres(db.Queries)

	if _, found, err := pins.Get(t.Context(), "hugo", "0.139.0", "x86_64-linux"); err != nil || found {
		t.Fatalf("Get() before = %v, %v", found, err)
	}
	pkg := nixpin.Package{Name: "hugo", Version: "0.139.0", System: "x86_64-linux", Revision: "r1", AttrPath: "hugo", StorePath: "/nix/store/a"}
	if err := pins.Put(t.Context(), pkg); err != nil {
		t.Fatal(err)
	}
	pkg.StorePath = "/nix/store/b"
	if err := pins.Put(t.Context(), pkg); err != nil {
		t.Fatal(err)
	}
	got, found, err := pins.Get(t.Context(), "hugo", "0.139.0", "x86_64-linux")
	if err != nil || !found || got != pkg {
		t.Errorf("Get() = %+v, %v, %v, want %+v", got, found, err, pkg)
	}
}
