# LogRiff developer tasks.
#
# Run `just` or `just help` to see the LogRiff banner and all recipes.

set shell := ["bash", "-eu", "-o", "pipefail", "-c"]

# ---------------------------------------------------------------------------
# Housekeeping
# ---------------------------------------------------------------------------

# Print the LogRiff banner.
[group('housekeeping')]
title:
    @cat .github/title.txt

# Show the banner and list available recipes.
[group('housekeeping')]
[default]
help: title
    @echo
    @just --list

# Install/sync dependencies for every implementation.
[group('housekeeping')]
setup:
    cd python && uv sync
    cd go && go mod download
    cd rust && cargo fetch

# Format all source code.
[group('housekeeping')]
fmt: python-fmt go-fmt rust-fmt

# Check formatting without changing files.
[group('housekeeping')]
fmt-check: python-fmt-check go-fmt-check rust-fmt-check

# Run all linters and static checks.
[group('housekeeping')]
lint: python-lint go-vet rust-clippy lua-check

# Run all available test suites.
[group('housekeeping')]
test: python-test data-test go-test rust-test

# Run linting and tests.
[group('housekeeping')]
check: lint test

# CI-style check plus the integration script.
[group('housekeeping')]
ci: check integration

# Remove build artifacts, caches, and generated data.
[group('housekeeping')]
clean: python-clean go-clean rust-clean cpp-clean lua-clean data-clean

# Remove everything `clean` does, plus the Python virtualenv.
[group('housekeeping')]
distclean: clean
    rm -rf python/.venv

# ---------------------------------------------------------------------------
# Python
# ---------------------------------------------------------------------------

# Sync the Python environment.
[group('python')]
python-sync:
    cd python && uv sync

# Run the Python CLI with optional arguments.
[group('python')]
python-run *args:
    cd python && uv run logriff {{args}}

# Run the Python CLI against the sample log.
[group('python')]
python-sample:
    cd python && uv run logriff ../data/sample.log

# Show the Python CLI help.
[group('python')]
python-help:
    cd python && uv run logriff --help

# Run the Python test suite.
[group('python')]
python-test:
    cd python && uv run pytest tests/ -q

# Run the Python tests with coverage.
[group('python')]
python-cov:
    cd python && uv run pytest --cov=logriff --cov-report=term-missing tests/

# Lint Python source and the data generator.
[group('python')]
python-lint:
    cd python && uvx ruff check . ../data

# Format Python source and the data generator.
[group('python')]
python-fmt:
    cd python && uvx ruff format . ../data

# Check Python formatting without changing files.
[group('python')]
python-fmt-check:
    cd python && uvx ruff format --check . ../data

# Build Python packages.
[group('python')]
python-build:
    cd python && uv build

# Remove Python build artifacts and caches.
[group('python')]
python-clean:
    rm -rf python/dist python/build python/.coverage python/.pytest_cache python/.ruff_cache
    find python -type d -name __pycache__ -prune -exec rm -rf {} +

# ---------------------------------------------------------------------------
# Go
# ---------------------------------------------------------------------------

# Build the Go implementation.
[group('go')]
go-build:
    cd go && go build -o logriff .

# Run the Go implementation with optional arguments.
[group('go')]
go-run *args:
    cd go && go run . {{args}}

# Run the Go tests.
[group('go')]
go-test:
    cd go && go test ./...

# Run Go static analysis.
[group('go')]
go-vet:
    cd go && go vet ./...

# Format Go source.
[group('go')]
go-fmt:
    cd go && gofmt -w .

# Check Go formatting without changing files.
[group('go')]
go-fmt-check:
    cd go && test -z "$(gofmt -l .)"

# Remove the Go binary.
[group('go')]
go-clean:
    rm -rf go/bin go/logriff go/logriff.exe

# ---------------------------------------------------------------------------
# Rust
# ---------------------------------------------------------------------------

# Build the Rust implementation.
[group('rust')]
rust-build:
    cd rust && cargo build

# Build the Rust implementation in release mode.
[group('rust')]
rust-release:
    cd rust && cargo build --release

# Run the Rust implementation with optional arguments.
[group('rust')]
rust-run *args:
    cd rust && cargo run -- {{args}}

# Run the Rust tests.
[group('rust')]
rust-test:
    cd rust && cargo test

# Type-check the Rust implementation.
[group('rust')]
rust-check:
    cd rust && cargo check

# Run Clippy with warnings treated as errors.
[group('rust')]
rust-clippy:
    cd rust && cargo clippy --all-targets -- -D warnings

# Format Rust source.
[group('rust')]
rust-fmt:
    cd rust && cargo fmt

# Check Rust formatting without changing files.
[group('rust')]
rust-fmt-check:
    cd rust && cargo fmt --check

# Remove Rust build artifacts.
[group('rust')]
rust-clean:
    cd rust && cargo clean

# ---------------------------------------------------------------------------
# C++
# ---------------------------------------------------------------------------

# Build the C++ implementation with CMake, falling back to the compiler.
[group('cpp')]
cpp-build:
    mkdir -p cpp/build
    if command -v cmake >/dev/null 2>&1; then \
        cmake -S cpp -B cpp/build -DCMAKE_BUILD_TYPE=Release; \
        cmake --build cpp/build; \
    else \
        echo "cmake not found; compiling directly with ${CXX:-clang++}"; \
        "${CXX:-clang++}" -std=c++20 -Wall -Wextra -Wpedantic \
            cpp/src/main.cpp -o cpp/build/logriff; \
    fi

# Run the C++ implementation with optional arguments.
[group('cpp')]
cpp-run *args: cpp-build
    ./cpp/build/logriff {{args}}

# Remove C++ build artifacts.
[group('cpp')]
cpp-clean:
    rm -rf cpp/build

# ---------------------------------------------------------------------------
# Lua
# ---------------------------------------------------------------------------

# Run the Lua implementation with optional arguments.
[group('lua')]
lua-run *args:
    lua lua/logriff.lua {{args}}

# Check the Lua source, or skip if Lua is not installed.
[group('lua')]
lua-check:
    if command -v luac >/dev/null 2>&1; then \
        luac -p lua/logriff.lua; \
    elif command -v lua >/dev/null 2>&1; then \
        lua -e 'assert(loadfile("lua/logriff.lua"))'; \
    else \
        echo "Skipping Lua check: lua/luac not installed"; \
    fi

# Remove Lua bytecode artifacts.
[group('lua')]
lua-clean:
    rm -f lua/*.luac luac.out

# ---------------------------------------------------------------------------
# Data generation
# ---------------------------------------------------------------------------

# Generate synthetic logs, e.g. `just data-generate --lines 100 --out data/test.log`.
[group('data')]
data-generate *args:
    python3 data/generate_logs.py {{args}}

# Generate a deterministic 1M-line log for benchmarking.
[group('data')]
data-1m:
    python3 data/generate_logs.py --lines 1000000 --seed 42 --out data/test_1m.log

# Run the data generator tests.
[group('data')]
data-test:
    cd python && uv run pytest ../data/tests -q

# Remove generated log data.
[group('data')]
data-clean:
    rm -f data/test_*.log data/generated*.log

# ---------------------------------------------------------------------------
# Integration
# ---------------------------------------------------------------------------

# Run the cross-language integration script.
[group('integration')]
integration:
    ./tests/integration_test.sh
