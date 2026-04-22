package types

import "time"

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Message struct {
	Role      Role
	Text      string
	Timestamp time.Time
}
