# vrclog-adapters

`vrclog-adapters` は、YamaPlayer、iwaSync3等、VRChat上で利用されるコミュニティ製ワールドアセット固有のログ形式を解釈し、[`vrclog-go`](https://github.com/vrclog/vrclog-go) のcanonical Eventを返すAdapterを提供するcompile-timeライブラリである。

```text
VRChat output_log
        │
        ▼
vrclog-go: Record
        │
        ├──────────────────────────────────────┐
        │                                      │
        ▼                                      ▼
vrclog-go built-in Adapter            vrclog-adapters
VRChat client log                      community project log
        │                                      │
        └──────────────────┬───────────────────┘
                           ▼
                 canonical Observation
```

## 責務分離

このリポジトリは、VRChatクライアント自体のログ（built-in Adapter）ではなく、コミュニティ製ワールドアセットが独自に出力するログ形式だけを扱う。ログのread/follow、cursor管理、Adapter実行順序、Observation ID生成は `vrclog-go` の責務であり、本リポジトリでは扱わない。

## 使用例

```go
import (
    vrclog "github.com/vrclog/vrclog-go"
    adapters "github.com/vrclog/vrclog-adapters"
)

allAdapters := []vrclog.Adapter{
    vrclog.NewVRChatAdapter(),
}
allAdapters = append(allAdapters, adapters.All()...)

engine, err := vrclog.NewEngine(allAdapters...)
```

個別Adapterだけを利用する場合:

```go
import "github.com/vrclog/vrclog-adapters/yamaplayer"
import "github.com/vrclog/vrclog-adapters/iwasync3"

engine, err := vrclog.NewEngine(
    vrclog.NewVRChatAdapter(),
    yamaplayer.New(),
    iwasync3.New(),
)
```

## Compatibility matrix

| Project | Adapter | Verified observation | Fixture provenance | Status |
|---|---|---|---|---|
| YamaPlayer | `community.yamaplayer` | original YouTube source URL, player-specific video error | version unknown, captured 2026-08-18 | verified |
| iwaSync3 | `community.iwasync3` | PlayerError; source URL via `vrchat.core` | version unknown, captured 2026-08-18 | verified |
| VizVid | none (core-only) | not evaluated | no fixture | unverified |

See [`compatibility/README.md`](compatibility/README.md) for details.

## Non-goals

- ログファイルのread/follow/rotation
- timestamp/header decode
- cursor管理
- Observation ID生成
- Adapterの実行順序管理
- runtime Adapter enable/disable、remote catalog、auto-update、plugin機構

## Privacy / redaction

Adapterが観測するURLには、unlisted content、private relay、signed token等が含まれ得る。Adapterはこれらを外部送信せず、error messageにURLを再出力しない。テストfixtureは実ログ由来だが、機密情報はredactionしたうえでcommitする。
