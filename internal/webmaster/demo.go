package webmaster

import "context"

const demoHostID = "https:example.invalid:443"

// NewDemo provides deterministic fixtures without Yandex credentials or network access.
func NewDemo() Webmaster { return demoWebmaster{} }

type demoWebmaster struct{}

func (demoWebmaster) DemoMode() bool { return true }

func (demoWebmaster) GetUser(ctx context.Context) (User, error) {
	if err := ctx.Err(); err != nil {
		return User{}, err
	}
	return User{UserID: 1}, nil
}

func (demoWebmaster) ListHosts(ctx context.Context) ([]Host, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return []Host{{
		HostID:         demoHostID,
		ASCIIHostURL:   "https://example.invalid/",
		UnicodeHostURL: "https://example.invalid/",
		Verified:       true,
		HostDataStatus: "NOT_INDEXED",
	}}, nil
}

func (d demoWebmaster) GetHost(ctx context.Context, hostID string) (Payload, error) {
	if err := demoAccess(ctx, hostID); err != nil {
		return nil, err
	}
	return Payload{"host_id": demoHostID, "unicode_host_url": "https://example.invalid/", "verified": true, "host_data_status": "NOT_INDEXED"}, nil
}

func (d demoWebmaster) GetSummary(ctx context.Context, hostID string) (Payload, error) {
	if err := demoAccess(ctx, hostID); err != nil {
		return nil, err
	}
	return Payload{"sqi": 120, "searchable_pages_count": 12, "excluded_pages_count": 1, "site_problems": map[string]any{"FATAL": 0, "CRITICAL": 0, "POSSIBLE_PROBLEM": 1, "RECOMMENDATION": 0}}, nil
}

func (d demoWebmaster) GetSQIHistory(ctx context.Context, query DateRange) (Payload, error) {
	if err := demoAccess(ctx, query.HostID); err != nil {
		return nil, err
	}
	return Payload{"points": []any{map[string]any{"date": "2026-09-01T00:00:00.000Z", "value": 120}}}, nil
}

func (d demoWebmaster) GetDiagnostics(ctx context.Context, hostID string) (Payload, error) {
	if err := demoAccess(ctx, hostID); err != nil {
		return nil, err
	}
	return Payload{"problems": map[string]any{}}, nil
}

func (d demoWebmaster) GetPopularQueries(ctx context.Context, query PopularQuery) (Payload, error) {
	if err := demoAccess(ctx, query.HostID); err != nil {
		return nil, err
	}
	return Payload{"count": 1, "queries": []any{map[string]any{"query_id": "q1", "query_text": "example search", "indicators": map[string]any{"TOTAL_SHOWS": 2400, "TOTAL_CLICKS": 120, "AVG_SHOW_POSITION": 7.2}}}}, nil
}

func (d demoWebmaster) GetQueryHistory(ctx context.Context, query QueryHistory) (Payload, error) {
	if err := demoAccess(ctx, query.HostID); err != nil {
		return nil, err
	}
	return Payload{"indicators": map[string]any{"TOTAL_SHOWS": []any{map[string]any{"date": "2026-09-01T00:00:00.000Z", "value": 2400}}}}, nil
}

func (d demoWebmaster) GetSingleQueryHistory(ctx context.Context, query QueryHistory) (Payload, error) {
	if err := demoAccess(ctx, query.HostID); err != nil {
		return nil, err
	}
	return Payload{"query_id": query.QueryID, "indicators": map[string]any{"TOTAL_SHOWS": []any{map[string]any{"date": "2026-09-01T00:00:00.000Z", "value": 80}}}}, nil
}

func (d demoWebmaster) GetQueryAnalytics(ctx context.Context, query QueryAnalytics) (Payload, error) {
	if err := demoAccess(ctx, query.HostID); err != nil {
		return nil, err
	}
	return Payload{"text_indicator_to_statistics": []any{map[string]any{"text_indicator": "https://example.invalid/guide", "TOTAL_SHOWS": 100}}}, nil
}

func (d demoWebmaster) GetIndexingHistory(ctx context.Context, query DateRange) (Payload, error) {
	return d.history(ctx, query.HostID, "HTTP_2XX")
}

func (d demoWebmaster) GetIndexingSamples(ctx context.Context, query DateRange) (Payload, error) {
	return d.samples(ctx, query.HostID)
}

func (d demoWebmaster) GetInSearchHistory(ctx context.Context, query DateRange) (Payload, error) {
	return d.history(ctx, query.HostID, "IN_SEARCH")
}

func (d demoWebmaster) GetInSearchSamples(ctx context.Context, query DateRange) (Payload, error) {
	return d.samples(ctx, query.HostID)
}

func (d demoWebmaster) GetSearchEventsHistory(ctx context.Context, query DateRange) (Payload, error) {
	return d.history(ctx, query.HostID, "APPEARED_IN_SEARCH")
}

func (d demoWebmaster) GetSearchEventsSamples(ctx context.Context, query DateRange) (Payload, error) {
	return d.samples(ctx, query.HostID)
}

