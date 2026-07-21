package runtime

import (
	"testing"

	"github.com/inteliagenciadigital/quepasa/models"
)

func TestChooseSpamSectionTokenUsesLowestReadyPriority(t *testing.T) {
	sections := []*models.QpSpamSection{
		{Token: "first-offline", Priority: 1, Enabled: true},
		{Token: "second-ready", Priority: 2, Enabled: true},
		{Token: "third-ready", Priority: 3, Enabled: true},
	}

	got, err := chooseSpamSectionToken(sections, func(token string) bool {
		return token == "second-ready" || token == "third-ready"
	}, func(int) int {
		t.Fatal("random picker should not be called for a single ready section")
		return 0
	})
	if err != nil {
		t.Fatalf("choose spam token: %v", err)
	}
	if got != "second-ready" {
		t.Fatalf("expected lowest ready priority token, got %q", got)
	}
}

func TestChooseSpamSectionTokenRandomizesWithinSamePriority(t *testing.T) {
	sections := []*models.QpSpamSection{
		{Token: "alpha", Priority: 1, Enabled: true},
		{Token: "beta", Priority: 1, Enabled: true},
		{Token: "gamma", Priority: 2, Enabled: true},
	}

	got, err := chooseSpamSectionToken(sections, func(token string) bool {
		return token == "alpha" || token == "beta" || token == "gamma"
	}, func(n int) int {
		if n != 2 {
			t.Fatalf("expected random pool with 2 same-priority tokens, got %d", n)
		}
		return 1
	})
	if err != nil {
		t.Fatalf("choose spam token: %v", err)
	}
	if got != "beta" {
		t.Fatalf("expected selected same-priority token, got %q", got)
	}
}

func TestChooseSpamSectionTokenFallsThroughToNextReadyPriorityGroup(t *testing.T) {
	sections := []*models.QpSpamSection{
		{Token: "offline-a", Priority: 1, Enabled: true},
		{Token: "offline-b", Priority: 1, Enabled: true},
		{Token: "ready-a", Priority: 2, Enabled: true},
		{Token: "ready-b", Priority: 2, Enabled: true},
		{Token: "ready-c", Priority: 3, Enabled: true},
	}

	got, err := chooseSpamSectionToken(sections, func(token string) bool {
		return token == "ready-a" || token == "ready-b" || token == "ready-c"
	}, func(n int) int {
		if n != 2 {
			t.Fatalf("expected random pool with next ready priority group, got %d", n)
		}
		return 1
	})
	if err != nil {
		t.Fatalf("choose spam token: %v", err)
	}
	if got != "ready-b" {
		t.Fatalf("expected selected next-priority token, got %q", got)
	}
}

func TestChooseSpamSectionTokenIgnoresDisabledSections(t *testing.T) {
	sections := []*models.QpSpamSection{
		{Token: "disabled", Priority: 1, Enabled: false},
		{Token: "ready", Priority: 2, Enabled: true},
	}

	got, err := chooseSpamSectionToken(sections, func(token string) bool {
		return token == "disabled" || token == "ready"
	}, nil)
	if err != nil {
		t.Fatalf("choose spam token: %v", err)
	}
	if got != "ready" {
		t.Fatalf("expected enabled ready token, got %q", got)
	}
}
