# LogRiff: Multi-Language CLI Log Filter & Aggregator

A standardized, multi-language coding exercise designed to compare idioms, memory management, I/O throughput, and concurrency across **Python**, **Go**, **Rust**, **C++**, **Lua**, **TypeScript**, and **C**.

TypeScript and C are scaffolded but not yet implemented: their directories contain build tooling, placeholder entry points, and notes, ready for the real CLI ports.

---

## 1. Project Overview

`logriff` is a high-performance command-line utility that ingests structured access logs from a file or standard input (`stdin`), filters records according to command-line flags, computes aggregate traffic metrics, and produces formatted text or JSON summaries.

### Primary Learning Objectives

| Language | Primary Focus Areas |
| :--- | :--- |
| **Python** | Baseline implementation, generator pipelines, `argparse`, `sys.stdin` streaming. |
| **Go** | `bufio.Scanner`, goroutines/worker pools, interfaces, table-driven testing. |
| **Rust** | Zero-copy string slicing (`&str` vs `String`), explicit error handling (`Result`), ownership/borrowing, `clap`. |
| **C++ (C++20)** | Streaming I/O optimizations, `std::string_view`, RAII, standard algorithms, value semantics. |
| **Lua (5.4 / LuaJIT)** | Native pattern matching (no POSIX regex), table-based maps, streaming via `io.lines`, CLI argument parsing without external libraries. |
| **TypeScript (planned)** | Static typing on Node.js, stream-based I/O, `package.json`/`tsconfig.json` tooling, tests with the Node test runner. |
| **C (planned)** | Manual memory management, explicit allocation/free, `stdio`-based streaming, C11 toolchain and warning flags, no language runtime. |

---

## 2. Input Data Format

`logriff` operates on a standardized, space-delimited Common/Combined Log Format variant:

```text
<remote_addr> - - [<timestamp>] "<method> <path> <protocol>" <status> <bytes_sent> <response_time_ms> "<user_agent>"
```

### Field Definitions

| Field | Type | Description | Example |
| :--- | :--- | :--- | :--- |
| `remote_addr` | String | Client IP address (IPv4 or IPv6) | `192.168.1.42` |
| `timestamp` | String | Bracketed ISO-like or standard Apache timestamp | `[10/Oct/2026:13:55:36 +0000]` |
| `method` | String | HTTP verb (`GET`, `POST`, `PUT`, `DELETE`, `HEAD`, `PATCH`, `OPTIONS`) | `GET` |
| `path` | String | URI resource path (excluding query parameters for aggregation) | `/api/v1/users` |
| `protocol` | String | Protocol version | `HTTP/1.1` |
| `status` | Integer | HTTP response status code ($100 \le \text{status} \le 599$) | `200` |
| `bytes_sent` | Integer | Size of response body in bytes ($0$ or positive integer) | `4521` |
| `response_time_ms` | Float | Server execution latency in milliseconds | `42.5` |
| `user_agent` | String | Quoted client user agent | `"Mozilla/5.0..."` |

### Sample Log Lines

```text
192.168.1.10 - - [06/Nov/2026:08:12:01 +0000] "GET /api/v1/health HTTP/1.1" 200 42 1.2 "Go-http-client/1.1"
10.0.0.5 - - [06/Nov/2026:08:12:02 +0000] "POST /api/v1/auth/login HTTP/1.1" 401 128 85.4 "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"
192.168.1.15 - - [06/Nov/2026:08:12:03 +0000] "GET /api/v1/items/4912 HTTP/2.0" 200 8920 14.8 "curl/8.1.2"
172.16.0.22 - - [06/Nov/2026:08:12:04 +0000] "GET /assets/main.css HTTP/1.1" 304 0 0.8 "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"
10.0.0.12 - - [06/Nov/2026:08:12:05 +0000] "DELETE /api/v1/sessions/current HTTP/1.1" 500 240 210.6 "PostmanRuntime/7.32.3"
```

---

## 3. CLI Specification & Interface

### Usage

```bash
logriff [OPTIONS] [FILE]
```

If `[FILE]` is omitted or passed as `-`, read input from `stdin`.

### Options & Flags

| Flag | Short | Value Type | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--status-min` | `-s` | Integer | None | Keep records with `status >= MIN`. |
| `--status-max` | `-e` | Integer | None | Keep records with `status <= MAX`. |
| `--method` | `-m` | String | None | Filter by HTTP verb (case-insensitive, e.g., `GET`, `POST`). |
| `--path-prefix`| `-p` | String | None | Keep only requests where `path` starts with given prefix. |
| `--min-duration`| | Float | None | Keep records with `response_time_ms >= duration`. |
| `--top-routes` | `-n` | Integer | `5` | Number of most requested routes to show in summary. |
| `--format` | `-f` | Enum (`text`, `json`) | `text` | Output format for aggregate results. |
| `--strict` | | Boolean Flag | `false` | Terminate immediately with non-zero exit code on malformed lines. |
| `--help` | `-h` | Boolean Flag | | Print help and options. |

---

## 4. Aggregations & Metrics Requirements

After processing input lines and applying all active filters, `logriff` must calculate:

1. **Record Counts:**
   * Total lines ingested ($N_{\text{total}}$).
   * Total lines matched after filtering ($N_{\text{matched}}$).
   * Malformed / unparseable lines skipped ($N_{\text{errors}}$).
2. **Byte Volume:**
   * Total bytes sent across matched records ($\sum \text{bytes}$).
3. **Latency Metrics (on matched records):**
   * Minimum latency ($\min$).
   * Mean latency ($\mu$).
   * 95th Percentile latency ($p95$).
   * Maximum latency ($\max$).
4. **Distribution Breakdowns:**
   * **Status Codes:** Count grouped by code (e.g., `200: 4801`, `404: 12`, `500: 3`).
   * **Top Routes:** Top $N$ most-frequently hit paths (without query strings) with hit counts.

---

## 5. Output Formats

### 5.1 Text Output (`--format text`)

```text
=== LogRiff Summary ===
Lines Processed : 1,000,000
Matched Records : 842,100
Skipped (Errors): 14

