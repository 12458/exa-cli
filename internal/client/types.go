package client

// APIError represents an error response from the Exa API
type APIError struct {
	Error string `json:"error"`
}

// TextOptions configures text content retrieval
type TextOptions struct {
	MaxCharacters   int    `json:"maxCharacters,omitempty"`
	IncludeHtmlTags bool   `json:"includeHtmlTags,omitempty"`
	Verbosity       string `json:"verbosity,omitempty"` // compact, standard, full
}

// SummaryOptions configures summary generation
type SummaryOptions struct {
	Query  string `json:"query,omitempty"`
	Schema any    `json:"schema,omitempty"` // JSON schema for structured extraction
}

// ContextOptions configures context output for RAG
type ContextOptions struct {
	MaxCharacters int `json:"maxCharacters,omitempty"`
}

// ContentsOptions specifies what content to retrieve
type ContentsOptions struct {
	Text       any `json:"text,omitempty"`       // bool or TextOptions
	Highlights any `json:"highlights,omitempty"` // bool
	Summary    any `json:"summary,omitempty"`    // bool or SummaryOptions
}

// SearchRequest represents a search API request
type SearchRequest struct {
	Query              string           `json:"query"`
	Type               string           `json:"type,omitempty"`
	NumResults         int              `json:"numResults,omitempty"`
	Contents           *ContentsOptions `json:"contents,omitempty"`
	IncludeDomains     []string         `json:"includeDomains,omitempty"`
	ExcludeDomains     []string         `json:"excludeDomains,omitempty"`
	StartPublishedDate string           `json:"startPublishedDate,omitempty"`
	EndPublishedDate   string           `json:"endPublishedDate,omitempty"`
	Category           string           `json:"category,omitempty"`
	MaxAgeHours        *int             `json:"maxAgeHours,omitempty"`
}

// ContentsRequest represents a contents API request
type ContentsRequest struct {
	IDs              []string `json:"ids"`
	Text             any      `json:"text,omitempty"`       // bool or TextOptions
	Highlights       any      `json:"highlights,omitempty"` // bool
	Summary          any      `json:"summary,omitempty"`    // bool or SummaryOptions
	Context          any      `json:"context,omitempty"`    // bool or ContextOptions
	Subpages         int      `json:"subpages,omitempty"`
	SubpageTarget    []string `json:"subpageTarget,omitempty"`
	MaxAgeHours      *int     `json:"maxAgeHours,omitempty"`
	LivecrawlTimeout int      `json:"livecrawlTimeout,omitempty"`
}

// SearchResult represents a single search result
type SearchResult struct {
	Title         string   `json:"title"`
	URL           string   `json:"url"`
	PublishedDate string   `json:"publishedDate,omitempty"`
	Author        string   `json:"author,omitempty"`
	Score         float64  `json:"score,omitempty"`
	ID            string   `json:"id"`
	Text          string   `json:"text,omitempty"`
	Highlights    []string `json:"highlights,omitempty"`
	Summary       string   `json:"summary,omitempty"`
}

// SearchResponse represents the response from search and find-similar APIs
type SearchResponse struct {
	Results            []SearchResult `json:"results"`
	AutopromptString   string         `json:"autopromptString,omitempty"`
	ResolvedSearchType string         `json:"resolvedSearchType,omitempty"`
}

// ContentStatus represents the status of a content fetch
type ContentStatus struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Error  *struct {
		Tag            string `json:"tag,omitempty"`
		HTTPStatusCode int    `json:"httpStatusCode,omitempty"`
	} `json:"error,omitempty"`
}

// ContentsResponse represents the response from the contents API
type ContentsResponse struct {
	Results  []SearchResult  `json:"results"`
	Statuses []ContentStatus `json:"statuses,omitempty"`
}
