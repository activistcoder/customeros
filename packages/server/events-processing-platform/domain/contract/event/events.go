package event

const (
	// Deprecated
	ContractCreateV1 = "V1_CONTRACT_CREATE"
	// Deprecated
	ContractUpdateV1                    = "V1_CONTRACT_UPDATE"
	ContractUpdateStatusV1              = "V1_CONTRACT_UPDATE_STATUS"
	ContractRefreshStatusV1             = "V1_CONTRACT_REFRESH_STATUS"
	ContractRefreshLtvV1                = "V1_CONTRACT_REFRESH_LTV"
	ContractRolloutRenewalOpportunityV1 = "V1_CONTRACT_ROLLOUT_RENEWAL_OPPORTUNITY"
	ContractDeleteV1                    = "V1_CONTRACT_DELETE"
)

const (
	FieldMaskName                   = "name"
	FieldMaskContractURL            = "contractURL"
	FieldMaskSignedAt               = "signedAt"
	FieldMaskEndedAt                = "endedAt"
	FieldMaskServiceStartedAt       = "serviceStartedAt"
	FieldMaskInvoicingStartDate     = "invoicingStartDate"
	FieldMaskBillingCycle           = "billingCycle"
	FieldMaskBillingCycleInMonths   = "billingCycleInMonths"
	FieldMaskCurrency               = "currency"
	FieldMaskAddressLine1           = "addressLine1"
	FieldMaskAddressLine2           = "addressLine2"
	FieldMaskZip                    = "zip"
	FieldMaskCountry                = "country"
	FieldMaskRegion                 = "region"
	FieldMaskLocality               = "locality"
	FieldMaskOrganizationLegalName  = "organizationLegalName"
	FieldMaskInvoiceEmail           = "invoiceEmail"
	FieldMaskInvoiceEmailCC         = "invoiceEmailCC"
	FieldMaskInvoiceEmailBCC        = "invoiceEmailBCC"
	FieldMaskStatus                 = "status"
	FieldMaskInvoiceNote            = "invoiceNote"
	FieldMaskNextInvoiceDate        = "nextInvoiceDate"
	FieldMaskCanPayWithCard         = "canPayWithCard"
	FieldMaskCanPayWithDirectDebit  = "canPayWithDirectDebit"
	FieldMaskCanPayWithBankTransfer = "canPayWithBankTransfer"
	FieldMaskPayOnline              = "payOnline"
	FieldMaskPayAutomatically       = "payAutomatically"
	FieldMaskInvoicingEnabled       = "invoicingEnabled"
	FieldMaskAutoRenew              = "autoRenew"
	FieldMaskCheck                  = "check"
	FieldMaskDueDays                = "dueDays"
	FieldMaskLengthInMonths         = "lengthInMonths"
	FieldMaskApproved               = "approved"
)
