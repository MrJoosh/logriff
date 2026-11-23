# LogRiff (C)

C implementation of the LogRiff CLI.

## Setup

No dependency installation is needed; a C11 compiler is used directly, with CMake as an optional build path.

## Usage

```bash
just c build
just c check
just c run --format json ../data/sample.log
```

Build output in `build/` is git-ignored.

See the repository root `README.md` for the full CLI specification.
