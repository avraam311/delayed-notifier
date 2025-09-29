package models

type Notification struct {
	UserID  int
	ToMail  []string
	Message string
}
