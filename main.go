package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/12458/exa-cli/internal/client"
	"github.com/12458/exa-cli/internal/config"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/toon-format/toon-go"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

// Version information set by ldflags during build
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd := &cli.Command{
		Name:                  "exa",
		Usage:                 "CLI tool for the Exa API",
		Version:               version,
		DefaultCommand:        "search",
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "api-key",
				Usage:   "Exa API key (overrides EXA_API_KEY environment variable)",
				Sources: cli.EnvVars("EXA_API_KEY"),
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output format: table, json, toon",
				Value:   "table",
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Quiet mode: output only URLs (search) or text (contents) for scripting",
			},
		},
		Commands: []*cli.Command{
			searchCmd(),
			contentsCmd(),
			configureCmd(),
			completionCmd(),
			versionCmd(),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

// isTerminal returns true if stdout is a terminal (not piped)
func isTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// getAPIKey returns the API key from flag, env var, or config file (in that priority order)
func getAPIKey(cmd *cli.Command) string {
	// Check flag/env first (handled by cli library)
	if key := cmd.Root().String("api-key"); key != "" {
		return key
	}
	// Fall back to config file
	return config.GetAPIKey()
}

func searchCmd() *cli.Command {
	return &cli.Command{
		Name:      "search",
		Aliases:   []string{"s"},
		Usage:     "Search the web using Exa",
		ArgsUsage: "<query>",
		UsageText: `Examples:
  exa search "latest AI news"
  exa search -n 5 --summary "golang best practices"
  exa search -i github.com -i stackoverflow.com "error handling"
  exa search -c news --max-age-hours 24 "tech layoffs"`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "type",
				Aliases: []string{"t"},
				Usage:   "Search type: auto, fast",
				Value:   "auto",
			},
			&cli.IntFlag{
				Name:    "num-results",
				Aliases: []string{"n"},
				Usage:   "Number of results (1-100)",
				Value:   10,
			},
			&cli.BoolFlag{
				Name:  "text",
				Usage: "Include full text content",
			},
			&cli.IntFlag{
				Name:  "text-max-chars",
				Usage: "Maximum characters for text content",
			},
			&cli.BoolFlag{
				Name:  "text-include-html",
				Usage: "Include HTML tags in text content",
			},
			&cli.StringFlag{
				Name:  "text-verbosity",
				Usage: "Text verbosity: compact, standard, full",
			},
			&cli.BoolFlag{
				Name:    "highlights",
				Aliases: []string{"H"},
				Usage:   "Include highlights",
			},
			&cli.BoolFlag{
				Name:    "summary",
				Aliases: []string{"s"},
				Usage:   "Include AI-generated summary",
			},
			&cli.StringFlag{
				Name:  "summary-query",
				Usage: "Custom query for summary generation",
			},
			&cli.StringFlag{
				Name:  "summary-schema",
				Usage: "JSON schema for structured summary extraction",
			},
			&cli.StringSliceFlag{
				Name:    "include-domains",
				Aliases: []string{"i"},
				Usage:   "Only include results from these domains",
			},
			&cli.StringSliceFlag{
				Name:    "exclude-domains",
				Aliases: []string{"x"},
				Usage:   "Exclude results from these domains",
			},
			&cli.StringFlag{
				Name:  "start-published-date",
				Usage: "Filter by publish date (ISO 8601)",
			},
			&cli.StringFlag{
				Name:  "end-published-date",
				Usage: "Filter by publish date (ISO 8601)",
			},
			&cli.StringFlag{
				Name:    "category",
				Aliases: []string{"c"},
				Usage:   "Content category: company, people, tweet, news, research paper, personal site, financial report",
			},
			&cli.IntFlag{
				Name:  "max-age-hours",
				Usage: "Maximum age of content in hours (0=always livecrawl, -1=cache only)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return fmt.Errorf("query is required")
			}
			query := cmd.Args().First()

			c, err := client.New(getAPIKey(cmd))
			if err != nil {
				return err
			}

			req := &client.SearchRequest{
				Query:      query,
				Type:       cmd.String("type"),
				NumResults: int(cmd.Int("num-results")),
			}

			// Build contents options
			hasTextOpts := cmd.Bool("text") || cmd.Int("text-max-chars") > 0 || cmd.Bool("text-include-html") || cmd.String("text-verbosity") != ""
			hasSummaryOpts := cmd.Bool("summary") || cmd.String("summary-query") != "" || cmd.String("summary-schema") != ""
			if hasTextOpts || cmd.Bool("highlights") || hasSummaryOpts {
				req.Contents = &client.ContentsOptions{}

				if hasTextOpts {
					if cmd.Int("text-max-chars") > 0 || cmd.Bool("text-include-html") || cmd.String("text-verbosity") != "" {
						req.Contents.Text = &client.TextOptions{
							MaxCharacters:   int(cmd.Int("text-max-chars")),
							IncludeHtmlTags: cmd.Bool("text-include-html"),
							Verbosity:       cmd.String("text-verbosity"),
						}
					} else {
						req.Contents.Text = true
					}
				}
				if cmd.Bool("highlights") {
					req.Contents.Highlights = true
				}
				if hasSummaryOpts {
					if cmd.String("summary-query") != "" || cmd.String("summary-schema") != "" {
						opts := &client.SummaryOptions{Query: cmd.String("summary-query")}
						if schema := cmd.String("summary-schema"); schema != "" {
							var schemaObj any
							if err := json.Unmarshal([]byte(schema), &schemaObj); err != nil {
								return fmt.Errorf("invalid summary-schema JSON: %w", err)
							}
							opts.Schema = schemaObj
						}
						req.Contents.Summary = opts
					} else {
						req.Contents.Summary = true
					}
				}
			}

			if domains := cmd.StringSlice("include-domains"); len(domains) > 0 {
				req.IncludeDomains = domains
			}
			if domains := cmd.StringSlice("exclude-domains"); len(domains) > 0 {
				req.ExcludeDomains = domains
			}
			if date := cmd.String("start-published-date"); date != "" {
				req.StartPublishedDate = date
			}
			if date := cmd.String("end-published-date"); date != "" {
				req.EndPublishedDate = date
			}
			if cat := cmd.String("category"); cat != "" {
				req.Category = cat
			}
			if cmd.IsSet("max-age-hours") {
				hours := int(cmd.Int("max-age-hours"))
				req.MaxAgeHours = &hours
			}

			result, err := c.Search(ctx, req)
			if err != nil {
				return err
			}

			return printOutput(cmd, result)
		},
	}
}