func (d demoWebmaster) GetExternalLinks(ctx context.Context, query SampleQuery) (Payload, error) {
	return d.samples(ctx, query.HostID)
}

func (d demoWebmaster) GetExternalLinksHistory(ctx context.Context, query DateRange) (Payload, error) {
	return d.history(ctx, query.HostID, "LINKS")
}

func (d demoWebmaster) GetBrokenInternalLinks(ctx context.Context, query SampleQuery) (Payload, error) {
	return d.samples(ctx, query.HostID)
}

func (d demoWebmaster) GetBrokenInternalLinksHistory(ctx context.Context, query DateRange) (Payload, error) {
	return d.history(ctx, query.HostID, "BROKEN")
}

func (d demoWebmaster) GetSitemaps(ctx context.Context, query SampleQuery) (Payload, error) {
	if err := demoAccess(ctx, query.HostID); err != nil {
		return nil, err
	}
	return Payload{"sitemaps": []any{map[string]any{"sitemap_id": "sm1", "sitemap_url": "https://example.invalid/sitemap.xml", "sitemap_type": "SITEMAP", "urls_count": 12, "errors_count": 0}}}, nil
}

func (d demoWebmaster) GetSitemap(ctx context.Context, query SampleQuery) (Payload, error) {
	if err := demoAccess(ctx, query.HostID); err != nil {
		return nil, err
	}
	return Payload{"sitemap_id": query.ID, "sitemap_url": "https://example.invalid/sitemap.xml"}, nil
}

func (d demoWebmaster) GetUserSitemaps(ctx context.Context, query SampleQuery) (Payload, error) {
	return d.GetSitemaps(ctx, query)
}

func (d demoWebmaster) GetUserSitemap(ctx context.Context, query SampleQuery) (Payload, error) {
	return d.GetSitemap(ctx, query)
}

func (d demoWebmaster) GetImportantURLs(ctx context.Context, query SampleQuery) (Payload, error) {
	if err := demoAccess(ctx, query.HostID); err != nil {
		return nil, err
	}
	return Payload{"count": 1, "urls": []any{map[string]any{"url": "https://example.invalid/", "indexing_status": "INDEXED"}}}, nil
}

func (d demoWebmaster) GetImportantURLHistory(ctx context.Context, query SampleQuery) (Payload, error) {
	if err := demoAccess(ctx, query.HostID); err != nil {
		return nil, err
	}
	return Payload{"url": query.URL, "points": []any{}}, nil
}

func (d demoWebmaster) GetRecrawlQuota(ctx context.Context, hostID string) (Payload, error) {
	if err := demoAccess(ctx, hostID); err != nil {
		return nil, err
	}
	return Payload{"daily_quota": 10, "quota_remainder": 10}, nil
}

func (d demoWebmaster) GetRecrawlQueue(ctx context.Context, query SampleQuery) (Payload, error) {
	if err := demoAccess(ctx, query.HostID); err != nil {
		return nil, err
	}
	return Payload{"count": 0, "tasks": []any{}}, nil
}

func (d demoWebmaster) GetRecrawlTask(ctx context.Context, query SampleQuery) (Payload, error) {
	if err := demoAccess(ctx, query.HostID); err != nil {
		return nil, err
	}
	return Payload{"task_id": query.ID, "url": "https://example.invalid/", "state": "DONE"}, nil
}

func (d demoWebmaster) SubmitRecrawl(ctx context.Context, query RecrawlRequest) (Payload, error) {
	if err := demoAccess(ctx, query.HostID); err != nil {
		return nil, err
	}
	return Payload{"task_id": "demo-task", "url": query.URL, "state": "PENDING"}, nil
}

func (d demoWebmaster) GetFeeds(ctx context.Context, hostID string) (Payload, error) {
	if err := demoAccess(ctx, hostID); err != nil {
		return nil, err
	}
	return Payload{"feeds": []any{}}, nil
}

func (d demoWebmaster) GetFeedStatus(ctx context.Context, query SampleQuery) (Payload, error) {
	if err := demoAccess(ctx, query.HostID); err != nil {
		return nil, err
	}
	return Payload{"task_id": query.ID, "state": "DONE"}, nil
}

func (demoWebmaster) ListRegionIDs() []Region   { return append([]Region(nil), CommonRegions...) }
func (demoWebmaster) ListFeedRegions() []Region { return append([]Region(nil), FeedRegions...) }

func (d demoWebmaster) history(ctx context.Context, hostID, indicator string) (Payload, error) {
	if err := demoAccess(ctx, hostID); err != nil {
		return nil, err
	}
	return Payload{"indicators": map[string]any{indicator: []any{map[string]any{"date": "2026-09-01T00:00:00.000Z", "value": 12}}}}, nil
}

func (d demoWebmaster) samples(ctx context.Context, hostID string) (Payload, error) {
	if err := demoAccess(ctx, hostID); err != nil {
		return nil, err
	}
	return Payload{"samples": []any{map[string]any{"url": "https://example.invalid/guide"}}}, nil
}

func demoAccess(ctx context.Context, hostID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if hostID != demoHostID {
		return &RequestError{Kind: "permission_denied", Status: 403}
	}
	return nil
}
