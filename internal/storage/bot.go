package storage

import "fmt"

const (
	CreateOrganization = "create"
	JoinToOrganization = "join"
	AddAddress         = "address"
	StopDish           = "stop"
	ActivateDish       = "activate"
)

var (
	ErrDoesntExistMessage   = fmt.Errorf("doesn't exist message by user")
	ErrUserWaitOtherMessage = fmt.Errorf("user waits other message")
)

type MessageType struct {
	Action    string
	MessageID int
}

type Messages struct {
	storeByUserID map[int64]*MessageType
}

func NewMessage() *Messages {
	return &Messages{
		storeByUserID: make(map[int64]*MessageType),
	}
}

func (m *Messages) WaitMessage(userID int64, action string, messageID int) {
	m.storeByUserID[userID] = &MessageType{
		Action:    action,
		MessageID: messageID,
	}
}

func (m *Messages) Extract(userID int64) (*MessageType, bool) {
	mt, ok := m.storeByUserID[userID]
	if !ok {
		return nil, false
	}
	delete(m.storeByUserID, userID)
	return mt, true
}
