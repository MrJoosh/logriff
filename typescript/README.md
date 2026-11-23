# TypeScript implementation (scaffold)

This directory is prepared for a future TypeScript implementation of `logriff`; no CLI logic exists yet. The shared CLI and output specification lives in [../README.md](../README.md).

## Layout

- `package.json` / `package-lock.json` / `tsconfig.json` — Node, reproducible dependency, and strict TypeScript configuration.
- `src/logriff.ts` — placeholder entry point that exits non-zero until implemented.
- `Justfile` — build/run/check recipes exposed through the root `Justfile`.

## Getting started

```bash
just typescript sync
just typescript check
just typescript build
just typescript run --format json ../data/sample.log
```

Replace `src/logriff.ts` with the real CLI when implementation begins. Compiled output in `dist/` and installed packages in `node_modules/` are git-ignored.
