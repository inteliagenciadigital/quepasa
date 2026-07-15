package whatsmeow

import (
	"context"
	"errors"
	"fmt"
	"strings"

	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "go.mau.fi/whatsmeow"
	types "go.mau.fi/whatsmeow/types"
	events "go.mau.fi/whatsmeow/types/events"
)

// Compile-time interface check
var _ whatsapp.WhatsappContactManagerInterface = (*WhatsmeowContactManager)(nil)

// WhatsmeowContactManager handles all contact-related operations for WhatsmeowConnection
type WhatsmeowContactManager struct {
	*WhatsmeowConnection                       // embedded connection for direct access
	maps                 *WhatsmeowContactMaps // global contact mappings singleton
}

// NewWhatsmeowContactManager creates a new WhatsmeowContactManager instance
func NewWhatsmeowContactManager(conn *WhatsmeowConnection) *WhatsmeowContactManager {
	return &WhatsmeowContactManager{
		WhatsmeowConnection: conn,
		maps:                GetGlobalContactMaps(), // Use singleton instance
	}
}

// GetContacts returns all contacts from WhatsApp
func (cm *WhatsmeowContactManager) GetContacts() (chats []whatsapp.WhatsappChat, err error) {
	if cm.Client == nil {
		err = errors.New("invalid client")
		return chats, err
	}

	if cm.Client.Store == nil {
		err = errors.New("invalid store")
		return chats, err
	}

	// Delegate to shared helper function
	return GetContactsFromDevice(cm.Client.Store)
}

// IsOnWhatsApp checks if phone numbers are registered on WhatsApp.
//
// WARNING: This method performs a live query against WhatsApp servers.
// Results are cached in-memory (per session singleton) to avoid repeated
// network calls for the same phone number. Bypassing the cache or calling
// with many distinct numbers in a short period may trigger WhatsApp's
// anti-abuse detection and result in account banning.
func (cm *WhatsmeowContactManager) IsOnWhatsApp(phones ...string) (registered []string, err error) {
	var uncached []string

	// Return cached results immediately; collect phones that still need lookup
	for _, phone := range phones {
		if jid, found := cm.maps.GetIsOnWhatsAppCache(phone); found {
			if jid != "" {
				registered = append(registered, jid)
			}
		} else {
			uncached = append(uncached, phone)
		}
	}

	if len(uncached) == 0 {
		return
	}

	// Live query only for phones not yet cached
	results, err := cm.Client.IsOnWhatsApp(context.Background(), uncached)
	if err != nil {
		return
	}

	// Build a set of which uncached phones got a positive result
	resolved := make(map[string]string, len(results))
	for _, result := range results {
		if result.IsIn {
			resolved[result.Query] = result.JID.String()
			registered = append(registered, result.JID.String())
		}
	}

	// Persist results in cache (including negatives, to avoid future lookups)
	for _, phone := range uncached {
		cm.maps.SetIsOnWhatsAppCache(phone, resolved[phone])
	}

	return
}

// GetProfilePicture gets profile picture information
func (cm *WhatsmeowContactManager) GetProfilePicture(wid string, knowingId string) (picture *whatsapp.WhatsappProfilePicture, err error) {
	jid, err := types.ParseJID(wid)
	if err != nil {
		return
	}

	params := &whatsmeow.GetProfilePictureParams{}
	params.ExistingID = knowingId
	params.Preview = false

	pictureInfo, err := cm.Client.GetProfilePictureInfo(context.Background(), jid, params)
	if err != nil {
		return
	}

	if pictureInfo != nil {
		picture = &whatsapp.WhatsappProfilePicture{
			Id:   pictureInfo.ID,
			Type: pictureInfo.Type,
			Url:  pictureInfo.URL,
		}
	}
	return
}

