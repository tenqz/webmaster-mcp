package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tenqz/yandex-webmaster-mcp/internal/webmaster"
)

// Toolset holds dependencies for MCP tool handlers.
type Toolset struct {
	API webmaster.Webmaster
}

func (t *Toolset) GetUser(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	user, err := t.API.GetUser(ctx)
	if err != nil {
		return toolFailure(err)
	}
	return jsonResult(UserOutput{UserID: user.UserID})
}

func (t *Toolset) ListHosts(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	hosts, err := t.API.ListHosts(ctx)
	if err != nil {
		return toolFailure(err)
	}
	return jsonResult(HostsOutput{Hosts: hosts, Count: len(hosts)})
}

func (t *Toolset) GetHost(ctx context.Context, _ *mcp.CallToolRequest, in HostInput) (*mcp.CallToolResult, any, error) {
	hostID, err := webmaster.RequireHostID(in.HostID)
	if err != nil {
		return toolError("%v", err)
	}
	payload, err := t.API.GetHost(ctx, hostID)
	if err != nil {
		return toolFailure(err)
	}
	return jsonResult(payload)
}

func (t *Toolset) GetSummary(ctx context.Context, _ *mcp.CallToolRequest, in HostInput) (*mcp.CallToolResult, any, error) {
	hostID, err := webmaster.RequireHostID(in.HostID)
	if err != nil {
		return toolError("%v", err)
	}
	payload, err := t.API.GetSummary(ctx, hostID)
	if err != nil {
		return toolFailure(err)
	}
	return jsonResult(payload)
}

func (t *Toolset) GetSQIHistory(ctx context.Context, _ *mcp.CallToolRequest, in DateRangeInput) (*mcp.CallToolResult, any, error) {
	query, err := webmaster.ValidateDateRange(webmaster.DateRange{HostID: in.HostID, DateFrom: in.DateFrom, DateTo: in.DateTo})
	if err != nil {
		return toolError("%v", err)
	}
	payload, err := t.API.GetSQIHistory(ctx, query)
	if err != nil {
		return toolFailure(err)
	}
	return jsonResult(payload)
}

func (t *Toolset) GetDiagnostics(ctx context.Context, _ *mcp.CallToolRequest, in HostInput) (*mcp.CallToolResult, any, error) {
	hostID, err := webmaster.RequireHostID(in.HostID)
	if err != nil {
		return toolError("%v", err)
	}
	payload, err := t.API.GetDiagnostics(ctx, hostID)
	if err != nil {
		return toolFailure(err)
	}
	return jsonResult(payload)
}

func (t *Toolset) GetPopularQueries(ctx context.Context, _ *mcp.CallToolRequest, in PopularQueriesInput) (*mcp.CallToolResult, any, error) {
	query, err := webmaster.ValidatePopularQuery(webmaster.PopularQuery{
		HostID: in.HostID, OrderBy: in.OrderBy, DeviceType: in.DeviceType, DateFrom: in.DateFrom, DateTo: in.DateTo, Limit: in.Limit, Offset: in.Offset,
	})
	if err != nil {
		return toolError("%v", err)
	}
	payload, err := t.API.GetPopularQueries(ctx, query)
	if err != nil {
		return toolFailure(err)
	}
	return jsonResult(payload)
}

func (t *Toolset) GetQueryHistory(ctx context.Context, _ *mcp.CallToolRequest, in QueryHistoryInput) (*mcp.CallToolResult, any, error) {
	query, err := webmaster.ValidateQueryHistory(webmaster.QueryHistory{HostID: in.HostID, DeviceType: in.DeviceType, DateFrom: in.DateFrom, DateTo: in.DateTo}, false)
	if err != nil {
		return toolError("%v", err)
	}
	payload, err := t.API.GetQueryHistory(ctx, query)
	if err != nil {
		return toolFailure(err)
	}
	return jsonResult(payload)
}

func (t *Toolset) GetSingleQueryHistory(ctx context.Context, _ *mcp.CallToolRequest, in QueryHistoryInput) (*mcp.CallToolResult, any, error) {
	query, err := webmaster.ValidateQueryHistory(webmaster.QueryHistory{HostID: in.HostID, QueryID: in.QueryID, DeviceType: in.DeviceType, DateFrom: in.DateFrom, DateTo: in.DateTo}, true)
	if err != nil {
		return toolError("%v", err)
	}
	payload, err := t.API.GetSingleQueryHistory(ctx, query)
	if err != nil {
		return toolFailure(err)
	}
	return jsonResult(payload)
}

