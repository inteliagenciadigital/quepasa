package whatsmeow

import (
	"testing"
)

// TestPhoneCanonical_Integration maps out the interactions between 
// the in-memory map logic and the new canonicalization library.
func TestPhoneCanonical_Integration(t *testing.T) {
	maps := GetGlobalContactMaps()

	// Clear maps for testing
	maps.mutex.Lock()
	maps.lidToPhone = make(map[string]string)
	maps.phoneToLID = make(map[string]string)
	maps.mutex.Unlock()

	lid := "123456789"
	legacyPhone := "551165647131" // 8-digit BR mobile
	canonicalPhone := "5511965647131" // 9-digit BR mobile

	// 1. Setting legacy phone should auto-canonicalize
	maps.SetPhoneFromLIDMap(lid, legacyPhone)

	// Get should return canonical
	retrievedPhone, exists := maps.GetPhoneFromLIDMap(lid)
	if !exists {
		t.Fatalf("GetPhoneFromLIDMap failed to find lid %s", lid)
	}

	// Note: GetPhoneFromLIDMap returns with "+" prefix
	if retrievedPhone != "+"+canonicalPhone {
		t.Errorf("Expected canonical phone +%s, got %s", canonicalPhone, retrievedPhone)
	}

	// 2. Reverse lookup by legacy phone should work (via alias indexing)
	retrievedLID, exists := maps.GetLIDFromPhoneMap(legacyPhone)
	if !exists {
		t.Fatalf("GetLIDFromPhoneMap failed to find legacy phone %s", legacyPhone)
	}
	if retrievedLID != lid {
		t.Errorf("Expected LID %s, got %s", lid, retrievedLID)
	}

	// 3. Reverse lookup by canonical phone should work
	retrievedLIDCanonical, exists := maps.GetLIDFromPhoneMap(canonicalPhone)
	if !exists {
		t.Fatalf("GetLIDFromPhoneMap failed to find canonical phone %s", canonicalPhone)
	}
	if retrievedLIDCanonical != lid {
		t.Errorf("Expected LID %s, got %s", lid, retrievedLIDCanonical)
	}

	// Clear maps
	maps.mutex.Lock()
	maps.lidToPhone = make(map[string]string)
	maps.phoneToLID = make(map[string]string)
	maps.mutex.Unlock()

	// 4. Test SetLIDFromPhoneMap auto-canonicalizes
	maps.SetLIDFromPhoneMap(legacyPhone, lid)

	retrievedPhone2, exists2 := maps.GetPhoneFromLIDMap(lid)
	if !exists2 {
		t.Fatalf("GetPhoneFromLIDMap failed to find lid %s after SetLIDFromPhoneMap", lid)
	}
	if retrievedPhone2 != "+"+canonicalPhone {
		t.Errorf("Expected canonical phone +%s, got %s", canonicalPhone, retrievedPhone2)
	}

	retrievedLID2, existsLID2 := maps.GetLIDFromPhoneMap(legacyPhone)
	if !existsLID2 {
		t.Fatalf("GetLIDFromPhoneMap failed for legacy phone after SetLIDFromPhoneMap")
	}
	if retrievedLID2 != lid {
		t.Errorf("Expected LID %s, got %s", lid, retrievedLID2)
	}
}