--- Traffic & Bandwidth ---
Total Volume    : 3.42 GB (3,672,810,496 bytes)

--- Latency (ms) ---
Min             : 0.12 ms
Mean            : 18.45 ms
p95             : 82.10 ms
Max             : 1412.30 ms

--- Status Codes ---
200 : 798,000  (94.76%)
401 :  31,100  ( 3.69%)
404 :  11,000  ( 1.31%)
500 :   2,000  ( 0.24%)

--- Top Routes (Top 5) ---
1. /api/v1/health             : 320,000
2. /api/v1/items              : 210,000
3. /api/v1/auth/login         : 180,000
4. /assets/main.css           :  82,100
5. /api/v1/users              :  50,000
========================
```

### 5.2 JSON Output (`--format json`)

```json
{
  "summary": {
    "total_lines": 1000000,
    "matched_lines": 842100,
    "error_lines": 14
  },
  "traffic": {
    "total_bytes": 3672810496
  },
  "latency_ms": {
    "min": 0.12,
    "mean": 18.45,
    "p95": 82.10,
    "max": 1412.30
  },
  "status_codes": {
    "200": 798000,
    "401": 31100,
    "404": 11000,
    "500": 2000
  },
  "top_routes": [
    {"path": "/api/v1/health", "count": 320000},
    {"path": "/api/v1/items", "count": 210000},
    {"path": "/api/v1/auth/login", "count": 180000},
    {"path": "/assets/main.css", "count": 82100},
    {"path": "/api/v1/users", "count": 50000}
  ]
}
```

---

## 6. Exit Codes

| Exit Code | Meaning |
| :--- | :--- |
| `0` | Execution successful (matched lines may be 0). |
| `1` | Command-line argument validation error. |
| `2` | I/O error (file not found, permission denied, pipe closed). |
| `3` | Parsing failure encountered while running with `--strict`. |

---

## 7. Performance & Resource Constraints

* **Constant Memory Baseline:** All implementations must stream input line-by-line. Ingesting a 10 GB file must not allocate 10 GB of RAM.
* **Latency Percentile Strategy:**
  * Exact calculation via sorting an in-memory array of latencies is acceptable for smaller sets ($< 5\text{M}$ records).
  * *Stretch goal / Advanced:* Implement an online percentile approximation (e.g., T-Digest, P-Square algorithm, or fixed bucket histograms) to maintain strict $O(1)$ auxiliary space.
* **Zero Allocations (C, C++ & Rust):**
  * Avoid allocating a new heap `String` per field on every line. Use views/slices referencing the per-line read buffer (or stack/`char *` views in C).

---

## 8. Repository Structure

```text
.
├── README.md              # This specification
├── data/
│   ├── generate_logs.py   # Synthetic test log generator
│   └── sample.log         # Small log snippet for deterministic tests
├── python/
│   ├── pyproject.toml     # uv project configuration
│   ├── .python-version
│   ├── uv.lock
│   └── src/logriff/
│       ├── __init__.py
│       ├── __main__.py
│       └── cli.py
├── go/
│   ├── go.mod
│   └── main.go
├── rust/
│   ├── Cargo.toml
│   └── src/main.rs
├── cpp/
│   ├── CMakeLists.txt
│   └── src/main.cpp
├── lua/
│   └── logriff.lua
├── typescript/
│   ├── Justfile           # npm/tsc tasks
│   ├── package.json       # Node/TypeScript project metadata
│   ├── package-lock.json  # Pinned npm dependencies
│   ├── tsconfig.json      # Strict TypeScript compiler settings
│   ├── README.md          # Scaffold status and implementation notes
│   └── src/logriff.ts     # Placeholder CLI entry point
├── c/
│   ├── Justfile           # cc/CMake tasks
│   ├── CMakeLists.txt     # C11 build configuration
│   ├── README.md          # Scaffold status and implementation notes
│   └── src/main.c         # Placeholder CLI entry point
└── tests/
    └── integration_test.sh # Cross-language output verification
```

Every implementation directory has its own `Justfile`, exposed through the root `Justfile` (for example, `just typescript check` and `just c build`).

---

## 9. Verification & Test Protocol

1. **Synthetic Data Generation:**
   Use `data/generate_logs.py` to create deterministic test data:
   ```bash
   python3 data/generate_logs.py --lines 1000000 --seed 42 --out data/test_1m.log
   ```

2. **Cross-Language Differential Testing:**
   Pipe identical input through each language's binary with `--format json` and compare outputs using `jq`:
   ```bash
   # Go vs Rust
   ./go/logriff --format json data/test_1m.log > go_out.json
   ./rust/target/release/logriff --format json data/test_1m.log > rust_out.json
   diff <(jq -S . go_out.json) <(jq -S . rust_out.json)
   ```

   Once the TypeScript and C ports are implemented, build them (`just typescript build`, `just c build`) and include their `--format json` output in the same differential check.

3. **Throughput Benchmark:**
   Benchmark line ingestion speed and resident memory:
   ```bash
   /usr/bin/time -v ./your_binary data/test_1m.log > /dev/null
   ```