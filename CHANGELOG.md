# Changelog

## Unreleased

### Added

- `All()` root composition returning a deterministic, fresh slice of community adapters (`yamaplayer` → `iwasync3`).
- `yamaplayer` package: `community.yamaplayer` adapter.
  - `youtube_resolve_url` rule — observes the original YouTube source URL from `[YamaStream] Resolve youtube url: ...`.
  - `video_error` rule — observes player-specific video errors from `[YamaStream] [<key>] Video error: ...`.
- `iwasync3` package: `community.iwasync3` adapter.
  - `player_error` rule — observes PlayerError messages from `[iwaSync3] ...PlayerError...`.
- Fixture-based test infrastructure (`testdata/<scenario>/{input.log,expected.json,metadata.json}`) with golden-file generation via `vrclog.EncodeObservationJSON`.
- Negative corpus and cross-adapter collision tests.
- `vrclog.NewVRChatAdapter()` + community adapter Engine integration tests.