// GetLIDFromPhone returns the @lid for a given phone number
// IMPORTANT: This method accepts phone numbers in E164 format (with +) or without +
// and always normalizes them by removing the + before creating JIDs for WhatsApp API calls
func (cm *WhatsmeowContactManager) GetLIDFromPhone(phone string) (string, error) {
	// Safety check: verify if ContactManager is not nil
	if cm == nil {
		return "", fmt.Errorf("contact manager is nil")
	}

	// Safety check: verify if maps is not nil
	if cm.maps == nil {
		return "", fmt.Errorf("contact maps is nil")
	}

	logger := cm.GetLogger()

	normalized := strings.TrimSpace(phone)
	normalized = strings.TrimPrefix(normalized, "+") // Remove leading + if present

	// First, check maps for existing mapping - this should be the very first check
	if cachedLID, exists := cm.maps.GetLIDFromPhoneMap(normalized); exists {
		logger.Debugf("Found LID in maps for phone %s: %s", phone, cachedLID)
		return cachedLID, nil
	}

	if cm.Client == nil {
		return "", fmt.Errorf("client not defined")
	}

	if cm.Client.Store == nil {
		return "", fmt.Errorf("store not defined")
	}

	logger.Debugf("Phone %s not found in maps, querying database", normalized)

	// Parse the phone number to JID format
	phoneJID := types.JID{
		User:   normalized,
		Server: whatsapp.WHATSAPP_SERVERDOMAIN_USER,
	}

	// try to get the LID from local store
	lidJID, err := cm.Client.Store.LIDs.GetLIDForPN(context.Background(), phoneJID)
	if err == nil && !lidJID.IsEmpty() {
		lid := lidJID.ToNonAD().String()
		logger.Debugf("LID found in database for phone %s: %s", phone, lid)

		// Caching successful mapping for future use
		cm.maps.SetLIDFromPhoneMap(normalized, lid)
		logger.Debugf("Phone->LID mapping cached: %s -> %s", normalized, lid)

		// BR ALIAS: cache the digit-9 variant in memory only — do NOT write to Store.LIDs.
		// Writing PutLIDMapping(lid, aliasJID) corrupts the reverse mapping: whatsmeow's
		// GetPNForLID then returns the alias instead of the canonical phone from the DB.
		// Confirmed in prod logs: DB returned 554796396152 (canonical 8-digit) but
		// PutLIDMapping was promoting 5547996396152 (9-digit alias) as the official mapping.
		// In-memory SetLIDFromPhoneMap preserves the fallback lookup without the corruption.
		if variantPhone, verr := library.AddDigit9BRAllDDDs("+" + normalized); verr == nil {
			variantNormalized := strings.TrimPrefix(variantPhone, "+")
			cm.maps.SetLIDFromPhoneMap(variantNormalized, lid)
			logger.Debugf("BR digit-9 alias cached in-memory only: %s -> %s", variantNormalized, lid)
		} else if variantPhone, verr := library.RemoveDigit9BRAllDDDs("+" + normalized); verr == nil {
			variantNormalized := strings.TrimPrefix(variantPhone, "+")
			cm.maps.SetLIDFromPhoneMap(variantNormalized, lid)
			logger.Debugf("BR digit-9 alias cached in-memory only: %s -> %s", variantNormalized, lid)
		}

		return lid, nil
	}

	// TEMPORARY WORKAROUND: direct lookup failed — try the BR digit-9 variant before giving up.
	// The phone in the store may be the opposite form (e.g. stored as 9-digit, queried as 8-digit).
	// All Brazilian DDDs are tried since the store mapping may exist for any DDD.
	var variantNormalizedFallback string
	if vp, verr := library.AddDigit9BRAllDDDs("+" + normalized); verr == nil {
		variantNormalizedFallback = strings.TrimPrefix(vp, "+")
	} else if vp, verr := library.RemoveDigit9BRAllDDDs("+" + normalized); verr == nil {
		variantNormalizedFallback = strings.TrimPrefix(vp, "+")
	}

	if variantNormalizedFallback != "" {
		// Check in-memory map for variant first
		if cachedLID, exists := cm.maps.GetLIDFromPhoneMap(variantNormalizedFallback); exists {
			logger.Debugf("LID found in maps via BR digit-9 variant for phone %s: %s", phone, cachedLID)
			cm.maps.SetLIDFromPhoneMap(normalized, cachedLID)
			return cachedLID, nil
		}

		variantJIDFallback := types.JID{User: variantNormalizedFallback, Server: whatsapp.WHATSAPP_SERVERDOMAIN_USER}
		if lidJID2, err2 := cm.Client.Store.LIDs.GetLIDForPN(context.Background(), variantJIDFallback); err2 == nil && !lidJID2.IsEmpty() {
			lid := lidJID2.ToNonAD().String()
			logger.Debugf("LID found in database via BR digit-9 variant %s: %s", variantNormalizedFallback, lid)

			// Cache both the original and the variant so future lookups skip DB
			cm.maps.SetLIDFromPhoneMap(normalized, lid)
			cm.maps.SetLIDFromPhoneMap(variantNormalizedFallback, lid)

			// Also persist original form in Store.LIDs so whatsmeow's internal path finds it
			originalJID := types.JID{User: normalized, Server: whatsapp.WHATSAPP_SERVERDOMAIN_USER}
			if perr := cm.Client.Store.LIDs.PutLIDMapping(context.Background(), lidJID2, originalJID); perr != nil {
				logger.Warnf("BR digit-9 fallback: Store.LIDs write failed for original %s: %v", normalized, perr)
			}

			return lid, nil
		}
	}

	logger.Debugf("No LID mapping found for phone %s", normalized)
	return "", nil
}

