# LogRiff developer tasks.
#
# Run `just` or `just help` to see the LogRiff banner and all recipes.

set shell := ["bash", "-eu", "-o", "pipefail", "-c"]

# C implementation tasks.
mod c 'c/Justfile'

# C++ implementation tasks.
mod cpp 'cpp/Justfile'

# Data generation tasks.
mod data 'data/Justfile'

# Go implementation tasks.
mod go 'go/Justfile'

# Lua implementation tasks.
mod lua 'lua/Justfile'

# Python implementation tasks.
mod python 'python/Justfile'

# Rust implementation tasks.
mod rust 'rust/Justfile'

# TypeScript implementation tasks.
mod typescript 'typescript/Justfile'

# ---------------------------------------------------------------------------
# Housekeeping
# ---------------------------------------------------------------------------

# Print the LogRiff banner.
[group('housekeeping')]
title:
    @cat .github/title.txt

# Show the banner and list available recipes, optionally for one module.
[group('housekeeping')]
[default]
help *args: title
    @echo
    @just --list {{args}} --list-submodules

# Install/sync dependencies for every implementation.
[group('housekeeping')]
setup: python::sync go::sync rust::sync typescript::sync

# Format all source code.
[group('housekeeping')]
fmt: python::fmt go::fmt rust::fmt

# Check formatting without changing files.
[group('housekeeping')]
fmt-check: python::fmt-check go::fmt-check rust::fmt-check

# Run all linters and static checks.
[group('housekeeping')]
lint: python::lint go::vet rust::lint lua::check c::check typescript::check

# Run all available test suites.
[group('housekeeping')]
test: python::test data::test go::test rust::test c::test typescript::test

# Run linting and tests.
[group('housekeeping')]
check: lint test

# CI-style check plus the integration script.
[group('housekeeping')]
ci: check integration

# Remove build artifacts, caches, and generated data.
[group('housekeeping')]
clean: python::clean go::clean rust::clean cpp::clean lua::clean data::clean c::clean typescript::clean

# Remove everything `clean` does, plus the Python virtualenv.
[group('housekeeping')]
distclean: clean
    rm -rf python/.venv

# ---------------------------------------------------------------------------
# Integration
# ---------------------------------------------------------------------------

# Run the cross-language integration script.
[group('integration')]
integration:
    ./tests/integration_test.sh
