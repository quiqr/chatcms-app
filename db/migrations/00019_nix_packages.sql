-- +goose Up

CREATE TABLE nix_packages (
    name        TEXT        NOT NULL,
    version     TEXT        NOT NULL,
    system      TEXT        NOT NULL,
    revision    TEXT        NOT NULL,
    attr_path   TEXT        NOT NULL,
    store_path  TEXT        NOT NULL,
    realised_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (name, version, system)
);

-- +goose Down

DROP TABLE nix_packages;
