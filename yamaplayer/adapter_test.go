package yamaplayer_test

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	vrclog "github.com/vrclog/vrclog-go"

	"github.com/vrclog/vrclog-adapters/yamaplayer"
)

var update = flag.Bool("update", false, "update golden fixture files")

func makeRecord(t time.Time, msg string) vrclog.Record {
	return vrclog.Record{
		ID:      "test-record",
		Time:    t,
		Level:   vrclog.LevelLog,
		Message: msg,
		Raw:     msg,
	}
}

var fixedTime = time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)

func TestID(t *testing.T) {
	a := yamaplayer.New()
	if a.ID() != "community.yamaplayer" {
		t.Fatalf("ID() = %q, want %q", a.ID(), "community.yamaplayer")
	}
}

func TestYoutubeResolveURL(t *testing.T) {
	a := yamaplayer.New()
	record := makeRecord(fixedTime, "[YamaStream] Resolve youtube url: https://www.youtube.com/watch?v=TESTVIDEO01")

	emissions, err := a.Decode(record)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if len(emissions) != 1 {
		t.Fatalf("len(emissions) = %d, want 1", len(emissions))
	}

	em := emissions[0]
	if em.Rule != "youtube_resolve_url" {
		t.Errorf("Rule = %q, want %q", em.Rule, "youtube_resolve_url")
	}
	ev, ok := em.Event.(vrclog.ResourceURLObserved)
	if !ok {
		t.Fatalf("Event type = %T, want ResourceURLObserved", em.Event)
	}
	if ev.Resource.URL != "https://www.youtube.com/watch?v=TESTVIDEO01" {
		t.Errorf("Resource.URL = %q, want exact preserved URL", ev.Resource.URL)
	}
	if ev.Resource.Kind != vrclog.ResourceKindVideo {
		t.Errorf("Resource.Kind = %q, want video", ev.Resource.Kind)
	}
	if ev.Resource.Role != vrclog.ResourceRoleSource {
		t.Errorf("Resource.Role = %q, want source", ev.Resource.Role)
	}
	if ev.Target == nil || ev.Target.Component != "yamaplayer" {
		t.Errorf("Target.Component = %+v, want yamaplayer", ev.Target)
	}
}

func TestYoutubeResolveURLWithQueryAndFragment(t *testing.T) {
	a := yamaplayer.New()
	url := "https://www.youtube.com/watch?v=TESTVIDEO01&t=30s#frag"
	record := makeRecord(fixedTime, "[YamaStream] Resolve youtube url: "+url)

	emissions, err := a.Decode(record)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if len(emissions) != 1 {
		t.Fatalf("len(emissions) = %d, want 1", len(emissions))
	}
	ev := emissions[0].Event.(vrclog.ResourceURLObserved)
	if ev.Resource.URL != url {
		t.Errorf("Resource.URL = %q, want %q (exact preservation)", ev.Resource.URL, url)
	}
}

func TestVideoError(t *testing.T) {
	a := yamaplayer.New()
	record := makeRecord(fixedTime, "[YamaStream] [1] Video error: 500.")

	emissions, err := a.Decode(record)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if len(emissions) != 1 {
		t.Fatalf("len(emissions) = %d, want 1", len(emissions))
	}

	em := emissions[0]
	if em.Rule != "video_error" {
		t.Errorf("Rule = %q, want %q", em.Rule, "video_error")
	}
	ev, ok := em.Event.(vrclog.MediaErrorObserved)
	if !ok {
		t.Fatalf("Event type = %T, want MediaErrorObserved", em.Event)
	}
	if ev.Stage != vrclog.MediaStagePlayback {
		t.Errorf("Stage = %q, want playback", ev.Stage)
	}
	if ev.Code != "500" {
		t.Errorf("Code = %q, want %q", ev.Code, "500")
	}
	if ev.Message != "500." {
		t.Errorf("Message = %q, want %q", ev.Message, "500.")
	}
	if ev.Target == nil {
		t.Fatalf("Target is nil")
	}
	if ev.Target.Component != "yamaplayer" {
		t.Errorf("Target.Component = %q, want yamaplayer", ev.Target.Component)
	}
	if ev.Target.Key != "1" {
		t.Errorf("Target.Key = %q, want %q", ev.Target.Key, "1")
	}
	if ev.Target.Backend != vrclog.MediaBackendUnknown {
		t.Errorf("Target.Backend = %q, want unknown", ev.Target.Backend)
	}
}

