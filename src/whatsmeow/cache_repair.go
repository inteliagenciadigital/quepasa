package whatsmeow

import (
	"strings"

	library "github.com/inteliagenciadigital/quepasa/library"
	log "github.com/sirupsen/logrus"
)

// RepairBRPhoneCache iterates over the in-memory contact map and promotes any
// legacy 8-digit Brazilian mobile numbers to their canonical 9-digit form.
// It also ensures that the reverse mappings and aliases are correctly indexed.
//
// Returns the number of entries that were repaired.
func RepairBRPhoneCache(maps *WhatsmeowContactMaps) int {
	if maps == nil {
		return 0
	}

	maps.mutex.Lock()
	defer maps.mutex.Unlock()

	repairedCount := 0
	updates := make(map[string]string)

	// First pass: identify all 8-digit BR mobile numbers in the LID->Phone map
	for lid, phone := range maps.lidToPhone {
		if strings.HasPrefix(phone, "55") && len(phone) == 12 && library.IsBRMobileCandidate(phone) {
			canonicalPhone := library.NormalizeCanonicalPhone(phone)
			if canonicalPhone != phone {
				updates[lid] = canonicalPhone
			}
		}
	}

	// Second pass: apply updates to LID->Phone and index in Phone->LID
	for lid, canonicalPhone := range updates {
		// Remove the old 8-digit entry from reverse map (if it was the primary)
		// We actually don't want to remove it from phoneToLID because it serves as an alias,
		// but we need to ensure the canonical phone also points to the LID.
		oldPhone := maps.lidToPhone[lid]

		// Update primary mapping
		maps.lidToPhone[lid] = canonicalPhone
		maps.phoneToLID[canonicalPhone] = lid

		// Ensure the old 8-digit phone remains as an alias pointing to the same LID
		maps.phoneToLID[oldPhone] = lid

		// Also index any other generated aliases just to be safe
		for _, alias := range library.BuildPhoneAliases(canonicalPhone) {
			maps.phoneToLID[alias] = lid
		}

		repairedCount++
	}

	if repairedCount > 0 {
		log.Infof("Repaired %d legacy 8-digit BR phone entries in contact cache", repairedCount)
	}

	return repairedCount
}
