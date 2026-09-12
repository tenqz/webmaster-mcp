package mcpserver

import "github.com/tenqz/yandex-webmaster-mcp/internal/webmaster"

// HostInput identifies a Webmaster site.
type HostInput struct {
	HostID string `json:"host_id" jsonschema:"Site identifier, e.g. https:example.com:443"`
}

// DateRangeInput is a host plus optional YYYY-MM-DD bounds.
type DateRangeInput struct {
	HostID   string `json:"host_id" jsonschema:"Site identifier, e.g. https:example.com:443"`
	DateFrom string `json:"date_from,omitempty" jsonschema:"Inclusive start YYYY-MM-DD"`
	DateTo   string `json:"date_to,omitempty" jsonschema:"Inclusive end YYYY-MM-DD"`
}

// PopularQueriesInput lists search queries that brought traffic to a host.
type PopularQueriesInput struct {
	HostID     string `json:"host_id" jsonschema:"Site identifier, e.g. https:example.com:443"`
	OrderBy    string `json:"order_by,omitempty" jsonschema:"TOTAL_SHOWS or TOTAL_CLICKS"`
	DeviceType string `json:"device_type,omitempty" jsonschema:"ALL, DESKTOP, MOBILE, TABLET or MOBILE_AND_TABLET"`
	DateFrom   string `json:"date_from,omitempty" jsonschema:"Inclusive start YYYY-MM-DD"`
	DateTo     string `json:"date_to,omitempty" jsonschema:"Inclusive end YYYY-MM-DD"`
	Limit      int    `json:"limit,omitempty" jsonschema:"Number of results, 1-500, default 100"`
	Offset     int    `json:"offset,omitempty" jsonschema:"Offset for pagination"`
}

// QueryHistoryInput is aggregated or per-query history.
type QueryHistoryInput struct {
	HostID     string `json:"host_id" jsonschema:"Site identifier, e.g. https:example.com:443"`
	QueryID    string `json:"query_id,omitempty" jsonschema:"Query identifier from get_popular_queries"`
	DeviceType string `json:"device_type,omitempty" jsonschema:"ALL, DESKTOP, MOBILE, TABLET or MOBILE_AND_TABLET"`
	DateFrom   string `json:"date_from,omitempty" jsonschema:"Inclusive start YYYY-MM-DD"`
	DateTo     string `json:"date_to,omitempty" jsonschema:"Inclusive end YYYY-MM-DD"`
}

// QueryAnalyticsInput is the query↔URL intersection report.
type QueryAnalyticsInput struct {
	HostID              string         `json:"host_id" jsonschema:"Site identifier, e.g. https:example.com:443"`
	TextIndicator       string         `json:"text_indicator,omitempty" jsonschema:"Group by URL or QUERY"`
	DeviceTypeIndicator string         `json:"device_type_indicator,omitempty" jsonschema:"ALL, DESKTOP, MOBILE, TABLET or MOBILE_AND_TABLET"`
	RegionIDs           []int64        `json:"region_ids,omitempty" jsonschema:"Yandex region IDs from get_region_ids"`
	Limit               int            `json:"limit,omitempty" jsonschema:"Number of rows, 1-500, default 20"`
	Offset              int            `json:"offset,omitempty" jsonschema:"Offset for pagination"`
	Filters             map[string]any `json:"filters,omitempty" jsonschema:"Optional API filters object"`
	SortByDate          map[string]any `json:"sort_by_date,omitempty" jsonschema:"Optional sort_by_date object"`
}

// SampleInput pages samples, sitemaps or recrawl tasks.
type SampleInput struct {
	HostID   string `json:"host_id" jsonschema:"Site identifier, e.g. https:example.com:443"`
	ID       string `json:"id,omitempty" jsonschema:"Resource identifier such as sitemap_id or task_id"`
	URL      string `json:"url,omitempty" jsonschema:"Absolute HTTP(S) URL"`
	Limit    int    `json:"limit,omitempty" jsonschema:"Number of results"`
	Offset   int    `json:"offset,omitempty" jsonschema:"Offset for pagination"`
	DateFrom string `json:"date_from,omitempty" jsonschema:"Inclusive start YYYY-MM-DD"`
	DateTo   string `json:"date_to,omitempty" jsonschema:"Inclusive end YYYY-MM-DD"`
}

// RecrawlInput queues one URL for reindexing.
type RecrawlInput struct {
	HostID string `json:"host_id" jsonschema:"Site identifier, e.g. https:example.com:443"`
	URL    string `json:"url" jsonschema:"Absolute HTTP(S) URL to recrawl"`
}

// UserOutput is the stable result of get_user.
type UserOutput struct {
	UserID int64 `json:"user_id"`
}

// HostsOutput is the stable result of list_hosts.
type HostsOutput struct {
	Hosts []webmaster.Host `json:"hosts"`
	Count int              `json:"count"`
}

// RegionsOutput is a static region catalog.
type RegionsOutput struct {
	Regions []webmaster.Region `json:"regions"`
	Count   int                `json:"count"`
}
