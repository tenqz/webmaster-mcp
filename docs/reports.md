# Report ingestion contract

The MCP implementation version identifies the adapter contract. Keep that version with every stored response. The adapter preserves upstream field names and extra fields. Report payloads are source evidence, not a normalized cross-search-engine metric model.

| Tool | Scope | Consumer rules |
| --- | --- | --- |
| get_summary | Current host state | `sqi`, `searchable_pages_count`, `excluded_pages_count` are independent values. Missing is unknown, not zero. |
| get_query_history | Host, date interval, device | Preserve `indicators` with each source date and value. TOTAL_CLICKS and TOTAL_SHOWS are counts; position indicators are averages, never sums. |
| get_popular_queries | Ranked sample | Persist `queries`, `count`, requested offset/limit. Never sum this sample to obtain site totals. |
| get_diagnostics | Current host diagnostics | Preserve severity and source details as evidence. Missing fields are not proof of no problems. |

Dates passed as YYYY-MM-DD become UTC RFC3339 timestamps on API requests. Preserve returned timestamps exactly; do not silently reinterpret them as GSC Pacific dates. Save requested boundaries separately from returned coverage. These reports can revise previous dates: refresh overlapping periods and retain the original fetched-at time of every batch.

For offset/limit reports, use bounded pagination until a short page or the provider's documented end. A page cap, error, or unsupported field means incomplete coverage; do not replace a previously complete dataset with it. History and summary tools do not expose offset pagination. Samples are not exhaustive exports even after every available page is fetched.

The SEO Agent collector uses get_summary, get_diagnostics and get_query_history independently. It records the host ID, tool, exact arguments, collection time, adapter version, status and raw structured JSON in PostgreSQL. One unavailable report must not erase another successful report. Never schedule submit_recrawl.

Fixture tests are synthetic contract examples, not recordings proving live account access. Run the built-in read-only smoke against your deployment after configuring OAuth. Do not commit real tokens or private account reports as fixtures.
