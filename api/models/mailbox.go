package models

type Mailbox struct {
	Username string `json:"username"`
	Domain   string `json:"domain"`
	Name     string `json:"name"`
}

func (Mailbox) TableName() string {
	return "mailbox"
}
