// Package iwasync3 decodes iwaSync3 log lines from VRChat output_log into
// vrclog-go canonical Events.
package iwasync3

import (
	"strings"

	vrclog "github.com/vrclog/vrclog-go"
)

const (
	prefixWithSpace = "[iwaSync3] "
	playerErrorText = "PlayerError"
)

type adapter struct{}

// New returns a stateless Adapter for the iwaSync3 community project.
func New() vrclog.Adapter { return adapter{} }

func (a adapter) ID() vrclog.AdapterID { return "community.iwasync3" }

func (a adapter) Decode(record vrclog.Record) ([]vrclog.Emission, error) {
	if record.Time.IsZero() {
		return nil, nil
	}

	msg := record.Message
	if !strings.HasPrefix(msg, prefixWithSpace) {
		return nil, nil
	}
	if !strings.Contains(msg, playerErrorText) {
		return nil, nil
	}

	message := strings.TrimRight(msg[len(prefixWithSpace):], " \t")

	return []vrclog.Emission{{
		Rule: "player_error",
		Event: vrclog.MediaErrorObserved{
			Stage:   vrclog.MediaStagePlayback,
			Message: message,
			Target: &vrclog.MediaTarget{
				Component: "iwasync3",
				Backend:   vrclog.MediaBackendUnknown,
			},
		},
	}}, nil
}
