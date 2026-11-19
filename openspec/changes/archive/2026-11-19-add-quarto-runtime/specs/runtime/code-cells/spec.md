# Spec Delta

## Purpose

Defines how a site's Python and R code cells get their packages and where
they run.

## ADDED Requirements

### Requirement: Python and R packages are declared with their revision

A runtime definition MAY name Python packages and R packages, each list with
the nixpkgs revision to build them from. ChatCMS SHALL build one Python
environment, with Jupyter added, and one R environment, and SHALL hand them
to the preview and the build.

#### Scenario: Python packages

- **WHEN** a site declares `python: { nixpkgs: <rev>, packages: [pandas] }`
- **THEN** its cells SHALL run with a Python that has pandas and Jupyter

#### Scenario: a package name that is not one

- **WHEN** a listed package name holds anything but letters, digits, `-`
  and `_`
- **THEN** the definition SHALL be refused, naming it

### Requirement: Code cells run without network or secrets

Code cells SHALL run only inside the workspace and the build sandbox. In the
build sandbox they SHALL have no network and no secret of the installation.

#### Scenario: a cell in a build

- **WHEN** a publish builds a Quarto site with a Python cell
- **THEN** the cell SHALL run, and the sandbox SHALL give it no network
