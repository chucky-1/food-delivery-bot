package storage

import "fmt"

const (
	CreateOrganization = "create_organization"
	JoinToOrganization = "join"
	AddAddress         = "add_address"
	AddFirstName       = "first_name"
	AddLastName        = "last_name"
	AddMiddleName      = "middle_name"
)

var (
	ErrDoesntExistMessage   = fmt.Errorf("doesn't exist message by user")
	ErrUserWaitOtherMessage = fmt.Errorf("user waits other message")
)

type MessageType struct {
	Action    string
	MessageID int

	// Data if action has 2 step
	DataOnFirstStep string
}

type Messages struct {
	storeByUserID map[int64]*MessageType
}

func NewMessage() *Messages {
	return &Messages{
		storeByUserID: make(map[int64]*MessageType),
	}
}

func (m *Messages) WaitMessage(userID int64, action string, messageID int, data string) {
	m.storeByUserID[userID] = &MessageType{
		Action:          action,
		MessageID:       messageID,
		DataOnFirstStep: data,
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
