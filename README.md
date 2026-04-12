# ManticoreScanner

CLI tool that scans npm dependencies for security risks using the ManticoreEngine behavioral analysis backend.

## Install

```bash
go install github.com/etsubu/manticore-scanner/cmd/manticore@latest
```

Or build from source:

```bash
git clone https://github.com/etsubu/manticore-scanner.git
cd manticore-scanner
go build -o manticore ./cmd/manticore
```

## Usage

### Basic scan

Scan the current directory (auto-detects `package-lock.json` or `package.json`):

```bash
manticore scan --api-key YOUR_API_KEY
```

### Scan a specific file

```bash
manticore scan --api-key YOUR_API_KEY --file path/to/package-lock.json
```

### Output formats

```bash
# Human-readable table (default)
manticore scan --api-key YOUR_API_KEY --format table

# JSON
manticore scan --api-key YOUR_API_KEY --format json

# SARIF (for GitHub Code Scanning)
manticore scan --api-key YOUR_API_KEY --format sarif --output results.sarif
```

### Fail on suspicious packages

Exit with code 1 if any package has a suspicion score at or above the threshold:

```bash
manticore scan --api-key YOUR_API_KEY --fail-on 50
```

### Skip devDependencies

```bash
manticore scan --api-key YOUR_API_KEY --production
```

### Post results to a GitHub PR

When running in GitHub Actions, post a comment with suspicious packages to the PR:

```bash
manticore scan --api-key YOUR_API_KEY --vcs-comment
```

## Environment variables

| Variable | Description                                         |
|---|-----------------------------------------------------|
| `MANTICORE_API_KEY` | API key (alternative to `--api-key`)                |
| `MANTICORE_API_URL` | API base URL (default: `https://api.manticore.com`) |
| `MANTICORE_TIMEOUT` | Polling timeout in seconds (default: `300`)         |
| `MANTICORE_FORMAT` | Output format: `table`, `json`, `sarif`             |

## GitHub Actions example

```yaml
- name: Scan dependencies
  env:
    MANTICORE_API_KEY: ${{ secrets.MANTICORE_API_KEY }}
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  run: manticore scan --fail-on 50 --vcs-comment
```

## Docker

```bash
docker run --rm \
  -v $(pwd):/workspace \
  -w /workspace \
  ghcr.io/etsubu/manticore-scanner:latest \
  scan --api-key YOUR_API_KEY
```