func TestVideoErrorDifferentIndex(t *testing.T) {
	for _, key := range []string{"0", "2", "player1"} {
		t.Run(key, func(t *testing.T) {
			a := yamaplayer.New()
			record := makeRecord(fixedTime, "[YamaStream] ["+key+"] Video error: Connection timeout")

			emissions, err := a.Decode(record)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if len(emissions) != 1 {
				t.Fatalf("len(emissions) = %d, want 1", len(emissions))
			}
			ev := emissions[0].Event.(vrclog.MediaErrorObserved)
			if ev.Target.Key != key {
				t.Errorf("Target.Key = %q, want %q", ev.Target.Key, key)
			}
		})
	}
}

func TestRepeatedDecodeIsDeterministic(t *testing.T) {
	a := yamaplayer.New()
	record := makeRecord(fixedTime, "[YamaStream] Resolve youtube url: https://www.youtube.com/watch?v=TESTVIDEO01")

	first, err1 := a.Decode(record)
	second, err2 := a.Decode(record)
	if err1 != nil || err2 != nil {
		t.Fatalf("Decode() errors = %v, %v", err1, err2)
	}

	b1, _ := json.Marshal(first)
	b2, _ := json.Marshal(second)
	if string(b1) != string(b2) {
		t.Errorf("repeated Decode produced different results:\n%s\n%s", b1, b2)
	}
}

func TestNoMatch(t *testing.T) {
	cases := map[string]string{
		"different_component_prefix":      "[iwaSync3] Resolve youtube url: https://www.youtube.com/watch?v=TESTVIDEO01",
		"prefix_without_rule_anchor":      "[YamaStream] Play track: something",
		"bare_url":                        "https://www.youtube.com/watch?v=TESTVIDEO01",
		"empty_message":                   "",
		"quoted_prefix_other_component":   "[Behaviour] Udon Debug.Log: saw string \"[YamaStream]\" in chat",
		"embedded_anchor_other_component": "[Behaviour] Udon Debug.Log: [YamaStream] Resolve youtube url: https://evil.example/malicious",
		"embedded_error_other_component":  "[Behaviour] Udon Debug.Log: [YamaStream] [1] Video error: 500.",
		"ftp_url_after_anchor":            "[YamaStream] Resolve youtube url: ftp://example.invalid/file",
		"userinfo_url_after_anchor":       "[YamaStream] Resolve youtube url: http://user:pass@example.invalid/file",
		"not_a_url_after_anchor":          "[YamaStream] Resolve youtube url: not-a-url",
		"scheme_only_after_anchor":        "[YamaStream] Resolve youtube url: https://",
	}

	for name, msg := range cases {
		t.Run(name, func(t *testing.T) {
			a := yamaplayer.New()
			record := makeRecord(fixedTime, msg)
			emissions, err := a.Decode(record)
			if err != nil {
				t.Fatalf("Decode() error = %v, want nil", err)
			}
			if emissions != nil {
				t.Fatalf("Decode() emissions = %+v, want nil", emissions)
			}
		})
	}
}

func TestZeroTimeRecord(t *testing.T) {
	a := yamaplayer.New()
	record := makeRecord(time.Time{}, "[YamaStream] Resolve youtube url: https://www.youtube.com/watch?v=TESTVIDEO01")

	emissions, err := a.Decode(record)
	if err != nil {
		t.Fatalf("Decode() error = %v, want nil", err)
	}
	if emissions != nil {
		t.Fatalf("Decode() emissions = %+v, want nil for zero-time record", emissions)
	}
}

