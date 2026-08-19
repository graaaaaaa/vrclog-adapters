# Compatibility matrix

| Project | Adapter | Verified observation | Fixture provenance | Status |
|---|---|---|---|---|
| YamaPlayer | `community.yamaplayer` | original YouTube source URL (`resource.url_observed`); player-specific video error (`media.error_observed`) | version unknown, captured 2026-08-18, source: public_issue | verified |
| iwaSync3 | `community.iwasync3` | PlayerError (`media.error_observed`); source URL is observed via `vrchat.core`, not this adapter | version unknown, captured 2026-08-18, source: public_issue | verified |
| VizVid | none (core-only) | not evaluated — no real log fixture available yet | no fixture | unverified |

## Status vocabulary

- `verified`: at least one real-log-derived fixture exercises the rule end-to-end through `vrclog.NewEngine`.
- `unverified`: no fixture exists yet; no support claim is made.

No other status tiers (e.g. `experimental`, `supported >=x`) are used.
