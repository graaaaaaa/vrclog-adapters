// Package yamaplayer decodes YamaPlayer log lines from VRChat output_log
// into vrclog-go canonical Events.
package yamaplayer

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	vrclog "github.com/vrclog/vrclog-go"
)

const (
	prefix        = "[YamaStream]"
	resolveAnchor = "Resolve youtube url: "
	errorAnchor   = "Video error:"
)

var reVideoError = regexp.MustCompile(`^\[YamaStream\] \[(\S+)\] Video error: (.+)$`)

type adapter struct{}

// New returns a stateless Adapter for the YamaPlayer community project.
func New() vrclog.Adapter { return adapter{} }

func (a adapter) ID() vrclog.AdapterID { return "community.yamaplayer" }

func (a adapter) Decode(record vrclog.Record) ([]vrclog.Emission, error) {
	if record.Time.IsZero() {
		return nil, nil
	}

	msg := record.Message
	if !strings.HasPrefix(msg, prefix) {
		return nil, nil
	}

	if strings.Contains(msg, resolveAnchor) {
		return decodeYoutubeResolveURL(msg)
	}

	if strings.Contains(msg, errorAnchor) {
		return decodeVideoError(msg)
	}

	return nil, nil
}

func decodeYoutubeResolveURL(msg string) ([]vrclog.Emission, error) {
	idx := strings.Index(msg, resolveAnchor)
	rest := strings.TrimRight(msg[idx+len(resolveAnchor):], " \t")

	if rest == "" {
		return nil, fmt.Errorf("yamaplayer: youtube_resolve_url: missing URL after anchor")
	}

	if !isHTTPURL(rest) {
		return nil, nil
	}

	return []vrclog.Emission{{
		Rule: "youtube_resolve_url",
		Event: vrclog.ResourceURLObserved{
			Resource: vrclog.RemoteResource{
				URL:  rest,
				Kind: vrclog.ResourceKindVideo,
				Role: vrclog.ResourceRoleSource,
			},
			Target: &vrclog.MediaTarget{
				Component: "yamaplayer",
			},
		},
	}}, nil
}

func decodeVideoError(msg string) ([]vrclog.Emission, error) {
	m := reVideoError.FindStringSubmatch(msg)
	if m == nil {
		return nil, fmt.Errorf("yamaplayer: video_error: malformed error line")
	}

	key := m[1]
	message := m[2]
	code := strings.TrimRight(message, ".")

	return []vrclog.Emission{{
		Rule: "video_error",
		Event: vrclog.MediaErrorObserved{
			Stage:   vrclog.MediaStagePlayback,
			Code:    code,
			Message: message,
			Target: &vrclog.MediaTarget{
				Component: "yamaplayer",
				Key:       key,
				Backend:   vrclog.MediaBackendUnknown,
			},
		},
	}}, nil
}

func isHTTPURL(rawURL string) bool {
	if strings.ContainsFunc(rawURL, func(r rune) bool {
		return unicode.IsControl(r) || unicode.IsSpace(r)
	}) {
		return false
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if u.User != nil {
		return false
	}
	return u.Host != ""
}