func TestMissingURLIsError(t *testing.T) {
	cases := map[string]string{
		"empty_after_anchor":     "[YamaStream] Resolve youtube url: ",
		"whitespace_only_anchor": "[YamaStream] Resolve youtube url:   ",
	}
	for name, msg := range cases {
		t.Run(name, func(t *testing.T) {
			a := yamaplayer.New()
			record := makeRecord(fixedTime, msg)
			emissions, err := a.Decode(record)
			if err == nil {
				t.Fatalf("Decode() error = nil, want error")
			}
			if emissions != nil {
				t.Fatalf("Decode() emissions = %+v, want nil alongside error", emissions)
			}
		})
	}
}

func TestMalformedVideoErrorIsError(t *testing.T) {
	cases := map[string]string{
		"empty_key":    "[YamaStream] [] Video error: 500",
		"space_in_key": "[YamaStream] [1 2] Video error: 500",
	}
	for name, msg := range cases {
		t.Run(name, func(t *testing.T) {
			a := yamaplayer.New()
			record := makeRecord(fixedTime, msg)

			emissions, err := a.Decode(record)
			if err == nil {
				t.Fatalf("Decode() error = nil, want error")
			}
			if emissions != nil {
				t.Fatalf("Decode() emissions = %+v, want nil alongside error", emissions)
			}
		})
	}
}

func TestErrorMessageDoesNotLeakURL(t *testing.T) {
	a := yamaplayer.New()
	record := makeRecord(fixedTime, "[YamaStream] Resolve youtube url: ")
	_, err := a.Decode(record)
	if err == nil {
		t.Fatalf("Decode() error = nil, want error")
	}
	msg := err.Error()
	if len(msg) == 0 {
		t.Fatalf("error message is empty")
	}
	for _, forbidden := range []string{"http://", "https://"} {
		if strings.Contains(msg, forbidden) {
			t.Errorf("error message %q leaks URL scheme %q", msg, forbidden)
		}
	}
}

func TestConcurrentDecode(t *testing.T) {
	a := yamaplayer.New()
	record := makeRecord(fixedTime, "[YamaStream] Resolve youtube url: https://www.youtube.com/watch?v=TESTVIDEO01")

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := a.Decode(record); err != nil {
				t.Errorf("Decode() error = %v", err)
			}
		}()
	}
	wg.Wait()
}

// TestFixtureGolden verifies the adapter's output against golden files
// generated from real VRChat log fixtures using vrclog.ReadFile +
// vrclog.EncodeObservationJSON. Run with -update to regenerate.
func TestFixtureGolden(t *testing.T) {
	scenarios := []string{"youtube_source_url", "video_error"}

	for _, scenario := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			dir := filepath.Join("testdata", scenario)
			inputPath := filepath.Join(dir, "input.log")
			expectedPath := filepath.Join(dir, "expected.json")

			engine, err := vrclog.NewEngine(yamaplayer.New())
			if err != nil {
				t.Fatalf("NewEngine() error = %v", err)
			}

			var observations []vrclog.Observation
			for record, err := range vrclog.ReadFile(context.Background(), vrclog.ReadFileConfig{Path: inputPath}) {
				if err != nil {
					t.Fatalf("ReadFile() error = %v", err)
				}
				result := engine.Process(record)
				if len(result.Diagnostics) != 0 {
					t.Fatalf("unexpected diagnostics: %+v", result.Diagnostics)
				}
				observations = append(observations, result.Observations...)
			}

			var encoded []json.RawMessage
			for _, obs := range observations {
				b, err := vrclog.EncodeObservationJSON(obs)
				if err != nil {
					t.Fatalf("EncodeObservationJSON() error = %v", err)
				}
				encoded = append(encoded, b)
			}
			got, err := json.MarshalIndent(encoded, "", "  ")
			if err != nil {
				t.Fatalf("MarshalIndent() error = %v", err)
			}
			got = append(got, '\n')

			if *update {
				if err := os.WriteFile(expectedPath, got, 0o644); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
				return
			}

			want, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("ReadFile(expected) error = %v", err)
			}
			if string(got) != string(want) {
				t.Errorf("golden mismatch for %s\ngot:\n%s\nwant:\n%s", scenario, got, want)
			}
		})
	}
}
