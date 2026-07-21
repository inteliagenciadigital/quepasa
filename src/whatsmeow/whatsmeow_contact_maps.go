package whatsmeow

import (
	"strings"
	"sync"

	library "github.com/inteliagenciadigital/quepasa/library"
)

// WhatsmeowContactMaps provides thread-safe mapping for LID/Phone relationships
type WhatsmeowContactMaps struct {
	mutex        sync.RWMutex
	lidToPhone   map[string]string // LID -> Phone mapping
	phoneToLID   map[string]string // Phone -> LID mapping
	isOnWhatsApp map[string]string // raw phone input -> resolved JID (or "" if not registered)
}

var (
	// Global singleton instance
	globalContactMaps *WhatsmeowContactMaps
	// Mutex to ensure thread-safe singleton initialization
	once sync.Once
)

// GetGlobalContactMaps returns the singleton instance of contact maps
func GetGlobalContactMaps() *WhatsmeowContactMaps {
	once.Do(func() {
		globalContactMaps = &WhatsmeowContactMaps{
			lidToPhone:   make(map[string]string),
			phoneToLID:   make(map[string]string),
			isOnWhatsApp: make(map[string]string),
		}
	})
	return globalContactMaps
}

// normalizePhone removes the "+" prefix for storage
func normalizePhone(phone string) string {
	return strings.TrimPrefix(phone, "+")
}

// formatPhone adds the "+" prefix for return
func formatPhone(phone string) string {
	if phone == "" {
		return ""
	}
	if strings.HasPrefix(phone, "+") {
		return phone
	}
	return "+" + phone
}

// GetPhoneFromLIDMap retrieves phone from LID mapping if exists (returns with + prefix)
func (c *WhatsmeowContactMaps) GetPhoneFromLIDMap(lid string) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	phone, exists := c.lidToPhone[lid]
	if exists {
		return formatPhone(phone), true
	}
	return "", false
}

// SetPhoneFromLIDMap stores LID->Phone mapping (stores phone without + prefix).
// BR mobile phones are automatically canonicalized to 9-digit form before storage.
func (c *WhatsmeowContactMaps) SetPhoneFromLIDMap(lid, phone string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Normalize phone for storage (remove + prefix) and canonicalize BR mobile
	normalizedPhone := library.NormalizeCanonicalPhone(normalizePhone(phone))

	// Avoid downgrading a 9-digit Brazilian number (13 digits total) to 8-digit (12 digits total)
	existing, exists := c.lidToPhone[lid]
	if !exists {
		c.lidToPhone[lid] = normalizedPhone
	} else {
		if strings.HasPrefix(existing, "55") && len(existing) == 13 &&
			strings.HasPrefix(normalizedPhone, "55") && len(normalizedPhone) == 12 {
			// Keep the 9-digit version
		} else {
			c.lidToPhone[lid] = normalizedPhone
		}
	}

	// Also store reverse mapping if phone is not empty
	if len(normalizedPhone) > 0 {
		c.phoneToLID[normalizedPhone] = lid

		// Index aliases (e.g. 8-digit BR variant) so reverse lookups work from either form
		for _, alias := range library.BuildPhoneAliases(normalizedPhone) {
			if _, aliasExists := c.phoneToLID[alias]; !aliasExists {
				c.phoneToLID[alias] = lid
			}
		}
	}
}

// GetLIDFromPhoneMap retrieves LID from phone mapping if exists (accepts phone with or without + prefix)
func (c *WhatsmeowContactMaps) GetLIDFromPhoneMap(phone string) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Normalize phone for lookup (remove + prefix)
	normalizedPhone := normalizePhone(phone)
	lid, exists := c.phoneToLID[normalizedPhone]
	return lid, exists
}

// SetLIDFromPhoneMap stores Phone->LID mapping (stores phone without + prefix).
// BR mobile phones are automatically canonicalized to 9-digit form before storage.
func (c *WhatsmeowContactMaps) SetLIDFromPhoneMap(phone, lid string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Normalize phone for storage (remove + prefix) and canonicalize BR mobile
	normalizedPhone := library.NormalizeCanonicalPhone(normalizePhone(phone))
	c.phoneToLID[normalizedPhone] = lid

	// Also store reverse mapping if lid is not empty, but protect 9-digit BR numbers
	if len(lid) > 0 {
		existing, exists := c.lidToPhone[lid]
		if !exists {
			c.lidToPhone[lid] = normalizedPhone
		} else {
			// Only overwrite if upgrading from 8-digit (12 chars starting with 55) to 9-digit (13 chars starting with 55)
			if strings.HasPrefix(existing, "55") && len(existing) == 12 &&
				strings.HasPrefix(normalizedPhone, "55") && len(normalizedPhone) == 13 {
				c.lidToPhone[lid] = normalizedPhone
			}
		}
	}

	// Also index aliases so reverse lookups work from either variant
	for _, alias := range library.BuildPhoneAliases(normalizedPhone) {
		if _, aliasExists := c.phoneToLID[alias]; !aliasExists {
			c.phoneToLID[alias] = lid
		}
	}
}

// GetIsOnWhatsAppCache returns the cached result for a given phone input.
// The second return value is true only when the key has been previously set
// (even if the resolved JID is empty, meaning the number was not found).
//
// WARNING: IsOnWhatsApp queries WhatsApp servers. Calling it too frequently
// may trigger anti-abuse detection and result in account banning.
// Always check this cache before calling the live API.
func (c *WhatsmeowContactMaps) GetIsOnWhatsAppCache(phone string) (jid string, found bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	jid, found = c.isOnWhatsApp[normalizePhone(phone)]
	return
}

// SetIsOnWhatsAppCache stores the resolved JID for a given phone input.
// Pass an empty jid string to record that the number is NOT on WhatsApp.
func (c *WhatsmeowContactMaps) SetIsOnWhatsAppCache(phone, jid string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.isOnWhatsApp[normalizePhone(phone)] = jid
}
