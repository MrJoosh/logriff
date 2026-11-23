# C implementation (scaffold)

This directory is prepared for a future C implementation of `logriff`; no CLI logic exists yet. The shared CLI and output specification lives in [../README.md](../README.md).

## Layout

- `CMakeLists.txt` / `Justfile` — C11 build configuration, with a direct `cc` fallback when CMake is unavailable.
- `src/main.c` — placeholder entry point that exits non-zero until implemented.
- `Justfile` — build/run/check recipes exposed through the root `Justfile`.

## Getting started

```bash
just c build
just c check
just c run --format json ../data/sample.log
```

Replace `src/main.c` with the real CLI when the C port is started. Build output in `build/` is git-ignored.