func contentsCmd() *cli.Command {
	return &cli.Command{
		Name:      "contents",
		Aliases:   []string{"c"},
		Usage:     "Get contents from URLs",
		ArgsUsage: "<url> [url...]",
		UsageText: `Examples:
  exa contents https://example.com
  exa contents --summary https://example.com https://another.com
  exa contents -q https://example.com | head -100`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "text",
				Aliases: []string{"t"},
				Usage:   "Include full text content",
				Value:   true,
			},
			&cli.IntFlag{
				Name:  "text-max-chars",
				Usage: "Maximum characters for text content",
			},
			&cli.BoolFlag{
				Name:  "text-include-html",
				Usage: "Include HTML tags in text content",
			},
			&cli.StringFlag{
				Name:  "text-verbosity",
				Usage: "Text verbosity: compact, standard, full",
			},
			&cli.BoolFlag{
				Name:    "highlights",
				Aliases: []string{"H"},
				Usage:   "Include highlights",
			},
			&cli.BoolFlag{
				Name:    "summary",
				Aliases: []string{"s"},
				Usage:   "Include AI-generated summary",
			},
			&cli.StringFlag{
				Name:  "summary-query",
				Usage: "Custom query for summary generation",
			},
			&cli.StringFlag{
				Name:  "summary-schema",
				Usage: "JSON schema for structured summary extraction",
			},
			&cli.IntFlag{
				Name:    "subpages",
				Aliases: []string{"p"},
				Usage:   "Number of subpages to crawl",
			},
			&cli.StringSliceFlag{
				Name:  "subpage-target",
				Usage: "Keywords to target when crawling subpages",
			},
			&cli.IntFlag{
				Name:  "max-age-hours",
				Usage: "Maximum age of content in hours",
			},
			&cli.IntFlag{
				Name:  "livecrawl-timeout",
				Usage: "Timeout in ms for live crawling",
			},
			&cli.BoolFlag{
				Name:    "context",
				Aliases: []string{"C"},
				Usage:   "Return all results combined into a single string for RAG",
			},
			&cli.IntFlag{
				Name:  "context-max-chars",
				Usage: "Maximum characters for context string",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return fmt.Errorf("at least one URL is required")
			}

			c, err := client.New(getAPIKey(cmd))
			if err != nil {
				return err
			}

			req := &client.ContentsRequest{
				IDs: cmd.Args().Slice(),
			}

			// Build text options
			if cmd.Bool("text") || cmd.Int("text-max-chars") > 0 || cmd.Bool("text-include-html") || cmd.String("text-verbosity") != "" {
				if cmd.Int("text-max-chars") > 0 || cmd.Bool("text-include-html") || cmd.String("text-verbosity") != "" {
					req.Text = &client.TextOptions{
						MaxCharacters:   int(cmd.Int("text-max-chars")),
						IncludeHtmlTags: cmd.Bool("text-include-html"),
						Verbosity:       cmd.String("text-verbosity"),
					}
				} else {
					req.Text = true
				}
			}
			if cmd.Bool("highlights") {
				req.Highlights = true
			}
			// Build summary options
			if cmd.Bool("summary") || cmd.String("summary-query") != "" || cmd.String("summary-schema") != "" {
				if cmd.String("summary-query") != "" || cmd.String("summary-schema") != "" {
					opts := &client.SummaryOptions{Query: cmd.String("summary-query")}
					if schema := cmd.String("summary-schema"); schema != "" {
						var schemaObj any
						if err := json.Unmarshal([]byte(schema), &schemaObj); err != nil {
							return fmt.Errorf("invalid summary-schema JSON: %w", err)
						}
						opts.Schema = schemaObj
					}
					req.Summary = opts
				} else {
					req.Summary = true
				}
			}
			if cmd.Int("subpages") > 0 {
				req.Subpages = int(cmd.Int("subpages"))
			}
			if targets := cmd.StringSlice("subpage-target"); len(targets) > 0 {
				req.SubpageTarget = targets
			}
			if cmd.IsSet("max-age-hours") {
				hours := int(cmd.Int("max-age-hours"))
				req.MaxAgeHours = &hours
			}
			if cmd.Int("livecrawl-timeout") > 0 {
				req.LivecrawlTimeout = int(cmd.Int("livecrawl-timeout"))
			}
			// Build context options
			if cmd.Bool("context") || cmd.Int("context-max-chars") > 0 {
				if cmd.Int("context-max-chars") > 0 {
					req.Context = &client.ContextOptions{MaxCharacters: int(cmd.Int("context-max-chars"))}
				} else {
					req.Context = true
				}
			}

			result, err := c.GetContents(ctx, req)
			if err != nil {
				return err
			}

			return printOutput(cmd, result)
		},
	}
}

