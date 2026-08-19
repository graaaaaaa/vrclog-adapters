package iwasync3_test

import (
	"context"
	"path/filepath"
	"testing"

	vrclog "github.com/vrclog/vrclog-go"

	"github.com/vrclog/vrclog-adapters/iwasync3"
)

func TestEngineIntegration_MixedWithCore(t *testing.T) {
	engine, err := vrclog.NewEngine(
		vrclog.NewVRChatAdapter(),
		iwasync3.New(),
	)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	path := filepath.Join("testdata", "mixed_with_core", "input.log")

	var observations []vrclog.Observation
	for record, err := range vrclog.ReadFile(context.Background(), vrclog.ReadFileConfig{Path: path}) {
		if err != nil {
			t.Fatalf("ReadFile() error = %v", err)
		}
		result := engine.Process(record)
		if len(result.Diagnostics) != 0 {
			t.Fatalf("unexpected diagnostics: %+v", result.Diagnostics)
		}
		observations = append(observations, result.Observations...)
	}

	var (
		coreSourceCount int
		iwaErrorCount   int
	)

	for _, obs := range observations {
		switch {
		case obs.AdapterID == "vrchat.core" && obs.RuleID == "video_resolve_attempt":
			coreSourceCount++
			ev, ok := obs.Event.(vrclog.ResourceURLObserved)
			if !ok {
				t.Fatalf("event type = %T, want ResourceURLObserved", obs.Event)
			}
			if ev.Resource.URL != "https://www.youtube.com/watch?v=TESTVIDEO01" {
				t.Errorf("source URL = %q, want original YouTube URL", ev.Resource.URL)
			}

		case obs.AdapterID == "community.iwasync3" && obs.RuleID == "player_error":
			iwaErrorCount++
		}

		if obs.AdapterID == "community.iwasync3" && obs.RuleID != "player_error" {
			t.Errorf("unexpected community.iwasync3 observation for rule %s", obs.RuleID)
		}
	}

	if coreSourceCount != 1 {
		t.Errorf("coreSourceCount = %d, want 1", coreSourceCount)
	}
	if iwaErrorCount != 1 {
		t.Errorf("iwaErrorCount = %d, want 1 (continuation line must not be parsed)", iwaErrorCount)
	}
}
