package model

import "time"

const confirmWindow = 800 * time.Millisecond

func (m *Model) isPendingConfirmationFor(key string) bool {
	if !m.chat.confirmPending || m.chat.confirmKey != key {
		return false
	}
	return time.Since(m.chat.confirmPressTime) <= confirmWindow
}

func (m *Model) setPendingConfirmation(key string) {
	m.chat.confirmPending = true
	m.chat.confirmKey = key
	m.chat.confirmPressTime = time.Now()
}

func (m *Model) clearPendingConfirmation() {
	m.chat.confirmPending = false
	m.chat.confirmKey = ""
	m.chat.confirmPressTime = time.Time{}
}