func configureCmd() *cli.Command {
	return &cli.Command{
		Name:  "configure",
		Usage: "Configure exa CLI settings (API key)",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Print("Enter your Exa API key: ")

			// Read password with masked input
			apiKey, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return fmt.Errorf("failed to read API key: %w", err)
			}
			fmt.Println() // Print newline after hidden input

			key := strings.TrimSpace(string(apiKey))
			if key == "" {
				return fmt.Errorf("API key cannot be empty")
			}

			cfg := &config.Config{APIKey: key}
			if err := config.Save(cfg); err != nil {
				return err
			}

			path, _ := config.Path()
			fmt.Printf("API key saved to %s\n", path)
			return nil
		},
	}
}

func completionCmd() *cli.Command {
	return &cli.Command{
		Name:  "completion",
		Usage: "Generate shell completion scripts",
		Commands: []*cli.Command{
			{
				Name:  "bash",
				Usage: "Generate bash completion script",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Print(bashCompletion)
					return nil
				},
			},
			{
				Name:  "zsh",
				Usage: "Generate zsh completion script",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Print(zshCompletion)
					return nil
				},
			},
			{
				Name:  "fish",
				Usage: "Generate fish completion script",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Print(fishCompletion)
					return nil
				},
			},
		},
	}
}

func versionCmd() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Show detailed version information",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Printf("exa %s\n", version)
			fmt.Printf("  commit: %s\n", commit)
			fmt.Printf("  built:  %s\n", date)
			return nil
		},
	}
}

