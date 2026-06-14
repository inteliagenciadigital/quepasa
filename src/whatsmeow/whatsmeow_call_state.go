package whatsmeow

import (
	"sync"
	"time"

	types "go.mau.fi/whatsmeow/types"
)

// VoIPCallSession represents an active or negotiating call session
type VoIPCallSession struct {
	CallID         string
	Peer           types.JID
	State          string // Offered, Answering, Accepted, Terminated, Rejected
	RemotePlatform string
	RemoteVersion  string
	OfferTime      time.Time
	AnswerTime     time.Time
	Keys           map[string][]byte
	Candidates     []string
}

// VoIPCallSessionManager manages in-memory VoIP sessions in a thread-safe manner
type VoIPCallSessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*VoIPCallSession
}

// GlobalVoIPSessionManager is the system-wide manager for active call sessions
var GlobalVoIPSessionManager = NewVoIPCallSessionManager()

// NewVoIPCallSessionManager initializes a new session manager
func NewVoIPCallSessionManager() *VoIPCallSessionManager {
	return &VoIPCallSessionManager{
		sessions: make(map[string]*VoIPCallSession),
	}
}

// AddSession registers a new call session
func (m *VoIPCallSessionManager) AddSession(session *VoIPCallSession) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.CallID] = session
}

// GetSession retrieves a call session by CallID
func (m *VoIPCallSessionManager) GetSession(callID string) (*VoIPCallSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, exists := m.sessions[callID]
	return session, exists
}

// UpdateState modifies the state of a call session
func (m *VoIPCallSessionManager) UpdateState(callID string, state string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if session, exists := m.sessions[callID]; exists {
		session.State = state
		if state == "Accepted" {
			session.AnswerTime = time.Now()
		}
	}
}

// RemoveSession deletes a session from tracking
func (m *VoIPCallSessionManager) RemoveSession(callID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, callID)
}

// ListSessions returns a list of all active call sessions
func (m *VoIPCallSessionManager) ListSessions() []*VoIPCallSession {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*VoIPCallSession, 0, len(m.sessions))
	for _, session := range m.sessions {
		list = append(list, session)
	}
	return list
}
