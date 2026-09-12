package webmaster

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var datePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

var deviceTypes = map[string]bool{
	"": true, "ALL": true, "DESKTOP": true, "MOBILE": true, "TABLET": true, "MOBILE_AND_TABLET": true,
}

// RequireHostID rejects blank host identifiers before they become path segments.
func RequireHostID(hostID string) (string, error) {
	hostID = strings.TrimSpace(hostID)
	if hostID == "" {
		return "", fmt.Errorf("host_id is required")
	}
	return hostID, nil
}

// ValidateDateRange checks host_id and optional YYYY-MM-DD bounds.
func ValidateDateRange(query DateRange) (DateRange, error) {
	var err error
	query.HostID, err = RequireHostID(query.HostID)
	if err != nil {
		return DateRange{}, err
	}
	query.DateFrom, err = optionalDate("date_from", query.DateFrom)
	if err != nil {
		return DateRange{}, err
	}
	query.DateTo, err = optionalDate("date_to", query.DateTo)
	if err != nil {
		return DateRange{}, err
	}
	if query.DateFrom != "" && query.DateTo != "" && query.DateFrom > query.DateTo {
		return DateRange{}, fmt.Errorf("date_from must be on or before date_to")
	}
	return query, nil
}

// ValidatePopularQuery fills defaults used by the popular-queries endpoint.
func ValidatePopularQuery(query PopularQuery) (PopularQuery, error) {
	rangeQuery, err := ValidateDateRange(DateRange{HostID: query.HostID, DateFrom: query.DateFrom, DateTo: query.DateTo})
	if err != nil {
		return PopularQuery{}, err
	}
	query.HostID, query.DateFrom, query.DateTo = rangeQuery.HostID, rangeQuery.DateFrom, rangeQuery.DateTo
	query.OrderBy = strings.TrimSpace(query.OrderBy)
	if query.OrderBy == "" {
		query.OrderBy = "TOTAL_SHOWS"
	}
	if query.OrderBy != "TOTAL_SHOWS" && query.OrderBy != "TOTAL_CLICKS" {
		return PopularQuery{}, fmt.Errorf("order_by must be TOTAL_SHOWS or TOTAL_CLICKS")
	}
	query.DeviceType = strings.ToUpper(strings.TrimSpace(query.DeviceType))
	if !deviceTypes[query.DeviceType] {
		return PopularQuery{}, fmt.Errorf("invalid device_type")
	}
	query.Limit, query.Offset, err = boundedPage(query.Limit, query.Offset, 100, 500)
	if err != nil {
		return PopularQuery{}, err
	}
	return query, nil
}

// ValidateQueryHistory checks host, optional query id, device and dates.
func ValidateQueryHistory(query QueryHistory, requireQueryID bool) (QueryHistory, error) {
	rangeQuery, err := ValidateDateRange(DateRange{HostID: query.HostID, DateFrom: query.DateFrom, DateTo: query.DateTo})
	if err != nil {
		return QueryHistory{}, err
	}
	query.HostID, query.DateFrom, query.DateTo = rangeQuery.HostID, rangeQuery.DateFrom, rangeQuery.DateTo
	query.QueryID = strings.TrimSpace(query.QueryID)
	if requireQueryID && query.QueryID == "" {
		return QueryHistory{}, fmt.Errorf("query_id is required")
	}
	query.DeviceType = strings.ToUpper(strings.TrimSpace(query.DeviceType))
	if !deviceTypes[query.DeviceType] {
		return QueryHistory{}, fmt.Errorf("invalid device_type")
	}
	return query, nil
}

// ValidateQueryAnalytics fills defaults for the query↔URL report.
func ValidateQueryAnalytics(query QueryAnalytics) (QueryAnalytics, error) {
	hostID, err := RequireHostID(query.HostID)
	if err != nil {
		return QueryAnalytics{}, err
	}
	query.HostID = hostID
	query.TextIndicator = strings.ToUpper(strings.TrimSpace(query.TextIndicator))
	if query.TextIndicator == "" {
		query.TextIndicator = "URL"
	}
	if query.TextIndicator != "URL" && query.TextIndicator != "QUERY" {
		return QueryAnalytics{}, fmt.Errorf("text_indicator must be URL or QUERY")
	}
	query.DeviceTypeIndicator = strings.ToUpper(strings.TrimSpace(query.DeviceTypeIndicator))
	if query.DeviceTypeIndicator == "" {
		query.DeviceTypeIndicator = "ALL"
	}
	if !deviceTypes[query.DeviceTypeIndicator] {
		return QueryAnalytics{}, fmt.Errorf("invalid device_type_indicator")
	}
	query.Limit, query.Offset, err = boundedPage(query.Limit, query.Offset, 20, 500)
	if err != nil {
		return QueryAnalytics{}, err
	}
	return query, nil
}

// ValidateSampleQuery checks host, optional resource id/url and paging.
func ValidateSampleQuery(query SampleQuery, requireID, requireURL bool, defaultLimit, maxLimit int) (SampleQuery, error) {
	rangeQuery, err := ValidateDateRange(DateRange{HostID: query.HostID, DateFrom: query.DateFrom, DateTo: query.DateTo})
	if err != nil {
		return SampleQuery{}, err
	}
	query.HostID, query.DateFrom, query.DateTo = rangeQuery.HostID, rangeQuery.DateFrom, rangeQuery.DateTo
	query.ID = strings.TrimSpace(query.ID)
	query.URL = strings.TrimSpace(query.URL)
	if requireID && query.ID == "" {
		return SampleQuery{}, fmt.Errorf("id is required")
	}
	if requireURL && query.URL == "" {
		return SampleQuery{}, fmt.Errorf("url is required")
	}
	if requireURL {
		parsed, err := url.Parse(query.URL)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return SampleQuery{}, fmt.Errorf("url must be an absolute HTTP(S) URL")
		}
	}
	query.Limit, query.Offset, err = boundedPage(query.Limit, query.Offset, defaultLimit, maxLimit)
	if err != nil {
		return SampleQuery{}, err
	}
	return query, nil
}

// ValidateRecrawlRequest checks host and the page queued for reindexing.
func ValidateRecrawlRequest(query RecrawlRequest) (RecrawlRequest, error) {
	hostID, err := RequireHostID(query.HostID)
	if err != nil {
		return RecrawlRequest{}, err
	}
	query.HostID = hostID
	query.URL = strings.TrimSpace(query.URL)
	parsed, err := url.Parse(query.URL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return RecrawlRequest{}, fmt.Errorf("url must be an absolute HTTP(S) URL")
	}
	return query, nil
}

func optionalDate(field, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if !datePattern.MatchString(value) {
		return "", fmt.Errorf("%s must be YYYY-MM-DD", field)
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return "", fmt.Errorf("%s is not a valid calendar date", field)
	}
	return value, nil
}

func boundedPage(limit, offset, fallback, maxLimit int) (int, int, error) {
	if limit < 0 || limit > maxLimit {
		return 0, 0, fmt.Errorf("limit must be between 0 and %d", maxLimit)
	}
	if limit == 0 {
		limit = fallback
	}
	if offset < 0 {
		return 0, 0, fmt.Errorf("offset must be >= 0")
	}
	return limit, offset, nil
}

// RFC3339Date converts YYYY-MM-DD to the UTC instant Yandex query params expect.
func RFC3339Date(value string) string {
	if value == "" {
		return ""
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return value
	}
	return parsed.UTC().Format(time.RFC3339Nano)
}