const bashCompletion = `# bash completion for exa CLI
_exa_completions() {
    local cur prev opts commands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    commands="search contents configure completion version help"
    global_opts="--api-key --output -o --quiet -q --help -h"
    search_opts="--type -t --num-results -n --text --text-max-chars --text-include-html --text-verbosity --highlights -H --summary -s --summary-query --summary-schema --include-domains -i --exclude-domains -x --start-published-date --end-published-date --category -c --max-age-hours"
    contents_opts="--text -t --text-max-chars --text-include-html --text-verbosity --highlights -H --summary -s --summary-query --summary-schema --subpages -p --subpage-target --max-age-hours --livecrawl-timeout --context -C --context-max-chars"

    case "${COMP_WORDS[1]}" in
        search|s)
            COMPREPLY=( $(compgen -W "${search_opts}" -- ${cur}) )
            return 0
            ;;
        contents|c)
            COMPREPLY=( $(compgen -W "${contents_opts}" -- ${cur}) )
            return 0
            ;;
        completion)
            COMPREPLY=( $(compgen -W "bash zsh fish" -- ${cur}) )
            return 0
            ;;
    esac

    if [[ ${cur} == -* ]]; then
        COMPREPLY=( $(compgen -W "${global_opts}" -- ${cur}) )
        return 0
    fi

    COMPREPLY=( $(compgen -W "${commands}" -- ${cur}) )
    return 0
}
complete -F _exa_completions exa
`

const zshCompletion = `#compdef exa

_exa() {
    local -a commands
    commands=(
        'search:Search the web using Exa'
        's:Search the web using Exa'
        'contents:Get contents from URLs'
        'c:Get contents from URLs'
        'configure:Configure exa CLI settings'
        'completion:Generate shell completion scripts'
        'version:Show detailed version information'
        'help:Shows a list of commands or help for one command'
    )

    _arguments -C \
        '--api-key[Exa API key]:key:' \
        '(-o --output)'{-o,--output}'[Output format]:format:(table json toon)' \
        '(-q --quiet)'{-q,--quiet}'[Quiet mode]' \
        '(-h --help)'{-h,--help}'[Show help]' \
        '1:command:->command' \
        '*::arg:->args'

    case "$state" in
        command)
            _describe -t commands 'exa commands' commands
            ;;
        args)
            case "${words[1]}" in
                search|s)
                    _arguments \
                        '(-t --type)'{-t,--type}'[Search type]:type:(auto fast)' \
                        '(-n --num-results)'{-n,--num-results}'[Number of results]:num:' \
                        '--text[Include full text content]' \
                        '--text-max-chars[Max chars for text]:chars:' \
                        '--text-include-html[Include HTML tags]' \
                        '--text-verbosity[Text verbosity]:verbosity:(compact standard full)' \
                        '(-H --highlights)'{-H,--highlights}'[Include highlights]' \
                        '(-s --summary)'{-s,--summary}'[Include AI summary]' \
                        '--summary-query[Custom query for summary]:query:' \
                        '--summary-schema[JSON schema for summary]:schema:' \
                        '*'{-i,--include-domains}'[Include domains]:domain:' \
                        '*'{-x,--exclude-domains}'[Exclude domains]:domain:' \
                        '--start-published-date[Start date]:date:' \
                        '--end-published-date[End date]:date:' \
                        '(-c --category)'{-c,--category}'[Category]:category:(company people tweet news "research paper" "personal site" "financial report")' \
                        '--max-age-hours[Max age in hours]:hours:' \
                        '*:query:'
                    ;;
                contents|c)
                    _arguments \
                        '(-t --text)'{-t,--text}'[Include full text content]' \
                        '--text-max-chars[Max chars for text]:chars:' \
                        '--text-include-html[Include HTML tags]' \
                        '--text-verbosity[Text verbosity]:verbosity:(compact standard full)' \
                        '(-H --highlights)'{-H,--highlights}'[Include highlights]' \
                        '(-s --summary)'{-s,--summary}'[Include AI summary]' \
                        '--summary-query[Custom query for summary]:query:' \
                        '--summary-schema[JSON schema for summary]:schema:' \
                        '(-p --subpages)'{-p,--subpages}'[Number of subpages]:num:' \
                        '*--subpage-target[Subpage target keywords]:keyword:' \
                        '--max-age-hours[Max age in hours]:hours:' \
                        '--livecrawl-timeout[Livecrawl timeout in ms]:timeout:' \
                        '(-C --context)'{-C,--context}'[Return combined context]' \
                        '--context-max-chars[Max chars for context]:chars:' \
                        '*:url:_urls'
                    ;;
                completion)
                    _arguments '1:shell:(bash zsh fish)'
                    ;;
            esac
            ;;
    esac
}

_exa "$@"
`

