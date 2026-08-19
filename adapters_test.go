package adapters_test

import (
	"testing"

	vrclog "github.com/vrclog/vrclog-go"

	adapters "github.com/vrclog/vrclog-adapters"
)

func TestAll_ReturnsCorrectCount(t *testing.T) {
	got := adapters.All()
	if len(got) != 2 {
		t.Fatalf("len(All()) = %d, want 2", len(got))
	}
}

func TestAll_DeterministicOrder(t *testing.T) {
	got := adapters.All()
	if got[0].ID() != "community.yamaplayer" {
		t.Errorf("All()[0].ID() = %q, want %q", got[0].ID(), "community.yamaplayer")
	}
	if got[1].ID() != "community.iwasync3" {
		t.Errorf("All()[1].ID() = %q, want %q", got[1].ID(), "community.iwasync3")
	}
}

func TestAll_FreshSlice(t *testing.T) {
	first := adapters.All()
	second := adapters.All()

	first[0] = nil

	if second[0] == nil {
		t.Fatalf("mutating first All() slice affected second call's slice")
	}
}

func TestAll_ValidAdapterIDs(t *testing.T) {
	engine, err := vrclog.NewEngine(adapters.All()...)
	if err != nil {
		t.Fatalf("NewEngine(All()...) error = %v", err)
	}
	if engine == nil {
		t.Fatalf("NewEngine(All()...) returned nil engine")
	}
}