func (t *Toolset) GetQueryAnalytics(ctx context.Context, _ *mcp.CallToolRequest, in QueryAnalyticsInput) (*mcp.CallToolResult, any, error) {
	query, err := webmaster.ValidateQueryAnalytics(webmaster.QueryAnalytics{
		HostID: in.HostID, TextIndicator: in.TextIndicator, DeviceTypeIndicator: in.DeviceTypeIndicator, RegionIDs: in.RegionIDs, Limit: in.Limit, Offset: in.Offset, Filters: in.Filters, SortByDate: in.SortByDate,
	})
	if err != nil {
		return toolError("%v", err)
	}
	payload, err := t.API.GetQueryAnalytics(ctx, query)
	if err != nil {
		return toolFailure(err)
	}
	return jsonResult(payload)
}

func (t *Toolset) dateRange(ctx context.Context, in DateRangeInput, fn func(context.Context, webmaster.DateRange) (webmaster.Payload, error)) (*mcp.CallToolResult, any, error) {
	query, err := webmaster.ValidateDateRange(webmaster.DateRange{HostID: in.HostID, DateFrom: in.DateFrom, DateTo: in.DateTo})
	if err != nil {
		return toolError("%v", err)
	}
	payload, err := fn(ctx, query)
	if err != nil {
		return toolFailure(err)
	}
	return jsonResult(payload)
}

func (t *Toolset) GetIndexingHistory(ctx context.Context, _ *mcp.CallToolRequest, in DateRangeInput) (*mcp.CallToolResult, any, error) {
	return t.dateRange(ctx, in, t.API.GetIndexingHistory)
}

func (t *Toolset) GetIndexingSamples(ctx context.Context, _ *mcp.CallToolRequest, in DateRangeInput) (*mcp.CallToolResult, any, error) {
	return t.dateRange(ctx, in, t.API.GetIndexingSamples)
}

func (t *Toolset) GetInSearchHistory(ctx context.Context, _ *mcp.CallToolRequest, in DateRangeInput) (*mcp.CallToolResult, any, error) {
	return t.dateRange(ctx, in, t.API.GetInSearchHistory)
}

func (t *Toolset) GetInSearchSamples(ctx context.Context, _ *mcp.CallToolRequest, in DateRangeInput) (*mcp.CallToolResult, any, error) {
	return t.dateRange(ctx, in, t.API.GetInSearchSamples)
}

func (t *Toolset) GetSearchEventsHistory(ctx context.Context, _ *mcp.CallToolRequest, in DateRangeInput) (*mcp.CallToolResult, any, error) {
	return t.dateRange(ctx, in, t.API.GetSearchEventsHistory)
}

func (t *Toolset) GetSearchEventsSamples(ctx context.Context, _ *mcp.CallToolRequest, in DateRangeInput) (*mcp.CallToolResult, any, error) {
	return t.dateRange(ctx, in, t.API.GetSearchEventsSamples)
}

func (t *Toolset) sample(ctx context.Context, in SampleInput, requireID, requireURL bool, defaultLimit, maxLimit int, fn func(context.Context, webmaster.SampleQuery) (webmaster.Payload, error)) (*mcp.CallToolResult, any, error) {
	query, err := webmaster.ValidateSampleQuery(webmaster.SampleQuery{HostID: in.HostID, ID: in.ID, URL: in.URL, Limit: in.Limit, Offset: in.Offset, DateFrom: in.DateFrom, DateTo: in.DateTo}, requireID, requireURL, defaultLimit, maxLimit)
	if err != nil {
		return toolError("%v", err)
	}
	payload, err := fn(ctx, query)
	if err != nil {
		return toolFailure(err)
	}
	return jsonResult(payload)
}

func (t *Toolset) GetExternalLinks(ctx context.Context, _ *mcp.CallToolRequest, in SampleInput) (*mcp.CallToolResult, any, error) {
	return t.sample(ctx, in, false, false, 10, 100, t.API.GetExternalLinks)
}

func (t *Toolset) GetExternalLinksHistory(ctx context.Context, _ *mcp.CallToolRequest, in DateRangeInput) (*mcp.CallToolResult, any, error) {
	return t.dateRange(ctx, in, t.API.GetExternalLinksHistory)
}

func (t *Toolset) GetBrokenInternalLinks(ctx context.Context, _ *mcp.CallToolRequest, in SampleInput) (*mcp.CallToolResult, any, error) {
	return t.sample(ctx, in, false, false, 10, 100, t.API.GetBrokenInternalLinks)
}