const fishCompletion = `# fish completion for exa CLI

# Disable file completion by default
complete -c exa -f

# Commands
complete -c exa -n __fish_use_subcommand -a search -d 'Search the web using Exa'
complete -c exa -n __fish_use_subcommand -a s -d 'Search the web using Exa'
complete -c exa -n __fish_use_subcommand -a contents -d 'Get contents from URLs'
complete -c exa -n __fish_use_subcommand -a c -d 'Get contents from URLs'
complete -c exa -n __fish_use_subcommand -a configure -d 'Configure exa CLI settings'
complete -c exa -n __fish_use_subcommand -a completion -d 'Generate shell completion scripts'
complete -c exa -n __fish_use_subcommand -a version -d 'Show detailed version information'
complete -c exa -n __fish_use_subcommand -a help -d 'Shows help'

# Global options
complete -c exa -l api-key -d 'Exa API key'
complete -c exa -s o -l output -d 'Output format' -a 'table json toon'
complete -c exa -s q -l quiet -d 'Quiet mode'
complete -c exa -s h -l help -d 'Show help'

# Search options
complete -c exa -n '__fish_seen_subcommand_from search s' -s t -l type -d 'Search type' -a 'auto fast'
complete -c exa -n '__fish_seen_subcommand_from search s' -s n -l num-results -d 'Number of results'
complete -c exa -n '__fish_seen_subcommand_from search s' -l text -d 'Include full text content'
complete -c exa -n '__fish_seen_subcommand_from search s' -l text-max-chars -d 'Max chars for text'
complete -c exa -n '__fish_seen_subcommand_from search s' -l text-include-html -d 'Include HTML tags'
complete -c exa -n '__fish_seen_subcommand_from search s' -l text-verbosity -d 'Text verbosity' -a 'compact standard full'
complete -c exa -n '__fish_seen_subcommand_from search s' -s H -l highlights -d 'Include highlights'
complete -c exa -n '__fish_seen_subcommand_from search s' -s s -l summary -d 'Include AI summary'
complete -c exa -n '__fish_seen_subcommand_from search s' -l summary-query -d 'Custom query for summary'
complete -c exa -n '__fish_seen_subcommand_from search s' -l summary-schema -d 'JSON schema for summary'
complete -c exa -n '__fish_seen_subcommand_from search s' -s i -l include-domains -d 'Include domains'
complete -c exa -n '__fish_seen_subcommand_from search s' -s x -l exclude-domains -d 'Exclude domains'
complete -c exa -n '__fish_seen_subcommand_from search s' -l start-published-date -d 'Start date'
complete -c exa -n '__fish_seen_subcommand_from search s' -l end-published-date -d 'End date'
complete -c exa -n '__fish_seen_subcommand_from search s' -s c -l category -d 'Category' -a 'company people tweet news "research paper" "personal site" "financial report"'
complete -c exa -n '__fish_seen_subcommand_from search s' -l max-age-hours -d 'Max age in hours'

# Contents options
complete -c exa -n '__fish_seen_subcommand_from contents c' -s t -l text -d 'Include full text content'
complete -c exa -n '__fish_seen_subcommand_from contents c' -l text-max-chars -d 'Max chars for text'
complete -c exa -n '__fish_seen_subcommand_from contents c' -l text-include-html -d 'Include HTML tags'
complete -c exa -n '__fish_seen_subcommand_from contents c' -l text-verbosity -d 'Text verbosity' -a 'compact standard full'
complete -c exa -n '__fish_seen_subcommand_from contents c' -s H -l highlights -d 'Include highlights'
complete -c exa -n '__fish_seen_subcommand_from contents c' -s s -l summary -d 'Include AI summary'
complete -c exa -n '__fish_seen_subcommand_from contents c' -l summary-query -d 'Custom query for summary'
complete -c exa -n '__fish_seen_subcommand_from contents c' -l summary-schema -d 'JSON schema for summary'
complete -c exa -n '__fish_seen_subcommand_from contents c' -s p -l subpages -d 'Number of subpages'
complete -c exa -n '__fish_seen_subcommand_from contents c' -l subpage-target -d 'Subpage target keywords'
complete -c exa -n '__fish_seen_subcommand_from contents c' -l max-age-hours -d 'Max age in hours'
complete -c exa -n '__fish_seen_subcommand_from contents c' -l livecrawl-timeout -d 'Livecrawl timeout in ms'
complete -c exa -n '__fish_seen_subcommand_from contents c' -s C -l context -d 'Return combined context'
complete -c exa -n '__fish_seen_subcommand_from contents c' -l context-max-chars -d 'Max chars for context'

# Completion subcommands
complete -c exa -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish' -d 'Shell type'
`

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func getOutputFormat(cmd *cli.Command) string {
	return cmd.Root().String("output")
}

