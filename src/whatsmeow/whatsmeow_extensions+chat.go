package whatsmeow

import (
	"context"

	library "github.com/inteliagenciadigital/quepasa/library"
	whatsapp "github.com/inteliagenciadigital/quepasa/whatsapp"
	whatsmeow "go.mau.fi/whatsmeow"
	types "go.mau.fi/whatsmeow/types"
)

/**
 * CleanJID removes the session suffix from a JID if present.
 * Example: 554792857088:72@s.whatsapp.net -> 554792857088@s.whatsapp.net
 *
 * @param jid The JID to clean
 * @return JID without session suffix
 */
func CleanJID(jid types.JID) types.JID {
	// Always reconstruct the JID using only User and Server
	// This automatically removes any session suffix that might be present in the original string representation
	cleanJID := types.JID{
		User:   jid.User,
		Server: jid.Server,
	}

	return cleanJID
}

/**
 * ExtractContactName extracts the best available contact name following the hierarchy:
 * BusinessName > FullName > PushName > FirstName
 *
 * @param cInfo Contact info from WhatsApp store
 * @return The best available name or empty string if none found
 */
func ExtractContactName(cInfo types.ContactInfo) string {
	if !cInfo.Found {
		return ""
	}
	// Priority: FullName (user's saved name) > BusinessName (business account) > PushName (public name) > FirstName
	if len(cInfo.FullName) > 0 {
		return cInfo.FullName
	}
	if len(cInfo.BusinessName) > 0 {
		return cInfo.BusinessName
	}
	if len(cInfo.PushName) > 0 {
		return cInfo.PushName
	}
	if len(cInfo.FirstName) > 0 {
		return cInfo.FirstName
	}
	return ""
}

func GetContactName(client *whatsmeow.Client, jid types.JID) string {
	if client == nil || client.Store == nil {
		return ""
	}

	if cInfo, err := client.Store.Contacts.GetContact(context.Background(), jid); err == nil && cInfo.Found {
		return ExtractContactName(cInfo)
	}

	return ""
}

func GetChatTitle(client *whatsmeow.Client, jid types.JID) (title string) {
	if client == nil {
		return ""
	}

	if jid.Server == whatsapp.WHATSAPP_SERVERDOMAIN_GROUP {
		title = GroupInfoCache.Get(jid.String())
		if len(title) > 0 {
			goto found
		}
		gInfo, _ := client.GetGroupInfo(context.Background(), jid)
		if gInfo != nil {
			title = gInfo.Name
			_ = GroupInfoCache.Append(jid.String(), title, "GetChatTitle")
			goto found
		}
	} else {
		title = GetContactName(client, jid)
		if len(title) > 0 {
			goto found
		}
	}
	return ""
found:
	return library.NormalizeForTitle(title)
}

/**
 * GetUsernameFromJID retrieves the username for a given JID from the WhatsApp store
 *
 * @param client The whatsmeow client
 * @param jid The JID to get the username for
 * @return The username or empty string if not found
 */
func GetUsernameFromJID(client *whatsmeow.Client, jid types.JID) string {
	if client == nil || client.Store == nil {
		return ""
	}

	// The current whatsmeow ContactInfo store no longer exposes a dedicated
	// username field. Keep the public contract stable by returning empty until
	// the upstream library provides username access again through a supported API.
	if _, err := client.Store.Contacts.GetContact(context.Background(), jid); err == nil {
		return ""
	}

	return ""
}

/**
 * GetPhoneFromJID retrieves the raw phone number for any JID type from the ContactManager's Store
 * This uses ContactManager.GetPhoneFromStore to avoid additional whatsmeow calls
 *
 * @param contactManager The contact manager interface (must be WhatsmeowContactManager)
 * @param jid The JID to get the phone for (works for @lid and @s.whatsapp.net)
 * @return The raw phone number (without E.164 formatting) or empty string if not found
 */
func GetPhoneFromJID(contactManager whatsapp.WhatsappContactManagerInterface, jid types.JID) string {
	// Type assertion to access GetPhoneFromStore method
	if cm, ok := contactManager.(*WhatsmeowContactManager); ok {
		phoneRaw := cm.GetPhoneFromStore(jid)
		if len(phoneRaw) > 0 {
			return library.NormalizeCanonicalPhone(phoneRaw)
		}
	}
	return ""
}

/**
 * GetPhoneE164FromJID retrieves the E.164 formatted phone number for any JID type
 * This uses ContactManager.GetPhoneFromContactId which handles E.164 formatting
 *
 * @param contactManager The contact manager interface
 * @param jid The JID to get the phone for (works for @lid and @s.whatsapp.net)
 * @return The E.164 formatted phone number or empty string if not found
 */
func GetPhoneE164FromJID(contactManager whatsapp.WhatsappContactManagerInterface, jid types.JID) string {
	jidString := jid.User + "@" + jid.Server
	phone, err := contactManager.GetPhoneFromContactId(jidString)
	if err == nil && len(phone) > 0 {
		return phone
	}
	return ""
}

func NewWhatsappChat(handler *WhatsmeowHandlers, jid types.JID) *whatsapp.WhatsappChat {
	contactManager := handler.GetContactManager()
	return NewWhatsappChatRaw(handler.Client, contactManager, jid)
}

func NewWhatsappChatRaw(client *whatsmeow.Client, contactManager whatsapp.WhatsappContactManagerInterface, jid types.JID) *whatsapp.WhatsappChat {
	chat := &whatsapp.WhatsappChat{}

	// Always use User@Server format WITHOUT session ID
	// The types.JID already separates the user from session suffix
	chat.Id = jid.User + "@" + jid.Server

	chat.Title = GetChatTitle(client, jid)

	switch jid.Server {
	case whatsapp.WHATSAPP_SERVERDOMAIN_USER:
		phone, err := contactManager.GetPhoneFromContactId(chat.Id)
		if err == nil && len(phone) > 0 {
			chat.Phone = phone

			// Try to get LID from phone number
			// This will populate LId if available
			lid, err := contactManager.GetLIDFromPhone(chat.Phone)
			if err == nil && len(lid) > 0 {
				chat.LId = lid
			}
		}
	case whatsapp.WHATSAPP_SERVERDOMAIN_LID:
		// For @lid contacts, use the raw phone-number JID (the same jid.User form a
		// native @s.whatsapp.net contact yields) as the chat id, so downstream
		// consumers (webhooks, integrations) key on the stable phone identity and do
		// not create a duplicate contact per @lid. Keep the @lid in LId.
		chat.LId = chat.Id
		if phoneRaw := GetPhoneFromJID(contactManager, jid); len(phoneRaw) > 0 {
			chat.Id = phoneRaw + "@" + whatsapp.WHATSAPP_SERVERDOMAIN_USER
			// Get E.164 formatted phone for display purposes
			if phoneE164 := GetPhoneE164FromJID(contactManager, jid); len(phoneE164) > 0 {
				chat.Phone = phoneE164
			}
		}
		// If phone not found, chat.Id and chat.LId remain equal (both @lid)
	}

	// Try to get username from contact info (newly launched by WhatsApp)
	chat.Username = GetUsernameFromJID(client, jid)

	return chat
}
