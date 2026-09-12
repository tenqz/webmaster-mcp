package webmaster

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

const defaultBaseURL = "https://api.webmaster.yandex.net/v4"

// Client implements Webmaster against the Yandex Webmaster REST API.
type Client struct {
	http    *http.Client
	token   string
	baseURL string
	options Options
	slots   chan struct{}
	mu      sync.Mutex
	userID  int64
	loading chan struct{}
}

// user shares successful discovery without holding a mutex during network I/O.
func (c *Client) user(ctx context.Context) (int64, error) {
	for {
		if err := ctx.Err(); err != nil {
			return 0, contextFailure(err)
		}
		c.mu.Lock()
		if c.userID != 0 {
			id := c.userID
			c.mu.Unlock()
			return id, nil
		}
		if pending := c.loading; pending != nil {
			c.mu.Unlock()
			select {
			case <-ctx.Done():
				return 0, contextFailure(ctx.Err())
			case <-pending:
				continue
			}
		}
		pending := make(chan struct{})
		c.loading = pending
		c.mu.Unlock()
		var user User
		err := c.getJSON(ctx, c.baseURL+"/user/", &user)
		if err == nil && user.UserID == 0 {
			err = &RequestError{Kind: "invalid_response"}
		}
		c.mu.Lock()
		if err == nil {
			c.userID = user.UserID
		}
		c.loading = nil
		close(pending)
		c.mu.Unlock()
		if err != nil {
			return 0, fmt.Errorf("get user: %w", err)
		}
		return user.UserID, nil
	}
}

func (c *Client) hostURL(ctx context.Context, hostID, suffix string, query url.Values) (string, error) {
	userID, err := c.user(ctx)
	if err != nil {
		return "", err
	}
	endpoint := fmt.Sprintf("%s/user/%d/hosts/%s%s", c.baseURL, userID, url.PathEscape(hostID), suffix)
	if encoded := query.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}
	return endpoint, nil
}

