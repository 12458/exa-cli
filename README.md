# exa

A command-line interface for the [Exa API](https://exa.ai) - search the web and extract content directly from your terminal.

## Installation

### Quick Install (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/12458/exa-cli/main/install.sh | bash
```

### Homebrew

```bash
brew install 12458/tap/exa
```

### Go Install

```bash
go install github.com/12458/exa-cli@latest
```

### Manual Download

Download the latest binary from [GitHub Releases](https://github.com/12458/exa-cli/releases).

## Configuration

Get your API key from [exa.ai](https://exa.ai) and configure the CLI:

```bash
exa configure
```

Or set the environment variable:

```bash
export EXA_API_KEY="your-api-key"
```

## Usage

### Search the Web

```bash
# Basic search
exa search "latest AI news"

# Limit results
exa search -n 5 "golang best practices"

# Filter by domain
exa search -i github.com -i stackoverflow.com "error handling"

# Filter by category and recency
exa search -c news --max-age-hours 24 "tech layoffs"

# Include AI-generated summaries
exa search -s "machine learning tutorials"

# Include full text content
exa search --text "climate change research"
```

### Get Content from URLs

```bash
# Extract content from a URL
exa contents https://example.com

# Multiple URLs with summaries
exa contents -s https://example.com https://another.com

# Output just the text (for piping)
exa contents -q https://example.com | head -100
```

### Output Formats

```bash
# Table (default)
exa search "query"

# JSON
exa search -o json "query"

# Quiet mode (URLs only)
exa search -q "query"
```

## Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `search` | `s` | Search the web using Exa |
| `contents` | `c` | Get contents from URLs |
| `configure` | | Set up API key |
| `completion` | | Generate shell completions |
| `version` | | Show version info |

## Search Flags

| Flag | Alias | Description |
|------|-------|-------------|
| `--type` | `-t` | Search type: `auto`, `fast` |
| `--num-results` | `-n` | Number of results (1-100) |
| `--include-domains` | `-i` | Only include these domains |
| `--exclude-domains` | `-x` | Exclude these domains |
| `--category` | `-c` | Filter by category |
| `--start-published-date` | | Start date (ISO 8601) |
| `--end-published-date` | | End date (ISO 8601) |
| `--max-age-hours` | | Maximum content age |
| `--text` | | Include full text |
| `--summary` | `-s` | Include AI summary |
| `--highlights` | `-H` | Include highlights |

## Contents Flags

| Flag | Alias | Description |
|------|-------|-------------|
| `--text` | `-t` | Include full text (default: true) |
| `--summary` | `-s` | Include AI summary |
| `--highlights` | `-H` | Include highlights |
| `--subpages` | `-p` | Number of subpages to crawl |
| `--context` | `-C` | Combine results for RAG |

## Global Flags

| Flag | Alias | Description |
|------|-------|-------------|
| `--api-key` | | Exa API key |
| `--output` | `-o` | Output format: `table`, `json`, `toon` |
| `--quiet` | `-q` | Quiet mode for scripting |

## Shell Completions

```bash
# Bash
exa completion bash > /etc/bash_completion.d/exa

# Zsh
exa completion zsh > "${fpath[1]}/_exa"

# Fish
exa completion fish > ~/.config/fish/completions/exa.fish
```

## Examples

### Research Workflow

```bash
# Find recent papers on a topic
exa search -c "research paper" -n 20 --max-age-hours 720 "transformer architecture"

# Get summaries of top results
exa search -s -n 5 "transformer architecture improvements 2024"
```

### Content Extraction

```bash
# Extract article text for processing
exa contents -q https://blog.example.com/article | wc -w

# Get structured JSON for downstream tools
exa contents -o json -s https://example.com | jq '.results[0].summary'
```

### Domain-Specific Search

```bash
# Search only documentation sites
exa search -i docs.python.org -i golang.org "context cancellation"

# Exclude social media
exa search -x twitter.com -x reddit.com "product reviews"
```

## API Key Priority

1. `--api-key` flag
2. `EXA_API_KEY` environment variable
3. Config file (`~/.config/exa/config.yaml`)

## License

MIT
