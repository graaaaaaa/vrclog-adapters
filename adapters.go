// Package adapters provides compile-time community adapters that decode
// VRChat community-project log lines into vrclog-go canonical Events.
package adapters

import (
	vrclog "github.com/vrclog/vrclog-go"

	"github.com/vrclog/vrclog-adapters/iwasync3"
	"github.com/vrclog/vrclog-adapters/yamaplayer"
)

// All returns a fresh, deterministically ordered slice of every community
// Adapter provided by this module. Callers may freely mutate the returned
// slice; it is not shared with subsequent calls.
func All() []vrclog.Adapter {
	return []vrclog.Adapter{
		yamaplayer.New(),
		iwasync3.New(),
	}
}
