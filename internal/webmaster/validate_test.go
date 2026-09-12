package webmaster

import "testing"

// TestValidatePopularQueryRejectsUnknownOrder documents allowed sort keys.
func TestValidatePopularQueryRejectsUnknownOrder(t *testing.T) {
	_, err := ValidatePopularQuery(PopularQuery{HostID: "https:example.com:443", OrderBy: "RANK"})
	if err == nil {
		t.Fatal("expected an error")
	}
}

// TestValidateDateRangeRejectsInvertedBounds documents date_from/date_to ordering.
func TestValidateDateRangeRejectsInvertedBounds(t *testing.T) {
	_, err := ValidateDateRange(DateRange{HostID: "https:example.com:443", DateFrom: "2026-09-10", DateTo: "2026-09-01"})
	if err == nil {
		t.Fatal("expected an error")
	}
}

// TestRequireHostIDRejectsBlank documents that host-scoped tools need a host_id.
func TestRequireHostIDRejectsBlank(t *testing.T) {
	if _, err := RequireHostID("  "); err == nil {
		t.Fatal("expected an error")
	}
}

// TestValidateRecrawlRequestRequiresHTTPURL documents recrawl URL shape.
func TestValidateRecrawlRequestRequiresHTTPURL(t *testing.T) {
	_, err := ValidateRecrawlRequest(RecrawlRequest{HostID: "https:example.com:443", URL: "example.invalid"})
	if err == nil {
		t.Fatal("expected an error")
	}
}

// TestValidatePopularQueryDefaults documents TOTAL_SHOWS and paging fallbacks.
func TestValidatePopularQueryDefaults(t *testing.T) {
	query, err := ValidatePopularQuery(PopularQuery{HostID: "https:example.com:443"})
	if err != nil {
		t.Fatal(err)
	}
	if query.OrderBy != "TOTAL_SHOWS" || query.Limit != 100 {
		t.Fatalf("query=%+v", query)
	}
}
