package library

import (
	"strings"
)

// PhoneIdentity holds the canonical representation of a phone number along with
// its classification and any search aliases. The canonical form is the
// authoritative representation used for persistence, cache keys, and message
// sending. Aliases exist solely for lookup/reconciliation purposes.
type PhoneIdentity struct {
	Canonical string   // Authoritative phone number (without + prefix)
	Aliases   []string // Search-only aliases (never used for sending)
	Country   string   // "BR" or "international"
	Type      string   // "mobile", "landline", "unknown"
}

// NormalizeCanonicalPhone returns the canonical form of a phone number.
//
// For Brazilian mobile numbers (country code 55) that are missing the 9th digit,
// the digit is added. Landline numbers and international numbers pass through
// unchanged.
//
// The input and output are both without the "+" prefix.
// If the input has a "+" prefix, it is stripped before processing.
//
// This function is safe to call on numbers that are already canonical — it is
// idempotent.
func NormalizeCanonicalPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = strings.TrimPrefix(phone, "+")

	if len(phone) == 0 {
		return phone
	}

	// Only process Brazilian numbers
	if !strings.HasPrefix(phone, "55") {
		return phone
	}

	// Brazilian number with 12 digits (country code 55 + DDD 2 digits + 8-digit subscriber)
	// This is the legacy format that may be missing the 9th digit for mobile numbers.
	if len(phone) == 12 {
		if IsBRMobileCandidate(phone) {
			// Add the 9th digit: insert "9" after the DDD
			return phone[:4] + "9" + phone[4:]
		}
		// Landline or non-mobile — do not modify
		return phone
	}

	// 13-digit BR number: already has the 9th digit (or is a valid long format)
	// 10-digit or 11-digit: old/short format, do not modify (edge case)
	return phone
}

// IsBRMobileCandidate checks whether a 12-digit Brazilian phone number (without +)
// looks like a mobile number that should have a 9th digit.
//
// Rules (based on ANATEL numbering plan):
//   - Country code: 55
//   - DDD: 2 digits (positions 2-3), any valid DDD
//   - First subscriber digit (position 4): must be 5-9 (mobile range)
//   - Numbers starting with 2, 3, or 4 after the DDD are landline
//
// Format: 55 DD XXXX XXXX  (12 digits total without +)
func IsBRMobileCandidate(phone string) bool {
	phone = strings.TrimPrefix(phone, "+")

	if len(phone) != 12 {
		return false
	}

	if !strings.HasPrefix(phone, "55") {
		return false
	}

	// First subscriber digit is at index 4
	firstSubscriberDigit := phone[4]

	// Mobile numbers have first subscriber digit 5-9
	// Landline numbers have first subscriber digit 2-4
	return firstSubscriberDigit >= '5' && firstSubscriberDigit <= '9'
}

// IsBRMobile13Digit checks whether a 13-digit Brazilian phone number (without +)
// is a mobile number with the 9th digit already present.
//
// Format: 55 DD 9XXXX XXXX (13 digits total without +)
// The 9th digit is at position 4, and the original first subscriber digit
// (now at position 5) must be 5-9.
func IsBRMobile13Digit(phone string) bool {
	phone = strings.TrimPrefix(phone, "+")

	if len(phone) != 13 {
		return false
	}

	if !strings.HasPrefix(phone, "55") {
		return false
	}

	// The inserted 9 is at index 4
	if phone[4] != '9' {
		return false
	}

	// Original first subscriber digit is now at index 5
	originalFirst := phone[5]
	return originalFirst >= '5' && originalFirst <= '9'
}

// BuildPhoneAliases generates search aliases for a canonical phone number.
//
// For Brazilian mobile numbers with the 9th digit, it generates the
// 8-digit variant (without the 9) as a lookup alias.
//
// Aliases are ONLY for search/reconciliation. They must NEVER be used for:
//   - Persistence as the primary phone
//   - Message sending
//   - Cache canonical value
//
// Returns an empty slice for landline and international numbers.
func BuildPhoneAliases(canonicalPhone string) []string {
	canonicalPhone = strings.TrimPrefix(canonicalPhone, "+")

	if len(canonicalPhone) == 0 {
		return nil
	}

	// Only generate aliases for 13-digit BR mobile numbers (with 9th digit)
	if IsBRMobile13Digit(canonicalPhone) {
		// Remove the 9th digit to create the legacy 8-digit alias
		alias := canonicalPhone[:4] + canonicalPhone[5:]
		return []string{alias}
	}

	return nil
}

// ClassifyBRPhone classifies a phone number by country and type.
//
// Returns:
//   - country: "BR" for Brazilian numbers, "international" for others
//   - phoneType: "mobile", "landline", or "unknown"
func ClassifyBRPhone(phone string) (country string, phoneType string) {
	phone = strings.TrimPrefix(phone, "+")

	if !strings.HasPrefix(phone, "55") {
		return "international", "unknown"
	}

	switch len(phone) {
	case 12:
		// 8-digit subscriber: could be landline or legacy mobile
		if IsBRMobileCandidate(phone) {
			return "BR", "mobile"
		}
		// First subscriber digit 2-4 = landline
		if len(phone) > 4 && phone[4] >= '2' && phone[4] <= '4' {
			return "BR", "landline"
		}
		return "BR", "unknown"

	case 13:
		// 9-digit subscriber: modern mobile format
		if IsBRMobile13Digit(phone) {
			return "BR", "mobile"
		}
		return "BR", "unknown"

	default:
		return "BR", "unknown"
	}
}