// GetPhoneFromLID returns the phone number for a given @lid
func (cm *WhatsmeowContactManager) GetPhoneFromLID(lid string) (string, error) {
	// Safety check: verify if ContactManager is not nil
	if cm == nil {
		return "", fmt.Errorf("contact manager is nil")
	}

	// Safety check: verify if maps is not nil
	if cm.maps == nil {
		return "", fmt.Errorf("contact maps is nil")
	}

	logger := cm.GetLogger()

	// First, check maps for existing mapping - this should be the very first check
	if cachedPhone, exists := cm.maps.GetPhoneFromLIDMap(lid); exists {
		logger.Debugf("Found phone in maps for LID %s: %s", lid, cachedPhone)
		return cachedPhone, nil
	}

	if cm.Client == nil {
		return "", fmt.Errorf("client not defined")
	}

	if cm.Client.Store == nil {
		return "", fmt.Errorf("store not defined")
	}

	logger.Debugf("LID %s not found in maps, querying database", lid)

	// Parse the LID to JID format
	lidJID, err := types.ParseJID(lid)
	if err != nil {
		return "", fmt.Errorf("invalid LID format: %v", err)
	}

	// Get the corresponding phone number from local store
	phoneJID, err := cm.Client.Store.LIDs.GetPNForLID(context.Background(), lidJID)
	if err != nil {
		return "", fmt.Errorf("failed to get phone for LID %s: %v", lid, err)
	}

	if phoneJID.IsEmpty() {
		return "", fmt.Errorf("no phone found for LID %s", lid)
	}

	phone := phoneJID.User
	logger.Debugf("Phone found in database for LID %s: %s", lid, phone)

	// Canonicalize BR mobile phones: promote 8-digit to 9-digit when applicable.
	// The whatsmeow DB may store the legacy 8-digit form; we always want canonical.
	canonicalPhone := library.NormalizeCanonicalPhone(phone)
	if canonicalPhone != phone {
		logger.Infof("phone canonicalized: input=%s, canonical=%s, lid=%s", phone, canonicalPhone, lid)
		phone = canonicalPhone
	}

	// Store successful mapping for future use
	cm.maps.SetPhoneFromLIDMap(lid, phone)
	logger.Debugf("LID->Phone mapping stored: %s -> %s", lid, phone)

	// BR ALIAS (reverse): cache the digit-9 variant in memory only — do NOT write to Store.LIDs.
	// PutLIDMapping(lid, aliasJID) corrupts the reverse mapping: whatsmeow's GetPNForLID
	// then returns the alias instead of the canonical phone returned by the DB.
	// Confirmed in prod logs (line 17): "BR digit-9 variant persisted in Store.LIDs (reverse):
	// 35386755649716@lid -> 5547996396152" caused subsequent lookups to return wrong number
	// instead of canonical 554796396152. In-memory cache preserves fallback without corruption.
	if variantPhone, verr := library.AddDigit9BRAllDDDs("+" + phone); verr == nil {
		variantNormalized := strings.TrimPrefix(variantPhone, "+")
		cm.maps.SetLIDFromPhoneMap(variantNormalized, lid)
		logger.Debugf("BR digit-9 alias cached in-memory only (reverse): %s -> %s", variantNormalized, lid)
	} else if variantPhone, verr := library.RemoveDigit9BRAllDDDs("+" + phone); verr == nil {
		variantNormalized := strings.TrimPrefix(variantPhone, "+")
		cm.maps.SetLIDFromPhoneMap(variantNormalized, lid)
		logger.Debugf("BR digit-9 alias cached in-memory only (reverse): %s -> %s", variantNormalized, lid)
	}

	// Retrieve the phone number again from the maps cache to ensure we return the upgraded/promoted 9-digit version
	if upgradedPhone, ok := cm.maps.GetPhoneFromLIDMap(lid); ok {
		phone = strings.TrimPrefix(upgradedPhone, "+")
	}

	return phone, nil
}