func (c *Client) hostGet(ctx context.Context, hostID, suffix string, query url.Values) (Payload, error) {
	ctx, cancel := context.WithTimeout(ctx, c.options.defaults().RequestTimeout)
	defer cancel()
	endpoint, err := c.hostURL(ctx, hostID, suffix, query)
	if err != nil {
		return nil, err
	}
	var payload Payload
	if err := c.getJSON(ctx, endpoint, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *Client) hostPost(ctx context.Context, hostID, suffix string, body any) (Payload, error) {
	ctx, cancel := context.WithTimeout(ctx, c.options.defaults().RequestTimeout)
	defer cancel()
	endpoint, err := c.hostURL(ctx, hostID, suffix, nil)
	if err != nil {
		return nil, err
	}
	var payload Payload
	if err := c.postJSON(ctx, endpoint, body, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func dateQuery(from, to string) url.Values {
	query := url.Values{}
	if from != "" {
		query.Set("date_from", RFC3339Date(from))
	}
	if to != "" {
		query.Set("date_to", RFC3339Date(to))
	}
	return query
}

func queryIndicators(values url.Values) {
	values.Add("query_indicator", "TOTAL_SHOWS")
	values.Add("query_indicator", "TOTAL_CLICKS")
	values.Add("query_indicator", "AVG_SHOW_POSITION")
}

func setDevice(values url.Values, device string) {
	if device != "" {
		values.Set("device_type_indicator", device)
	}
}

func setPaging(values url.Values, limit, offset int) {
	if limit > 0 {
		values.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		values.Set("offset", strconv.Itoa(offset))
	}
}

// GetUser returns the OAuth token owner.
func (c *Client) GetUser(ctx context.Context) (User, error) {
	ctx, cancel := context.WithTimeout(ctx, c.options.defaults().RequestTimeout)
	defer cancel()
	userID, err := c.user(ctx)
	if err != nil {
		return User{}, err
	}
	return User{UserID: userID}, nil
}

// ListHosts returns sites added to Webmaster for the token owner.
func (c *Client) ListHosts(ctx context.Context) ([]Host, error) {
	ctx, cancel := context.WithTimeout(ctx, c.options.defaults().RequestTimeout)
	defer cancel()
	userID, err := c.user(ctx)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Hosts []Host `json:"hosts"`
	}
	if err := c.getJSON(ctx, fmt.Sprintf("%s/user/%d/hosts", c.baseURL, userID), &payload); err != nil {
		return nil, fmt.Errorf("list hosts: %w", err)
	}
	return payload.Hosts, nil
}

// GetHost returns detailed information about one site.
func (c *Client) GetHost(ctx context.Context, hostID string) (Payload, error) {
	return c.hostGet(ctx, hostID, "", nil)
}

// GetSummary returns SQI, indexed pages and problem counts.
func (c *Client) GetSummary(ctx context.Context, hostID string) (Payload, error) {
	return c.hostGet(ctx, hostID, "/summary", nil)
}

// GetSQIHistory returns Site Quality Index points over time.
func (c *Client) GetSQIHistory(ctx context.Context, query DateRange) (Payload, error) {
	return c.hostGet(ctx, query.HostID, "/sqi-history", dateQuery(query.DateFrom, query.DateTo))
}

// GetDiagnostics returns site problems grouped by severity.
func (c *Client) GetDiagnostics(ctx context.Context, hostID string) (Payload, error) {
	return c.hostGet(ctx, hostID, "/diagnostics", nil)
}

// GetPopularQueries returns search queries that brought traffic to the site.
func (c *Client) GetPopularQueries(ctx context.Context, query PopularQuery) (Payload, error) {
	values := dateQuery(query.DateFrom, query.DateTo)
	values.Set("order_by", query.OrderBy)
	queryIndicators(values)
	values.Add("query_indicator", "AVG_CLICK_POSITION")
	setDevice(values, query.DeviceType)
	setPaging(values, query.Limit, query.Offset)
	return c.hostGet(ctx, query.HostID, "/search-queries/popular", values)
}

// GetQueryHistory returns aggregated statistics for all queries over time.
func (c *Client) GetQueryHistory(ctx context.Context, query QueryHistory) (Payload, error) {
	values := dateQuery(query.DateFrom, query.DateTo)
	queryIndicators(values)
	setDevice(values, query.DeviceType)
	return c.hostGet(ctx, query.HostID, "/search-queries/all/history", values)
}

// GetSingleQueryHistory returns statistics over time for one query id.
func (c *Client) GetSingleQueryHistory(ctx context.Context, query QueryHistory) (Payload, error) {
	values := dateQuery(query.DateFrom, query.DateTo)
	queryIndicators(values)
	setDevice(values, query.DeviceType)
	return c.hostGet(ctx, query.HostID, "/search-queries/"+url.PathEscape(query.QueryID)+"/history", values)
}

// GetQueryAnalytics returns the query↔URL intersection report.
func (c *Client) GetQueryAnalytics(ctx context.Context, query QueryAnalytics) (Payload, error) {
	body := map[string]any{
		"text_indicator":        query.TextIndicator,
		"device_type_indicator": query.DeviceTypeIndicator,
		"limit":                 query.Limit,
		"offset":                query.Offset,
	}
	if len(query.RegionIDs) > 0 {
		body["region_ids"] = query.RegionIDs
	}
	if query.Filters != nil {
		body["filters"] = query.Filters
	}
	if query.SortByDate != nil {
		body["sort_by_date"] = query.SortByDate
	}
	return c.hostPost(ctx, query.HostID, "/query-analytics/list", body)
}

// GetIndexingHistory returns pages downloaded by the robot, grouped by HTTP status.
func (c *Client) GetIndexingHistory(ctx context.Context, query DateRange) (Payload, error) {
	return c.hostGet(ctx, query.HostID, "/indexing/history", dateQuery(query.DateFrom, query.DateTo))
}

// GetIndexingSamples returns examples of downloaded pages.
func (c *Client) GetIndexingSamples(ctx context.Context, query DateRange) (Payload, error) {
	return c.hostGet(ctx, query.HostID, "/indexing/samples", dateQuery(query.DateFrom, query.DateTo))
}

// GetInSearchHistory returns pages in search results over time.
func (c *Client) GetInSearchHistory(ctx context.Context, query DateRange) (Payload, error) {
	return c.hostGet(ctx, query.HostID, "/search-urls/in-search/history", dateQuery(query.DateFrom, query.DateTo))
}

// GetInSearchSamples returns examples of pages in search.
func (c *Client) GetInSearchSamples(ctx context.Context, query DateRange) (Payload, error) {
	return c.hostGet(ctx, query.HostID, "/search-urls/in-search/samples", dateQuery(query.DateFrom, query.DateTo))
}

// GetSearchEventsHistory returns pages added to or removed from search.
func (c *Client) GetSearchEventsHistory(ctx context.Context, query DateRange) (Payload, error) {
	return c.hostGet(ctx, query.HostID, "/search-urls/events/history", dateQuery(query.DateFrom, query.DateTo))
}

// GetSearchEventsSamples returns examples of search events.
func (c *Client) GetSearchEventsSamples(ctx context.Context, query DateRange) (Payload, error) {
	return c.hostGet(ctx, query.HostID, "/search-urls/events/samples", dateQuery(query.DateFrom, query.DateTo))
}

// GetExternalLinks returns backlink samples.
func (c *Client) GetExternalLinks(ctx context.Context, query SampleQuery) (Payload, error) {
	values := url.Values{}
	setPaging(values, query.Limit, query.Offset)
	return c.hostGet(ctx, query.HostID, "/links/external/samples", values)
}

// GetExternalLinksHistory returns backlink counts over time.
func (c *Client) GetExternalLinksHistory(ctx context.Context, query DateRange) (Payload, error) {
	return c.hostGet(ctx, query.HostID, "/links/external/history", dateQuery(query.DateFrom, query.DateTo))
}

// GetBrokenInternalLinks returns broken internal link samples.
func (c *Client) GetBrokenInternalLinks(ctx context.Context, query SampleQuery) (Payload, error) {
	values := url.Values{}
	setPaging(values, query.Limit, query.Offset)
	return c.hostGet(ctx, query.HostID, "/links/internal/broken/samples", values)
}

// GetBrokenInternalLinksHistory returns broken internal link counts over time.
func (c *Client) GetBrokenInternalLinksHistory(ctx context.Context, query DateRange) (Payload, error) {
	return c.hostGet(ctx, query.HostID, "/links/internal/broken/history", dateQuery(query.DateFrom, query.DateTo))
}

// GetSitemaps returns sitemap files detected for the site.
func (c *Client) GetSitemaps(ctx context.Context, query SampleQuery) (Payload, error) {
	values := url.Values{}
	setPaging(values, query.Limit, query.Offset)
	return c.hostGet(ctx, query.HostID, "/sitemaps", values)
}

// GetSitemap returns one detected sitemap.
func (c *Client) GetSitemap(ctx context.Context, query SampleQuery) (Payload, error) {
	return c.hostGet(ctx, query.HostID, "/sitemaps/"+url.PathEscape(query.ID), nil)
}

// GetUserSitemaps returns sitemaps added by the user.
func (c *Client) GetUserSitemaps(ctx context.Context, query SampleQuery) (Payload, error) {
	values := url.Values{}
	setPaging(values, query.Limit, query.Offset)
	return c.hostGet(ctx, query.HostID, "/user-added-sitemaps", values)
}

// GetUserSitemap returns one user-added sitemap.
func (c *Client) GetUserSitemap(ctx context.Context, query SampleQuery) (Payload, error) {
	return c.hostGet(ctx, query.HostID, "/user-added-sitemaps/"+url.PathEscape(query.ID), nil)
}

// GetImportantURLs returns monitored important pages.
func (c *Client) GetImportantURLs(ctx context.Context, query SampleQuery) (Payload, error) {
	values := url.Values{}
	setPaging(values, query.Limit, query.Offset)
	return c.hostGet(ctx, query.HostID, "/important-urls", values)
}

// GetImportantURLHistory returns change history for one important URL.
func (c *Client) GetImportantURLHistory(ctx context.Context, query SampleQuery) (Payload, error) {
	values := dateQuery(query.DateFrom, query.DateTo)
	values.Set("url", query.URL)
	return c.hostGet(ctx, query.HostID, "/important-urls/history", values)
}

// GetRecrawlQuota returns the daily reindexing quota.
func (c *Client) GetRecrawlQuota(ctx context.Context, hostID string) (Payload, error) {
	return c.hostGet(ctx, hostID, "/recrawl/quota", nil)
}

// GetRecrawlQueue returns pending or recent recrawl tasks.
func (c *Client) GetRecrawlQueue(ctx context.Context, query SampleQuery) (Payload, error) {
	values := url.Values{}
	setPaging(values, query.Limit, query.Offset)
	return c.hostGet(ctx, query.HostID, "/recrawl/queue", values)
}

// GetRecrawlTask returns one recrawl task.
func (c *Client) GetRecrawlTask(ctx context.Context, query SampleQuery) (Payload, error) {
	return c.hostGet(ctx, query.HostID, "/recrawl/queue/"+url.PathEscape(query.ID), nil)
}

// SubmitRecrawl queues a URL for reindexing and consumes quota.
func (c *Client) SubmitRecrawl(ctx context.Context, query RecrawlRequest) (Payload, error) {
	return c.hostPost(ctx, query.HostID, "/recrawl/queue", map[string]string{"url": query.URL})
}

// GetFeeds returns data feeds loaded for the site.
func (c *Client) GetFeeds(ctx context.Context, hostID string) (Payload, error) {
	return c.hostGet(ctx, hostID, "/feeds/list", nil)
}

// GetFeedStatus returns an async feed upload task.
func (c *Client) GetFeedStatus(ctx context.Context, query SampleQuery) (Payload, error) {
	values := url.Values{}
	if query.ID != "" {
		values.Set("task_id", query.ID)
	}
	return c.hostGet(ctx, query.HostID, "/feeds/add/info", values)
}

// ListRegionIDs returns the static query-analytics region catalog.
func (c *Client) ListRegionIDs() []Region { return append([]Region(nil), CommonRegions...) }

// ListFeedRegions returns the static feed-upload region catalog.
func (c *Client) ListFeedRegions() []Region { return append([]Region(nil), FeedRegions...) }