func (t *Toolset) GetBrokenInternalLinksHistory(ctx context.Context, _ *mcp.CallToolRequest, in DateRangeInput) (*mcp.CallToolResult, any, error) {
	return t.dateRange(ctx, in, t.API.GetBrokenInternalLinksHistory)
}

func (t *Toolset) GetSitemaps(ctx context.Context, _ *mcp.CallToolRequest, in SampleInput) (*mcp.CallToolResult, any, error) {
	return t.sample(ctx, in, false, false, 10, 100, t.API.GetSitemaps)
}

func (t *Toolset) GetSitemap(ctx context.Context, _ *mcp.CallToolRequest, in SampleInput) (*mcp.CallToolResult, any, error) {
	return t.sample(ctx, in, true, false, 10, 100, t.API.GetSitemap)
}

func (t *Toolset) GetUserSitemaps(ctx context.Context, _ *mcp.CallToolRequest, in SampleInput) (*mcp.CallToolResult, any, error) {
	return t.sample(ctx, in, false, false, 10, 100, t.API.GetUserSitemaps)
}

func (t *Toolset) GetUserSitemap(ctx context.Context, _ *mcp.CallToolRequest, in SampleInput) (*mcp.CallToolResult, any, error) {
	return t.sample(ctx, in, true, false, 10, 100, t.API.GetUserSitemap)
}

func (t *Toolset) GetImportantURLs(ctx context.Context, _ *mcp.CallToolRequest, in SampleInput) (*mcp.CallToolResult, any, error) {
	return t.sample(ctx, in, false, false, 10, 100, t.API.GetImportantURLs)
}

func (t *Toolset) GetImportantURLHistory(ctx context.Context, _ *mcp.CallToolRequest, in SampleInput) (*mcp.CallToolResult, any, error) {
	return t.sample(ctx, in, false, true, 10, 100, t.API.GetImportantURLHistory)
}

func (t *Toolset) GetRecrawlQuota(ctx context.Context, _ *mcp.CallToolRequest, in HostInput) (*mcp.CallToolResult, any, error) {
	hostID, err := webmaster.RequireHostID(in.HostID)
	if err != nil {
		return toolError("%v", err)
	}
	payload, err := t.API.GetRecrawlQuota(ctx, hostID)
	if err != nil {
		return toolFailure(err)
	}
	return jsonResult(payload)
}

func (t *Toolset) GetRecrawlQueue(ctx context.Context, _ *mcp.CallToolRequest, in SampleInput) (*mcp.CallToolResult, any, error) {
	return t.sample(ctx, in, false, false, 10, 100, t.API.GetRecrawlQueue)
}

func (t *Toolset) GetRecrawlTask(ctx context.Context, _ *mcp.CallToolRequest, in SampleInput) (*mcp.CallToolResult, any, error) {
	return t.sample(ctx, in, true, false, 10, 100, t.API.GetRecrawlTask)
}

func (t *Toolset) SubmitRecrawl(ctx context.Context, _ *mcp.CallToolRequest, in RecrawlInput) (*mcp.CallToolResult, any, error) {
	query, err := webmaster.ValidateRecrawlRequest(webmaster.RecrawlRequest{HostID: in.HostID, URL: in.URL})
	if err != nil {
		return toolError("%v", err)
	}
	payload, err := t.API.SubmitRecrawl(ctx, query)
	if err != nil {
		return toolFailure(err)
	}
	return jsonResult(payload)
}

func (t *Toolset) GetFeeds(ctx context.Context, _ *mcp.CallToolRequest, in HostInput) (*mcp.CallToolResult, any, error) {
	hostID, err := webmaster.RequireHostID(in.HostID)
	if err != nil {
		return toolError("%v", err)
	}
	payload, err := t.API.GetFeeds(ctx, hostID)
	if err != nil {
		return toolFailure(err)
	}
	return jsonResult(payload)
}

func (t *Toolset) GetFeedStatus(ctx context.Context, _ *mcp.CallToolRequest, in SampleInput) (*mcp.CallToolResult, any, error) {
	return t.sample(ctx, in, true, false, 10, 100, t.API.GetFeedStatus)
}

func (t *Toolset) ListRegionIDs(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult, any, error) {
	regions := t.API.ListRegionIDs()
	return jsonResult(RegionsOutput{Regions: regions, Count: len(regions)})
}

func (t *Toolset) ListFeedRegions(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult, any, error) {
	regions := t.API.ListFeedRegions()
	return jsonResult(RegionsOutput{Regions: regions, Count: len(regions)})
}