// GetPhoneFromStore retrieves the phone number directly from the whatsmeow store for any JID type
// This method accesses the Store directly without going through additional layers
//
// @param jid The JID to get the phone for (works for @lid and @s.whatsapp.net)
// @return The raw phone number (without E.164 formatting) or empty string if not found
func (cm *WhatsmeowContactManager) GetPhoneFromStore(jid types.JID) string {
	if cm.Client == nil || cm.Client.Store == nil {
		return ""
	}

	// For @lid contacts, use GetPNForLID
	if jid.Server == whatsapp.WHATSAPP_SERVERDOMAIN_LID {
		if cm.Client.Store.LIDs != nil {
			if pnJID, err := cm.Client.Store.LIDs.GetPNForLID(context.Background(), jid.ToNonAD()); err == nil && !pnJID.IsEmpty() && len(pnJID.User) > 0 {
				return pnJID.User
			}
		}
		return ""
	}

	// For @s.whatsapp.net, just return the User part
	return jid.User
}

// GetUserInfo retrieves comprehensive user information for given JIDs
func (cm *WhatsmeowContactManager) GetUserInfo(jids []string) ([]interface{}, error) {
	if cm.Client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	if cm.Client.Store == nil {
		return nil, fmt.Errorf("store not defined")
	}

	// Convert string JIDs to types.JID
	var parsedJIDs []types.JID
	for _, jidStr := range jids {
		// Check if it's a phone number (no @ symbol) and validate E164 format
		if !strings.Contains(jidStr, "@") {
			// This is a phone number, validate and format to E164
			validPhone, err := whatsapp.GetPhoneIfValid(jidStr)
			if err != nil {
				return nil, fmt.Errorf("invalid phone number format for %s: %v (must be E164 format starting with +)", jidStr, err)
			}

			// Remove the + from E164 format for JID creation
			phoneNumber := strings.TrimPrefix(validPhone, "+")
			jid := types.JID{
				User:   phoneNumber,
				Server: whatsapp.WHATSAPP_SERVERDOMAIN_USER,
			}
			parsedJIDs = append(parsedJIDs, jid)
		} else {
			// This is already a JID, parse normally
			jid, err := types.ParseJID(jidStr)
			if err != nil {
				return nil, fmt.Errorf("invalid JID format for %s: %v", jidStr, err)
			}
			parsedJIDs = append(parsedJIDs, jid)
		}
	}

	// Get user info from WhatsApp - this returns a map[types.JID]types.UserInfo
	userInfoMap, err := cm.Client.GetUserInfo(context.Background(), parsedJIDs)
	logentry := cm.GetLogger()
	logentry.Debugf("GetUserInfo for JIDs: %v, result: %v", parsedJIDs, userInfoMap)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}

	// Convert map to interface array for generic return type
	result := make([]interface{}, 0, len(userInfoMap))
	for jid, info := range userInfoMap {
		// Get contact info from local store - try both JID and corresponding phone/LID
		contactInfo, contactErr := cm.Client.Store.Contacts.GetContact(context.TODO(), jid)

		// Get LID/Phone mapping information
		var lid, phoneNumber string
		var phoneJID types.JID

		if strings.Contains(jid.String(), whatsapp.WHATSAPP_SERVERDOMAIN_LID_SUFFIX) {
			// This is a LID, try to get corresponding phone
			lid = jid.ToNonAD().String()
			pnJID, err := cm.Client.Store.LIDs.GetPNForLID(context.TODO(), jid)
			if err == nil && !pnJID.IsEmpty() {
				phoneNumber = pnJID.User
				phoneJID = pnJID

				// If we didn't get contact info from LID, try with phone JID
				if contactErr != nil {
					contactInfo, contactErr = cm.Client.Store.Contacts.GetContact(context.TODO(), phoneJID)
				}
			}
		} else {
			// This is a phone number JID, try to get corresponding LID
			phoneNumber = jid.User
			lidJID, err := cm.Client.Store.LIDs.GetLIDForPN(context.TODO(), jid)
			if err == nil && !lidJID.IsEmpty() {
				lid = lidJID.ToNonAD().String()

				// If we didn't get contact info from phone JID, try with LID
				if contactErr != nil {
					contactInfo, contactErr = cm.Client.Store.Contacts.GetContact(context.TODO(), lidJID)
				}
			}
		}

		// Format phone to E164 if available
		var phoneE164 string
		if phoneNumber != "" {
			if phone, err := whatsapp.GetPhoneIfValid(phoneNumber); err == nil {
				phoneE164 = phone
			}
		}

		// Determine the best display name
		var displayName string
		if contactErr == nil {
			if contactInfo.FullName != "" {
				displayName = contactInfo.FullName
			} else if contactInfo.BusinessName != "" {
				displayName = contactInfo.BusinessName
			} else if contactInfo.PushName != "" {
				displayName = contactInfo.PushName
			}
		}

		// If no local contact name, use verified name from user info
		if displayName == "" && info.VerifiedName != nil {
			displayName = info.VerifiedName.Details.GetVerifiedName()
		}

		// Check if we have meaningful contact information
		hasContactInfo := contactErr == nil && (contactInfo.FullName != "" || contactInfo.BusinessName != "" || contactInfo.PushName != "")
		hasVerifiedName := info.VerifiedName != nil && info.VerifiedName.Details.GetVerifiedName() != ""
		hasStatus := info.Status != ""
		hasPictureID := info.PictureID != ""
		hasDevices := len(info.Devices) > 0
		hasLID := lid != ""

		// Only include contacts that have meaningful information beyond just phone/JID
		if !hasContactInfo && !hasVerifiedName && !hasStatus && !hasPictureID && !hasDevices && !hasLID {
			logentry.Debugf("Skipping contact %s - no meaningful information possible non whatsapp number", jid.String())
			continue
		}

		// Create a comprehensive response with omitempty support
		userInfoResponse := WhatsmeowUserInfoResponse{
			JID:          jid.String(),
			LID:          lid,
			Phone:        phoneNumber,
			PhoneE164:    phoneE164,
			Status:       info.Status,
			PictureID:    info.PictureID,
			Devices:      info.Devices,
			VerifiedName: info.VerifiedName,
			DisplayName:  displayName,
		}

		// Add contact-specific information if available
		if contactErr == nil {
			userInfoResponse.FullName = contactInfo.FullName
			userInfoResponse.BusinessName = contactInfo.BusinessName
			userInfoResponse.PushName = contactInfo.PushName
		}

		result = append(result, userInfoResponse)
	}

	return result, nil
}