func isQuietMode(cmd *cli.Command) bool {
	return cmd.Root().Bool("quiet")
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func printSearchTable(cmd *cli.Command, resp *client.SearchResponse) {
	useColor := isTerminal()

	// Disable color globally if not a TTY
	if !useColor {
		color.NoColor = true
	}

	headerFmt := color.New(color.FgWhite, color.Bold).SprintFunc()
	numFmt := color.New(color.FgCyan).SprintFunc()

	// Determine which columns to show based on flags
	showText := cmd.Bool("text")
	showSummary := cmd.Bool("summary") || cmd.String("summary-query") != "" || cmd.String("summary-schema") != ""

	// Build dynamic column headers
	var headers []any
	headers = append(headers, "#", "Title", "URL")
	if showText {
		headers = append(headers, "Text")
	}
	if showSummary {
		headers = append(headers, "Summary")
	}
	if !showText && !showSummary {
		headers = append(headers, "Published")
	}

	tbl := table.New(headers...)
	tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
		return headerFmt(fmt.Sprintf(format, vals...))
	})

	// Use shorter title when showing text/summary columns
	titleMaxLen := 55
	if showText || showSummary {
		titleMaxLen = 40
	}

	for i, r := range resp.Results {
		title := truncate(r.Title, titleMaxLen)
		url := truncate(r.URL, 45)
		num := fmt.Sprintf("%d", i+1)
		if useColor {
			num = numFmt(num)
		}

		// Build row based on columns
		var row []any
		row = append(row, num, title, url)
		if showText {
			text := truncate(r.Text, 60)
			if text == "" {
				text = "-"
			}
			row = append(row, text)
		}
		if showSummary {
			summary := truncate(r.Summary, 60)
			if summary == "" {
				summary = "-"
			}
			row = append(row, summary)
		}
		if !showText && !showSummary {
			date := r.PublishedDate
			if date == "" {
				date = "-"
			}
			row = append(row, date)
		}
		tbl.AddRow(row...)
	}
	tbl.Print()
}

func printSearchQuiet(resp *client.SearchResponse) {
	for _, r := range resp.Results {
		fmt.Println(r.URL)
	}
}

func printContentsMarkdown(resp *client.ContentsResponse) {
	for i, r := range resp.Results {
		if i > 0 {
			fmt.Println()
		}
		fmt.Println("---")
		fmt.Printf("title: %q\n", r.Title)
		fmt.Printf("url: %s\n", r.URL)
		if r.PublishedDate != "" {
			fmt.Printf("date: %q\n", r.PublishedDate)
		}
		if r.Author != "" {
			fmt.Printf("author: %q\n", r.Author)
		}
		fmt.Println("---")
		if r.Text != "" {
			fmt.Println()
			fmt.Println(r.Text)
		}
		if r.Summary != "" {
			fmt.Println()
			fmt.Println("## Summary")
			fmt.Println()
			fmt.Println(r.Summary)
		}
		if len(r.Highlights) > 0 {
			fmt.Println()
			fmt.Println("## Highlights")
			fmt.Println()
			for _, h := range r.Highlights {
				fmt.Printf("- %s\n", h)
			}
		}
	}
}

func printContentsQuiet(resp *client.ContentsResponse) {
	for i, r := range resp.Results {
		if i > 0 {
			fmt.Println()
		}
		if r.Text != "" {
			fmt.Println(r.Text)
		}
	}
}

func printTOON(v any) error {
	encoded, err := toon.Marshal(v, toon.WithLengthMarkers(true))
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(encoded)
	return err
}

func printOutput(cmd *cli.Command, v any) error {
	quiet := isQuietMode(cmd)
	format := getOutputFormat(cmd)

	// Quiet mode overrides format for specific output types
	if quiet {
		switch resp := v.(type) {
		case *client.SearchResponse:
			printSearchQuiet(resp)
			return nil
		case *client.ContentsResponse:
			printContentsQuiet(resp)
			return nil
		}
	}

	switch format {
	case "json":
		return printJSON(v)
	case "toon":
		return printTOON(v)
	default: // "table"
		switch resp := v.(type) {
		case *client.SearchResponse:
			printSearchTable(cmd, resp)
		case *client.ContentsResponse:
			printContentsMarkdown(resp)
		default:
			return printJSON(v)
		}
	}
	return nil
}
