# Spec Delta

## Purpose

Defines how a site that declares its runtime is brought in.

## ADDED Requirements

### Requirement: A runtime definition makes a site recognisable

A site with a readable `quiqr/runtime.yaml` SHALL be brought in as a site of
that runtime, without a Hugo configuration, and its Quiqr model SHALL be
read as for any site. A runtime definition that cannot be read SHALL be the
problem reported.

#### Scenario: a Quarto book

- **WHEN** an implementer brings in a Quarto book with a runtime definition
  and a model
- **THEN** it SHALL be reported as a site with its model's pages and
  collections

#### Scenario: a broken runtime definition

- **WHEN** the runtime definition cannot be read
- **THEN** the report SHALL say what is wrong with it