// BlockContact blocks a contact by their WID/JID so they cannot send messages to this account.
func (cm *WhatsmeowContactManager) BlockContact(wid string) error {
	if cm.Client == nil {
		return fmt.Errorf("client not defined")
	}

	jid, err := types.ParseJID(wid)
	if err != nil {
		return fmt.Errorf("invalid contact id: %w", err)
	}

	_, err = cm.Client.UpdateBlocklist(context.Background(), jid, events.BlocklistChangeActionBlock)
	return err
}

// UnblockContact removes a block previously placed on a contact.
func (cm *WhatsmeowContactManager) UnblockContact(wid string) error {
	if cm.Client == nil {
		return fmt.Errorf("client not defined")
	}

	jid, err := types.ParseJID(wid)
	if err != nil {
		return fmt.Errorf("invalid contact id: %w", err)
	}

	_, err = cm.Client.UpdateBlocklist(context.Background(), jid, events.BlocklistChangeActionUnblock)
	return err
}

// GetPhoneFromContactId attempts to get phone number from contact Id using available mapping
func (cm *WhatsmeowContactManager) GetPhoneFromContactId(contactId string) (string, error) {
	if strings.Contains(contactId, whatsapp.WHATSAPP_SERVERDOMAIN_USER_SUFFIX) {
		phone, err := whatsapp.GetPhoneIfValid(contactId)
		if err == nil {
			return phone, nil // Return phone if valid
		}
	}

	logentry := cm.GetLogger()
	logentry = logentry.WithField("entry", "WhatsmeowContactManager.GetPhoneFromContactId")

	// Try to get phone from different sources
	if strings.Contains(contactId, whatsapp.WHATSAPP_SERVERDOMAIN_LID_SUFFIX) {

		logentry.Debug("Attempting to get phone from LId")

		// For @lid, try to get the corresponding phone number using contact manager interface
		if retrieved, err := cm.GetPhoneFromLID(contactId); err == nil && len(retrieved) > 0 {
			logentry.Debugf("Retrieved phone from LId mapping: %s", retrieved)

			// Canonicalize BR mobile phones before formatting to E164
			canonicalRetrieved := library.NormalizeCanonicalPhone(strings.TrimPrefix(retrieved, "+"))
			if canonicalRetrieved != strings.TrimPrefix(retrieved, "+") {
				logentry.Infof("phone from LID canonicalized: input=%s, canonical=%s, contactId=%s", retrieved, canonicalRetrieved, contactId)
				retrieved = canonicalRetrieved
			}

			// Format the phone to E164 if needed
			if phone, err := whatsapp.GetPhoneIfValid(retrieved); err == nil {
				logentry.Debug("Phone formatted to E164")
				return phone, nil
			}
		} else {
			logentry.WithError(err).Error("Failed to get phone from LID mapping")
			return "", err
		}
	}

	// If still not found, return error
	logentry.Infof("Can't find suitable E164 phone for contact Id: %s", contactId)
	return "", fmt.Errorf("no phone found for contact Id: %s", contactId)
}
