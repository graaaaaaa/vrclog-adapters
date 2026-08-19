package yamaplayer_test

import (
	"context"
	"path/filepath"
	"testing"

	vrclog "github.com/vrclog/vrclog-go"

	"github.com/vrclog/vrclog-adapters/yamaplayer"
)

func TestEngineIntegration_MixedWithCore(t *testing.T) {
	engine, err := vrclog.NewEngine(
		vrclog.NewVRChatAdapter(),
		yamaplayer.New(),
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
		yamaSourceCount   int
		coreResolveCount  int
		coreResolvedCount int
		yamaErrorCount    int
	)

	for _, obs := range observations {
		switch {
		case obs.AdapterID == "community.yamaplayer" && obs.RuleID == "youtube_resolve_url":
			yamaSourceCount++
			ev, ok := obs.Event.(vrclog.ResourceURLObserved)
			if !ok {
				t.Fatalf("event type = %T, want ResourceURLObserved", obs.Event)
			}
			if ev.Resource.URL != "https://www.youtube.com/watch?v=TESTVIDEO01" {
				t.Errorf("source URL = %q, want original YouTube URL", ev.Resource.URL)
			}
			if ev.Resource.Role != vrclog.ResourceRoleSource {
				t.Errorf("source Role = %q, want source", ev.Resource.Role)
			}

		case obs.AdapterID == "vrchat.core" && obs.RuleID == "video_resolve_attempt":
			coreResolveCount++
			ev, ok := obs.Event.(vrclog.ResourceURLObserved)
			if !ok {
				t.Fatalf("event type = %T, want ResourceURLObserved", obs.Event)
			}
			if ev.Resource.URL != "https://relay.example.invalid/TESTVIDEO01" {
				t.Errorf("relay URL = %q, want relay URL", ev.Resource.URL)
			}

		case obs.AdapterID == "vrchat.core" && obs.RuleID == "video_resolved":
			coreResolvedCount++

		case obs.AdapterID == "community.yamaplayer" && obs.RuleID == "video_error":
			yamaErrorCount++
		}

		if obs.AdapterID == "community.yamaplayer" {
			if _, ok := obs.Event.(vrclog.ResourceURLObserved); ok && obs.RuleID != "youtube_resolve_url" {
				t.Errorf("unexpected community.yamaplayer resource observation for rule %s", obs.RuleID)
			}
		}
	}

	if yamaSourceCount != 1 {
		t.Errorf("yamaSourceCount = %d, want 1", yamaSourceCount)
	}
	if coreResolveCount != 1 {
		t.Errorf("coreResolveCount = %d, want 1", coreResolveCount)
	}
	if coreResolvedCount != 1 {
		t.Errorf("coreResolvedCount = %d, want 1", coreResolvedCount)
	}
	if yamaErrorCount != 1 {
		t.Errorf("yamaErrorCount = %d, want 1", yamaErrorCount)
	}

	// yamaplayer adapter must not duplicate the relay URL emitted by core.
	for _, obs := range observations {
		if obs.AdapterID != "community.yamaplayer" {
			continue
		}
		if ev, ok := obs.Event.(vrclog.ResourceURLObserved); ok {
			if ev.Resource.URL == "https://relay.example.invalid/TESTVIDEO01" {
				t.Errorf("community.yamaplayer duplicated core's relay URL")
			}
		}
	}
}
