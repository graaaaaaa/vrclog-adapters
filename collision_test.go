package adapters_test

import (
	"testing"
	"time"

	vrclog "github.com/vrclog/vrclog-go"

	adapters "github.com/vrclog/vrclog-adapters"
)

var fixedTime = time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)

func makeRecord(msg string) vrclog.Record {
	return vrclog.Record{
		ID:      "test-record",
		Time:    fixedTime,
		Level:   vrclog.LevelLog,
		Message: msg,
		Raw:     msg,
	}
}

// TestNoCrossAdapterCollision verifies that a log line owned by one
// community project is not also claimed by another community adapter.
func TestNoCrossAdapterCollision(t *testing.T) {
	cases := map[string]struct {
		msg          string
		wantAdapter  vrclog.AdapterID
		wantEmission bool
	}{
		"yamaplayer_url": {
			msg:          "[YamaStream] Resolve youtube url: https://www.youtube.com/watch?v=TESTVIDEO01",
			wantAdapter:  "community.yamaplayer",
			wantEmission: true,
		},
		"yamaplayer_error": {
			msg:          "[YamaStream] [1] Video error: 500.",
			wantAdapter:  "community.yamaplayer",
			wantEmission: true,
		},
		"iwasync3_error": {
			msg:          "[iwaSync3] There was a `PlayerError` error in the video.",
			wantAdapter:  "community.iwasync3",
			wantEmission: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			record := makeRecord(tc.msg)

			for _, a := range adapters.All() {
				emissions, err := a.Decode(record)
				if err != nil {
					t.Fatalf("%s: Decode() error = %v", a.ID(), err)
				}

				if a.ID() == tc.wantAdapter {
					if tc.wantEmission && len(emissions) == 0 {
						t.Errorf("%s: expected an emission, got none", a.ID())
					}
					continue
				}

				if len(emissions) != 0 {
					t.Errorf("%s: unexpectedly emitted for a line owned by %s: %+v", a.ID(), tc.wantAdapter, emissions)
				}
			}
		})
	}
}

// TestGenericCoreLineNoCommunityEmission verifies that lines belonging to
// vrclog-go's built-in vrchat.core adapter are not also claimed by any
// community adapter.
func TestGenericCoreLineNoCommunityEmission(t *testing.T) {
	cases := []string{
		"[Video Playback] Attempting to resolve URL 'https://relay.example.invalid/TESTVIDEO01'",
		"[Video Playback] URL 'https://relay.example.invalid/TESTVIDEO01' resolved to 'https://cdn.example.invalid/TESTVIDEO01.mp4'",
		"[AVProVideo] Opening https://relay.example.invalid/TESTVIDEO01",
	}

	for _, msg := range cases {
		record := makeRecord(msg)
		for _, a := range adapters.All() {
			emissions, err := a.Decode(record)
			if err != nil {
				t.Fatalf("%s: Decode() error = %v", a.ID(), err)
			}
			if len(emissions) != 0 {
				t.Errorf("%s: unexpectedly emitted for a generic core line %q: %+v", a.ID(), msg, emissions)
			}
		}
	}
}
