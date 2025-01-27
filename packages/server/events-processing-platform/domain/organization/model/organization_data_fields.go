package model

import (
	cmnmod "github.com/openline-ai/openline-customer-os/packages/server/events/event/common"
	"time"
)

type OrganizationDataFields struct {
	Name               string
	Hide               bool
	Description        string
	Website            string
	Industry           string
	SubIndustry        string
	IndustryGroup      string
	TargetAudience     string
	ValueProposition   string
	IsPublic           bool
	Employees          int64
	Market             string
	LastFundingRound   string
	LastFundingAmount  string
	ReferenceId        string
	Note               string
	YearFounded        *int64
	Headquarters       string
	EmployeeGrowthRate string
	LogoUrl            string
	IconUrl            string
	SlackChannelId     string
	Relationship       string
	Stage              string
	LeadSource         string
	IcpFit             bool
}

type OrganizationFields struct {
	ID                     string
	Tenant                 string
	OrganizationDataFields OrganizationDataFields
	CreatedAt              *time.Time
	UpdatedAt              *time.Time
	Source                 cmnmod.Source
	ExternalSystem         cmnmod.ExternalSystem
}
