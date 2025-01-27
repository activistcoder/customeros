package data_fields

import (
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/utils"
	neo4jenum "github.com/openline-ai/openline-customer-os/packages/server/customer-os-neo4j-repository/enum"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-neo4j-repository/model"
	"time"
)

type ContractSaveFields struct {
	CreatedAt              *time.Time            `json:"createdAt,omitempty"`
	AppSource              *string               `json:"appSource,omitempty"`
	Source                 *string               `json:"source,omitempty"`
	OrganizationId         *string               `json:"organizationId,omitempty"`
	CreatedByUserId        *string               `json:"createdByUserId,omitempty"`
	Name                   *string               `json:"name,omitempty"`
	ContractUrl            *string               `json:"contractUrl,omitempty"`
	ServiceStartedAt       *time.Time            `json:"serviceStartedAt,omitempty"`
	SignedAt               *time.Time            `json:"signedAt,omitempty"`
	EndedAt                *time.Time            `json:"endedAt,omitempty"`
	LengthInMonths         *int64                `json:"lengthInMonths,omitempty"`
	Status                 *string               `json:"status,omitempty"`
	BillingCycleInMonths   *int64                `json:"billingCycleInMonths,omitempty"`
	Currency               *neo4jenum.Currency   `json:"currency,omitempty"`
	InvoicingStartDate     *time.Time            `json:"invoicingStartDate,omitempty"`
	NextInvoiceDate        *time.Time            `json:"nextInvoiceDate,omitempty"`
	InvoicingEnabled       *bool                 `json:"invoicingEnabled,omitempty"`
	PayOnline              *bool                 `json:"payOnline,omitempty"`
	PayAutomatically       *bool                 `json:"payAutomatically,omitempty"`
	CanPayWithCard         *bool                 `json:"canPayWithCard,omitempty"`
	CanPayWithDirectDebit  *bool                 `json:"canPayWithDirectDebit,omitempty"`
	CanPayWithBankTransfer *bool                 `json:"canPayWithBankTransfer,omitempty"`
	AutoRenew              *bool                 `json:"autoRenew,omitempty"`
	Check                  *bool                 `json:"check,omitempty"`
	DueDays                *int64                `json:"dueDays,omitempty"`
	Country                *string               `json:"country,omitempty"`
	Approved               *bool                 `json:"approved,omitempty"`
	AddressLine1           *string               `json:"addressLine1,omitempty"`
	AddressLine2           *string               `json:"addressLine2,omitempty"`
	Locality               *string               `json:"locality,omitempty"`
	Region                 *string               `json:"region,omitempty"`
	Zip                    *string               `json:"zip,omitempty"`
	OrganizationLegalName  *string               `json:"organizationLegalName,omitempty"`
	InvoiceEmail           *string               `json:"invoiceEmail,omitempty"`
	InvoiceEmailCC         *[]string             `json:"invoiceEmailCC,omitempty"`
	InvoiceEmailBCC        *[]string             `json:"invoiceEmailBCC,omitempty"`
	InvoiceNote            *string               `json:"invoiceNote,omitempty"`
	ExternalSystem         *model.ExternalSystem `json:"externalSystem,omitempty"`
}

func (c ContractSaveFields) GetCreatedByUserId() string {
	return utils.IfNotNilString(c.CreatedByUserId)
}

func (c ContractSaveFields) GetOrganizationId() string {
	return utils.IfNotNilString(c.OrganizationId)
}
