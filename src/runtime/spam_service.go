package runtime

import (
	"fmt"
	"math/rand"

	"github.com/inteliagenciadigital/quepasa/models"
	"github.com/inteliagenciadigital/quepasa/whatsapp"
)

// GetSpamSession returns the section that should send a /spam request.
// When spam_sections has configured rows, only those rows are eligible and the
// lowest configured priority is respected. When multiple ready sections share
// the same priority, one of them is selected at random. When the table is empty,
// the legacy behavior is preserved by returning the first ready live session.
func GetSpamSession() (*models.QpWhatsappSession, error) {
	db := models.GetDatabase()
	if db == nil || db.SpamSections == nil {
		return GetFirstReadySession()
	}

	sections, err := db.SpamSections.ListAll()
	if err != nil {
		return nil, err
	}

	if len(sections) == 0 {
		return GetFirstReadySession()
	}

	token, err := chooseSpamSectionToken(sections, func(token string) bool {
		session, ok := FindLiveSessionByToken(token)
		return ok && session != nil && session.GetStatus() == whatsapp.Ready
	}, rand.Intn)
	if err != nil {
		return nil, err
	}

	session, ok := FindLiveSessionByToken(token)
	if !ok || session == nil {
		return nil, fmt.Errorf("configured spam section disappeared before send")
	}
	return session, nil
}

func chooseSpamSectionToken(sections []*models.QpSpamSection, isReady func(token string) bool, pick func(int) int) (string, error) {
	if pick == nil {
		pick = rand.Intn
	}

	var readyTokens []string
	currentPriority := 0
	for _, item := range sections {
		if item == nil || !item.Enabled {
			continue
		}

		priority := item.EffectivePriority()
		if len(readyTokens) > 0 && priority != currentPriority {
			break
		}
		if isReady != nil && isReady(item.Token) {
			if len(readyTokens) == 0 {
				currentPriority = priority
			}
			readyTokens = append(readyTokens, item.Token)
		}
	}

	if len(readyTokens) == 0 {
		return "", fmt.Errorf("no configured spam section is ready")
	}
	if len(readyTokens) == 1 {
		return readyTokens[0], nil
	}

	index := pick(len(readyTokens))
	if index < 0 || index >= len(readyTokens) {
		index = 0
	}
	return readyTokens[index], nil
}
