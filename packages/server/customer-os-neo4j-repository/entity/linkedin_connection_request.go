package entity

import (
	"time"
)

type LinkedinConnectionRequest struct {
	Id        string
	CreatedAt time.Time
	UpdatedAt time.Time

	ProducerId   string
	ProducerType string

	SocialUrl string
	UserId    string

	ScheduledAt time.Time
	Status      LinkedinConnectionRequestStatus
}

type LinkedinConnectionRequestStatus string

const (
	LinkedinConnectionRequestStatusPending  LinkedinConnectionRequestStatus = "PENDING"
	LinkedinConnectionRequestStatusAccepted LinkedinConnectionRequestStatus = "ACCEPTED"
	LinkedinConnectionRequestStatusDeclined LinkedinConnectionRequestStatus = "DECLINED"
)

func GetLinkedinConnectionRequestStatus(s string) LinkedinConnectionRequestStatus {
	return LinkedinConnectionRequestStatus(s)
}
