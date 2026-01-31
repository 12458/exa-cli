package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"12458/exa-cli/internal/client"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "exa",
		Usage: "CLI tool for the Exa API",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "api-key",
				Usage:   "Exa API key (overrides EXA_API_KEY environment variable)",
				Sources: cli.EnvVars("EXA_API_KEY"),
			},
		},
		Commands: []*cli.Command{
			searchCmd(),
			contentsCmd(),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func searchCmd() *cli.Command {
	return &cli.Command{
		Name:  "search",
		Usage: "Search the web using Exa",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "query",
				Aliases:  []string{"q"},
				Usage:    "Search query",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "type",
				Usage: "Search type: auto, fast",
				Value: "auto",
			},
			&cli.IntFlag{
				Name:  "num-results",
				Usage: "Number of results (1-100)",
				Value: 10,
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
				Name:  "highlights",
				Usage: "Include highlights",
			},
			&cli.BoolFlag{
				Name:  "summary",
				Usage: "Include AI-generated summary",
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
				Name:  "include-domains",
				Usage: "Only include results from these domains",
			},
			&cli.StringSliceFlag{
				Name:  "exclude-domains",
				Usage: "Exclude results from these domains",
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
				Name:  "category",
				Usage: "Content category: company, people, tweet, news, research paper, personal site, financial report",
			},
			&cli.IntFlag{
				Name:  "max-age-hours",
				Usage: "Maximum age of content in hours (0=always livecrawl, -1=cache only)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c, err := client.New(cmd.Root().String("api-key"))
			if err != nil {
				return err
			}

			req := &client.SearchRequest{
				Query:      cmd.String("query"),
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

			return printJSON(result)
		},
	}
}

func contentsCmd() *cli.Command {
	return &cli.Command{
		Name:  "contents",
		Usage: "Get contents from URLs",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "url",
				Aliases:  []string{"u"},
				Usage:    "URLs to fetch content from (can specify multiple)",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "text",
				Usage: "Include full text content",
				Value: true,
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
				Name:  "highlights",
				Usage: "Include highlights",
			},
			&cli.BoolFlag{
				Name:  "summary",
				Usage: "Include AI-generated summary",
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
				Name:  "subpages",
				Usage: "Number of subpages to crawl",
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
				Name:  "context",
				Usage: "Return all results combined into a single string for RAG",
			},
			&cli.IntFlag{
				Name:  "context-max-chars",
				Usage: "Maximum characters for context string",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			c, err := client.New(cmd.Root().String("api-key"))
			if err != nil {
				return err
			}

			req := &client.ContentsRequest{
				IDs: cmd.StringSlice("url"),
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

			return printJSON(result)
		},
	}
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
