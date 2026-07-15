package library

import (
	"testing"
)

// =============================================================================
// NormalizeCanonicalPhone tests
// =============================================================================

func TestNormalizeCanonicalPhone_BRMobileWith9(t *testing.T) {
	// Already canonical — should be idempotent
	tests := []struct {
		input    string
		expected string
	}{
		{"5511965647131", "5511965647131"},
		{"5547996396152", "5547996396152"},
		{"5521999887766", "5521999887766"},
		{"5531987654321", "5531987654321"},
		{"5561998765432", "5561998765432"},
	}

	for _, tt := range tests {
		result := NormalizeCanonicalPhone(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeCanonicalPhone(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeCanonicalPhone_BRMobileWithout9(t *testing.T) {
	// 8-digit mobile must be upgraded to 9-digit
	tests := []struct {
		input    string
		expected string
	}{
		{"551165647131", "5511965647131"},   // SP DDD 11
		{"554796396152", "5547996396152"},   // SC DDD 47
		{"552199887766", "5521999887766"},   // RJ DDD 21
		{"553187654321", "5531987654321"},   // MG DDD 31
		{"556198765432", "5561998765432"},   // DF DDD 61
		{"551298765432", "5512998765432"},   // SP DDD 12
		{"551998765432", "5519998765432"},   // SP DDD 19
		{"554198765432", "5541998765432"},   // PR DDD 41
		{"557198765432", "5571998765432"},   // BA DDD 71
		{"558598765432", "5585998765432"},   // CE DDD 85
	}

	for _, tt := range tests {
		result := NormalizeCanonicalPhone(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeCanonicalPhone(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeCanonicalPhone_BRLandline(t *testing.T) {
	// Landlines must NOT get a 9th digit inserted
	tests := []struct {
		input    string
		expected string
	}{
		{"551133334444", "551133334444"},   // SP landline (starts with 3)
		{"551144445555", "551144445555"},   // SP landline (starts with 4)
		{"552122223333", "552122223333"},   // RJ landline (starts with 2)
		{"553134567890", "553134567890"},   // MG landline (starts with 3)
		{"554733334444", "554733334444"},   // SC landline (starts with 3)
	}

	for _, tt := range tests {
		result := NormalizeCanonicalPhone(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeCanonicalPhone(%q) = %q, want %q (landline must not change)", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeCanonicalPhone_International(t *testing.T) {
	// International numbers must pass through unchanged
	tests := []struct {
		input    string
		expected string
	}{
		{"14155551234", "14155551234"},       // USA
		{"447911123456", "447911123456"},     // UK
		{"34612345678", "34612345678"},       // Spain
		{"4915112345678", "4915112345678"},   // Germany
		{"819012345678", "819012345678"},     // Japan
		{"5491112345678", "5491112345678"},   // Argentina
	}

	for _, tt := range tests {
		result := NormalizeCanonicalPhone(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeCanonicalPhone(%q) = %q, want %q (international must not change)", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeCanonicalPhone_WithPlusPrefix(t *testing.T) {
	// Should strip + prefix and still canonicalize
	result := NormalizeCanonicalPhone("+554796396152")
	expected := "5547996396152"
	if result != expected {
		t.Errorf("NormalizeCanonicalPhone(+554796396152) = %q, want %q", result, expected)
	}
}

func TestNormalizeCanonicalPhone_EmptyAndEdgeCases(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"+", ""},
		{"55", "55"},       // Too short for any rule
		{"551", "551"},     // Too short
		{"5511", "5511"},   // Too short
	}

	for _, tt := range tests {
		result := NormalizeCanonicalPhone(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeCanonicalPhone(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeCanonicalPhone_Idempotent(t *testing.T) {
	// Applying twice should give the same result
	inputs := []string{
		"5511965647131",  // Already canonical
		"551165647131",   // Will be upgraded
		"551133334444",   // Landline
		"14155551234",    // International
	}

	for _, input := range inputs {
		first := NormalizeCanonicalPhone(input)
		second := NormalizeCanonicalPhone(first)
		if first != second {
			t.Errorf("NormalizeCanonicalPhone is not idempotent for %q: first=%q, second=%q", input, first, second)
		}
	}
}

// =============================================================================
// IsBRMobileCandidate tests
// =============================================================================

func TestIsBRMobileCandidate(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"551165647131", true},   // subscriber starts with 6
		{"554796396152", true},   // subscriber starts with 9
		{"551185551234", true},   // subscriber starts with 8
		{"551175551234", true},   // subscriber starts with 7
		{"551155551234", true},   // subscriber starts with 5
		{"551133334444", false},  // subscriber starts with 3 (landline)
		{"551144445555", false},  // subscriber starts with 4 (landline)
		{"551122223333", false},  // subscriber starts with 2 (landline)
		{"14155551234", false},   // not BR
		{"5511965647131", false}, // 13 digits (wrong function)
		{"55116564", false},      // too short
	}

	for _, tt := range tests {
		result := IsBRMobileCandidate(tt.input)
		if result != tt.expected {
			t.Errorf("IsBRMobileCandidate(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

// =============================================================================
// IsBRMobile13Digit tests
// =============================================================================

func TestIsBRMobile13Digit(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"5511965647131", true},   // 9 + subscriber 6
		{"5547996396152", true},   // 9 + subscriber 9
		{"5521987654321", true},   // 9 + subscriber 8
		{"5511933334444", false},  // 9 + subscriber 3 (not mobile pattern)
		{"5511865647131", false},  // no 9 at position 4
		{"551165647131", false},   // 12 digits (wrong function)
		{"14155551234", false},    // not BR
	}

	for _, tt := range tests {
		result := IsBRMobile13Digit(tt.input)
		if result != tt.expected {
			t.Errorf("IsBRMobile13Digit(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

// =============================================================================
// BuildPhoneAliases tests
// =============================================================================

func TestBuildPhoneAliases_BRMobileWith9(t *testing.T) {
	// Should generate 8-digit alias
	aliases := BuildPhoneAliases("5511965647131")
	if len(aliases) != 1 {
		t.Fatalf("BuildPhoneAliases(5511965647131) returned %d aliases, want 1", len(aliases))
	}
	if aliases[0] != "551165647131" {
		t.Errorf("BuildPhoneAliases(5511965647131)[0] = %q, want %q", aliases[0], "551165647131")
	}
}

func TestBuildPhoneAliases_BRMobileWith9_DDD47(t *testing.T) {
	aliases := BuildPhoneAliases("5547996396152")
	if len(aliases) != 1 {
		t.Fatalf("BuildPhoneAliases(5547996396152) returned %d aliases, want 1", len(aliases))
	}
	if aliases[0] != "554796396152" {
		t.Errorf("BuildPhoneAliases(5547996396152)[0] = %q, want %q", aliases[0], "554796396152")
	}
}

func TestBuildPhoneAliases_BRLandline(t *testing.T) {
	aliases := BuildPhoneAliases("551133334444")
	if len(aliases) != 0 {
		t.Errorf("BuildPhoneAliases(551133334444) returned %d aliases, want 0 (landline)", len(aliases))
	}
}

func TestBuildPhoneAliases_International(t *testing.T) {
	aliases := BuildPhoneAliases("14155551234")
	if len(aliases) != 0 {
		t.Errorf("BuildPhoneAliases(14155551234) returned %d aliases, want 0 (international)", len(aliases))
	}
}

func TestBuildPhoneAliases_Empty(t *testing.T) {
	aliases := BuildPhoneAliases("")
	if len(aliases) != 0 {
		t.Errorf("BuildPhoneAliases('') returned %d aliases, want 0", len(aliases))
	}
}

func TestBuildPhoneAliases_AliasNeverEqualsCanonical(t *testing.T) {
	canonicals := []string{
		"5511965647131",
		"5547996396152",
		"551133334444",
		"14155551234",
	}

	for _, canonical := range canonicals {
		aliases := BuildPhoneAliases(canonical)
		for _, alias := range aliases {
			if alias == canonical {
				t.Errorf("BuildPhoneAliases(%q) returned alias equal to canonical: %q", canonical, alias)
			}
		}
	}
}

// =============================================================================
// ClassifyBRPhone tests
// =============================================================================

func TestClassifyBRPhone(t *testing.T) {
	tests := []struct {
		input         string
		wantCountry   string
		wantPhoneType string
	}{
		{"5511965647131", "BR", "mobile"},
		{"5547996396152", "BR", "mobile"},
		{"551165647131", "BR", "mobile"},      // 8-digit mobile candidate
		{"551133334444", "BR", "landline"},
		{"552122223333", "BR", "landline"},
		{"14155551234", "international", "unknown"},
		{"447911123456", "international", "unknown"},
	}

	for _, tt := range tests {
		country, phoneType := ClassifyBRPhone(tt.input)
		if country != tt.wantCountry || phoneType != tt.wantPhoneType {
			t.Errorf("ClassifyBRPhone(%q) = (%q, %q), want (%q, %q)",
				tt.input, country, phoneType, tt.wantCountry, tt.wantPhoneType)
		}
	}
}

// =============================================================================
// Scenario tests (from PRD acceptance criteria)
// =============================================================================

func TestScenario1_BRMobileCorrectPreserved(t *testing.T) {
	// Given a correct BR mobile number with 9
	// When canonicalized
	// Then the 9 must be preserved
	input := "5511965647131"
	result := NormalizeCanonicalPhone(input)
	if result != input {
		t.Errorf("Scenario 1 FAILED: correct BR mobile changed from %q to %q", input, result)
	}
}

func TestScenario2_AliasWithout9Reconciled(t *testing.T) {
	// Given an alias without 9
	// When the system reconciles with canonical
	// Then canonical must be the 9-digit version
	alias := "551165647131"
	canonical := NormalizeCanonicalPhone(alias)
	expected := "5511965647131"
	if canonical != expected {
		t.Errorf("Scenario 2 FAILED: alias %q should canonicalize to %q, got %q", alias, expected, canonical)
	}

	// And the alias must exist as a search alias, not as canonical
	aliases := BuildPhoneAliases(canonical)
	found := false
	for _, a := range aliases {
		if a == alias {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Scenario 2 FAILED: alias %q not found in BuildPhoneAliases(%q)", alias, canonical)
	}
}

func TestScenario4_LandlineBR(t *testing.T) {
	// Given a Brazilian landline number
	// When canonicalized
	// Then it must NOT get a 9 inserted
	input := "551133334444"
	result := NormalizeCanonicalPhone(input)
	if result != input {
		t.Errorf("Scenario 4 FAILED: landline %q changed to %q", input, result)
	}
}

func TestScenario5_International(t *testing.T) {
	// Given a non-Brazilian number
	// When canonicalized
	// Then no BR rules must be applied
	input := "14155551234"
	result := NormalizeCanonicalPhone(input)
	if result != input {
		t.Errorf("Scenario 5 FAILED: international %q changed to %q", input, result)
	}
}
