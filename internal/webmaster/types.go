package webmaster

import "context"

// Payload is a JSON object from Yandex Webmaster or a demo fixture.
// Tools pass it through as structured MCP content without re-shaping wire fields.
type Payload map[string]any

// Region is a Yandex geo identifier used by query analytics and feeds.
type Region struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// User is the OAuth token owner required on every Webmaster path.
type User struct {
	UserID int64 `json:"user_id"`
}

// Host is a site registered in Webmaster.
type Host struct {
	HostID         string `json:"host_id"`
	ASCIIHostURL   string `json:"ascii_host_url,omitempty"`
	UnicodeHostURL string `json:"unicode_host_url,omitempty"`
	Verified       bool   `json:"verified"`
	HostDataStatus string `json:"host_data_status,omitempty"`
}

// DateRange selects an optional inclusive history window for a host.
type DateRange struct {
	HostID   string
	DateFrom string
	DateTo   string
}

// PopularQuery lists search queries that brought traffic to a host.
type PopularQuery struct {
	HostID     string
	OrderBy    string
	DeviceType string
	DateFrom   string
	DateTo     string
	Limit      int
	Offset     int
}

// QueryHistory is aggregated search-query statistics over time.
type QueryHistory struct {
	HostID     string
	QueryID    string
	DeviceType string
	DateFrom   string
	DateTo     string
}

// QueryAnalytics is the query↔URL intersection report for recent days.
type QueryAnalytics struct {
	HostID              string
	TextIndicator       string
	DeviceTypeIndicator string
	RegionIDs           []int64
	Limit               int
	Offset              int
	Filters             map[string]any
	SortByDate          map[string]any
}

// SampleQuery pages sample URLs or sitemaps for a host.
type SampleQuery struct {
	HostID   string
	ID       string
	URL      string
	Limit    int
	Offset   int
	DateFrom string
	DateTo   string
}

// RecrawlRequest queues one URL for reindexing.
type RecrawlRequest struct {
	HostID string
	URL    string
}

// Webmaster is the capability used by MCP tools.
// Live code talks to Yandex; tests and demo provide in-memory implementations.
type Webmaster interface {
	GetUser(ctx context.Context) (User, error)
	ListHosts(ctx context.Context) ([]Host, error)
	GetHost(ctx context.Context, hostID string) (Payload, error)
	GetSummary(ctx context.Context, hostID string) (Payload, error)
	GetSQIHistory(ctx context.Context, query DateRange) (Payload, error)
	GetDiagnostics(ctx context.Context, hostID string) (Payload, error)
	GetPopularQueries(ctx context.Context, query PopularQuery) (Payload, error)
	GetQueryHistory(ctx context.Context, query QueryHistory) (Payload, error)
	GetSingleQueryHistory(ctx context.Context, query QueryHistory) (Payload, error)
	GetQueryAnalytics(ctx context.Context, query QueryAnalytics) (Payload, error)
	GetIndexingHistory(ctx context.Context, query DateRange) (Payload, error)
	GetIndexingSamples(ctx context.Context, query DateRange) (Payload, error)
	GetInSearchHistory(ctx context.Context, query DateRange) (Payload, error)
	GetInSearchSamples(ctx context.Context, query DateRange) (Payload, error)
	GetSearchEventsHistory(ctx context.Context, query DateRange) (Payload, error)
	GetSearchEventsSamples(ctx context.Context, query DateRange) (Payload, error)
	GetExternalLinks(ctx context.Context, query SampleQuery) (Payload, error)
	GetExternalLinksHistory(ctx context.Context, query DateRange) (Payload, error)
	GetBrokenInternalLinks(ctx context.Context, query SampleQuery) (Payload, error)
	GetBrokenInternalLinksHistory(ctx context.Context, query DateRange) (Payload, error)
	GetSitemaps(ctx context.Context, query SampleQuery) (Payload, error)
	GetSitemap(ctx context.Context, query SampleQuery) (Payload, error)
	GetUserSitemaps(ctx context.Context, query SampleQuery) (Payload, error)
	GetUserSitemap(ctx context.Context, query SampleQuery) (Payload, error)
	GetImportantURLs(ctx context.Context, query SampleQuery) (Payload, error)
	GetImportantURLHistory(ctx context.Context, query SampleQuery) (Payload, error)
	GetRecrawlQuota(ctx context.Context, hostID string) (Payload, error)
	GetRecrawlQueue(ctx context.Context, query SampleQuery) (Payload, error)
	GetRecrawlTask(ctx context.Context, query SampleQuery) (Payload, error)
	SubmitRecrawl(ctx context.Context, query RecrawlRequest) (Payload, error)
	GetFeeds(ctx context.Context, hostID string) (Payload, error)
	GetFeedStatus(ctx context.Context, query SampleQuery) (Payload, error)
	ListRegionIDs() []Region
	ListFeedRegions() []Region
}
