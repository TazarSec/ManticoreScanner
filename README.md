# ManticoreScanner

CLI tool that scans npm dependencies for security risks using the ManticoreEngine behavioral analysis backend.

## Install

```bash
go install github.com/KeisaritGroup/ManticoreScanner/cmd/manticore@latest
```

Or build from source:

```bash
git clone https://github.com/KeisaritGroup/ManticoreScanner.git
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

## GitHub Action

Use the published action to install and run the scanner in a workflow. The action
downloads the matching release binary for the runner platform, verifies its
SHA-256 checksum, and invokes `manticore scan`.

```yaml
permissions:
  contents: read
  pull-requests: write

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: KeisaritGroup/ManticoreScanner@v1
        with:
          api-key: ${{ secrets.MANTICORE_API_KEY }}
          fail-on: 50
          vcs-comment: true
```

Pin to an exact release (`@v1.2.3`) for reproducible builds, or `@v1` to track
the latest release in the v1 line.

### Action inputs

| Input | Description |
|---|---|
| `api-key` | Manticore API key (required). |
| `api-url` | Override the API base URL. |
| `file` | Path to `package.json` / `package-lock.json`. Auto-detected if empty. |
| `format` | Output format: `table`, `json`, `sarif`. |
| `output` | Write results to this path instead of stdout. |
| `fail-on` | Fail the job if any suspicion score is at or above this threshold. |
| `production` | Set to `true` to skip devDependencies. |
| `vcs-comment` | Set to `true` to post a PR comment with findings. |
| `version` | Pin a specific release tag (defaults to the ref used to reference the action). |
| `working-directory` | Directory to run the scan from. |

## Docker

```bash
docker run --rm \
  -v $(pwd):/workspace \
  -w /workspace \
  ghcr.io/KeisaritGroup/manticore-scanner:latest \
  scan --api-key YOUR_API_KEY
```
