package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	commonpb "github.com/openline-ai/openline-customer-os/packages/server/events-processing-proto/gen/proto/go/api/grpc/v1/common"
	issuepb "github.com/openline-ai/openline-customer-os/packages/server/events-processing-proto/gen/proto/go/api/grpc/v1/issue"
	cosModel "github.com/openline-ai/openline-customer-os/packages/server/user-admin-api/model"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
	"time"
)

type SourceData struct {
	Users []struct {
		FirstName       string  `json:"firstName"`
		LastName        string  `json:"lastName"`
		Email           string  `json:"email"`
		ProfilePhotoURL *string `json:"profilePhotoUrl,omitempty"`
	} `json:"users"`
	Contacts []struct {
		FirstName       string  `json:"firstName"`
		LastName        string  `json:"lastName"`
		Email           string  `json:"email"`
		ProfilePhotoURL *string `json:"profilePhotoUrl,omitempty"`
	} `json:"contacts"`
	TenantBillingProfiles []struct {
		LegalName                     string `json:"legalName"`
		Email                         string `json:"email"`
		AddressLine1                  string `json:"addressLine1"`
		Locality                      string `json:"locality"`
		Country                       string `json:"country"`
		Zip                           string `json:"zip"`
		DomesticPaymentsBankInfo      string `json:"domesticPaymentsBankInfo"`
		InternationalPaymentsBankInfo string `json:"internationalPaymentsBankInfo"`
		VatNumber                     string `json:"vatNumber"`
		SendInvoicesFrom              string `json:"sendInvoicesFrom"`
		CanPayWithCard                bool   `json:"canPayWithCard"`
		CanPayWithDirectDebitSEPA     bool   `json:"canPayWithDirectDebitSEPA"`
		CanPayWithDirectDebitACH      bool   `json:"canPayWithDirectDebitACH"`
		CanPayWithDirectDebitBacs     bool   `json:"canPayWithDirectDebitBacs"`
		CanPayWithPigeon              bool   `json:"canPayWithPigeon"`
		CanPayWithBankTransfer        bool   `json:"canPayWithBankTransfer"`
		Check                         bool   `json:"check"`
	} `json:"tenantBillingProfiles"`
	Organizations []struct {
		Id                    string `json:"id"`
		Name                  string `json:"name"`
		Domain                string `json:"domain"`
		OnboardingStatusInput []struct {
			Status   string `json:"status"`
			Comments string `json:"comments"`
		} `json:"onboardingStatusInput"`
		Contracts []struct {
			ContractName            string     `json:"contractName"`
			CommittedPeriodInMonths int64      `json:"committedPeriodInMonths"`
			ContractUrl             string     `json:"contractUrl"`
			ServiceStarted          time.Time  `json:"serviceStarted"`
			ContractSigned          time.Time  `json:"contractSigned"`
			InvoicingStartDate      *time.Time `json:"invoicingStartDate"`
			BillingCycle            string     `json:"billingCycle"`
			Currency                string     `json:"currency"`
			AddressLine1            string     `json:"addressLine1"`
			AddressLine2            string     `json:"addressLine2"`
			Zip                     string     `json:"zip"`
			Locality                string     `json:"locality"`
			Country                 string     `json:"country"`
			OrganizationLegalName   string     `json:"organizationLegalName"`
			InvoiceEmail            string     `json:"invoiceEmail"`
			InvoiceNote             string     `json:"invoiceNote"`
			ServiceLines            []struct {
				Description    string     `json:"description"`
				BillingCycle   string     `json:"billingCycle"`
				Price          int        `json:"price"`
				Quantity       int        `json:"quantity"`
				ServiceStarted *time.Time `json:"serviceStarted"`
				ServiceEnded   *time.Time `json:"serviceEnded,omitempty"`
			} `json:"serviceLines"`
		} `json:"contracts,omitempty"`
		People []struct {
			Email       string `json:"email"`
			JobRole     string `json:"jobRole"`
			Description string `json:"description"`
		} `json:"people"`
		Emails []struct {
			From        string    `json:"from"`
			To          []string  `json:"to"`
			Cc          []string  `json:"cc"`
			Bcc         []string  `json:"bcc"`
			Subject     string    `json:"subject"`
			Body        string    `json:"body"`
			ContentType string    `json:"contentType"`
			Date        time.Time `json:"date"`
		} `json:"emails"`
		Meetings []struct {
			CreatedBy string    `json:"createdBy"`
			Attendees []string  `json:"attendees"`
			Subject   string    `json:"subject"`
			Agenda    string    `json:"agenda"`
			StartedAt time.Time `json:"startedAt"`
			EndedAt   time.Time `json:"endedAt"`
		} `json:"meetings"`
		LogEntries []struct {
			CreatedBy   string    `json:"createdBy"`
			Content     string    `json:"content"`
			ContentType string    `json:"contentType"`
			Date        time.Time `json:"date"`
		} `json:"logEntries"`
		Issues []struct {
			CreatedBy   string    `json:"createdBy"`
			CreatedAt   time.Time `json:"createdAt"`
			Subject     string    `json:"subject"`
			Status      string    `json:"status"`
			Priority    string    `json:"priority"`
			Description string    `json:"description"`
		} `json:"issues"`
		Slack [][]struct {
			CreatedBy string    `json:"createdBy"`
			CreatedAt time.Time `json:"createdAt"`
			Message   string    `json:"message"`
		} `json:"slack"`
		Intercom [][]struct {
			CreatedBy string    `json:"createdBy"`
			CreatedAt time.Time `json:"createdAt"`
			Message   string    `json:"message"`
		} `json:"intercom"`
	} `json:"organizations"`
	MasterPlans []struct {
		Name       string `json:"name"`
		Milestones []struct {
			Name          string   `json:"name"`
			Order         int64    `json:"order"`
			DurationHours int64    `json:"durationHours"`
			Optional      bool     `json:"optional"`
			Items         []string `json:"items"`
		} `json:"milestones"`
	} `json:"masterPlans"`
}

type TenantDataInjector interface {
	InjectTenantData(context context.Context, tenant, username string, sourceData *SourceData) error
}

type tenantDataInjector struct {
	services *Services
}

func NewTenantDataInjector(services *Services) TenantDataInjector {
	return &tenantDataInjector{
		services: services,
	}
}

// match (n:User_LightBlok)--(e:Email) where e.email <> "customerosdemo@gmail.com" detach delete n;
// match (n:Email_LightBlok) where n.email <> "customerosdemo@gmail.com" detach delete n;
// match (n:Contact_LightBlok) detach delete n;
// match (n:JobRole_LightBlok) detach delete n;
// match (n:Organization_LightBlok) detach delete n;
// match (n:InteractionSession_LightBlok) detach delete n;
// match (n:InteractionEvent_LightBlok) detach delete n;
// match (n:Note_LightBlok) detach delete n;
// match (n:Action_LightBlok) detach delete n;
// match (n:Meeting_LightBlok) detach delete n;
// match (n:Issue_LightBlok) detach delete n;
// match (n:LogEntry_LightBlok) detach delete n;
func (t *tenantDataInjector) InjectTenantData(context context.Context, tenant, username string, sourceData *SourceData) error {
	appSource := "user-admin-api"

	var userIds = make([]EmailAddressWithId, len(sourceData.Users))
	var contactIds = make([]EmailAddressWithId, len(sourceData.Contacts))

	//users creation
	for _, user := range sourceData.Users {
		userResponse, err := t.services.CustomerOsClient.GetUserByEmail(tenant, user.Email)
		if err != nil {
			return err
		}
		if userResponse == nil {
			userResponse, err := t.services.CustomerOsClient.CreateUser(&cosModel.UserInput{
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Email: cosModel.EmailInput{
					Email: user.Email,
				},
				AppSource:       &appSource,
				ProfilePhotoURL: user.ProfilePhotoURL,
			}, tenant, []cosModel.Role{cosModel.RoleUser, cosModel.RoleOwner})
			if err != nil {
				return err
			}

			userIds = append(userIds, EmailAddressWithId{
				Email: user.Email,
				Id:    userResponse.ID,
			})
		} else {
			userIds = append(userIds, EmailAddressWithId{
				Email: user.Email,
				Id:    userResponse.ID,
			})
		}
	}

	//contacts creation
	for _, contact := range sourceData.Contacts {
		contactId, err := t.services.CustomerOsClient.CreateContact(tenant, username, contact.FirstName, contact.LastName, contact.Email, contact.ProfilePhotoURL)
		if err != nil {
			return err
		}

		contactIds = append(contactIds, EmailAddressWithId{
			Email: contact.Email,
			Id:    contactId,
		})
	}

	//create tenant billingProfile
	for _, tenantBillingProfile := range sourceData.TenantBillingProfiles {
		tenantBillingProfileInput := cosModel.TenantBillingProfileInput{
			LegalName:                     tenantBillingProfile.LegalName,
			Email:                         tenantBillingProfile.Email,
			AddressLine1:                  tenantBillingProfile.AddressLine1,
			Locality:                      tenantBillingProfile.Locality,
			Country:                       tenantBillingProfile.Country,
			Zip:                           tenantBillingProfile.Zip,
			DomesticPaymentsBankInfo:      tenantBillingProfile.DomesticPaymentsBankInfo,
			InternationalPaymentsBankInfo: tenantBillingProfile.InternationalPaymentsBankInfo,
			VatNumber:                     tenantBillingProfile.VatNumber,
			SendInvoicesFrom:              tenantBillingProfile.SendInvoicesFrom,
			CanPayWithCard:                tenantBillingProfile.CanPayWithCard,
			CanPayWithDirectDebitSEPA:     tenantBillingProfile.CanPayWithDirectDebitSEPA,
			CanPayWithDirectDebitACH:      tenantBillingProfile.CanPayWithDirectDebitACH,
			CanPayWithDirectDebitBacs:     tenantBillingProfile.CanPayWithDirectDebitBacs,
			CanPayWithPigeon:              tenantBillingProfile.CanPayWithPigeon,
			CanPayWithBankTransfer:        tenantBillingProfile.CanPayWithBankTransfer,
		}
		tenantBillingProfileId, err := t.services.CustomerOsClient.CreateTenantBillingProfile(tenant, username, tenantBillingProfileInput)
		if err != nil {
			return err
		}
		if tenantBillingProfileId == "" {
			return errors.New("tenantBillingProfileId is nil")
		}

	}
	//create orgs
	for _, organization := range sourceData.Organizations {

		var organizationId string
		if organization.Id != "" {
			organizationId = organization.Id
		} else {
			var err error
			organizationId, err = t.services.CustomerOsClient.CreateOrganization(tenant, username, cosModel.OrganizationInput{Name: &organization.Name, Domains: []string{organization.Domain}})
			if err != nil {
				return err
			}
			for _, onboardingStatusInput := range organization.OnboardingStatusInput {
				if onboardingStatusInput.Status != "" {
					organizationOnboardingStatus := cosModel.OrganizationUpdateOnboardingStatus{
						OrganizationId: organizationId,
						Status:         onboardingStatusInput.Status,
						Comments:       onboardingStatusInput.Comments,
					}
					_, err := t.services.CustomerOsClient.UpdateOrganizationOnboardingStatus(tenant, username, organizationOnboardingStatus)
					if err != nil {
						return err
					}
				}
			}
		}

		//TODO FIX DATA
		//create Contracts with Service Lines in org
		for _, contract := range organization.Contracts {
			contractInput := cosModel.ContractInput{
				OrganizationId:          organizationId,
				ContractName:            contract.ContractName,
				CommittedPeriodInMonths: contract.CommittedPeriodInMonths,
				ContractUrl:             contract.ContractUrl,
				ServiceStarted:          contract.ServiceStarted,
				ContractSigned:          contract.ContractSigned,
			}
			contractId, err := t.services.CustomerOsClient.CreateContract(tenant, username, contractInput)
			if err != nil {
				return err
			}
			if contractId == "" {
				return errors.New("contractId is nil")
			}

			waitForContractToExist(t.services, tenant, contractId)

			contractUpdateInput := cosModel.ContractUpdateInput{
				ContractId:            contractId,
				Patch:                 true,
				InvoicingStartDate:    contract.InvoicingStartDate,
				BillingCycle:          contract.BillingCycle,
				Currency:              contract.Currency,
				AddressLine1:          contract.AddressLine1,
				AddressLine2:          contract.AddressLine2,
				Zip:                   contract.Zip,
				Locality:              contract.Locality,
				Country:               contract.Country,
				OrganizationLegalName: contract.OrganizationLegalName,
				InvoiceEmail:          contract.InvoiceEmail,
				InvoiceNote:           contract.InvoiceNote,
			}
			contractId, err = t.services.CustomerOsClient.UpdateContract(tenant, username, contractUpdateInput)
			if err != nil {
				return err
			}
			if contractId == "" {
				return errors.New("contractId is nil")
			}

			for _, serviceLine := range contract.ServiceLines {

				serviceLineInput := func() interface{} {
					if serviceLine.ServiceEnded == nil {
						return cosModel.ServiceLineInput{
							ContractId:     contractId,
							Description:    serviceLine.Description,
							BillingCycle:   serviceLine.BillingCycle,
							Price:          serviceLine.Price,
							Quantity:       serviceLine.Quantity,
							ServiceStarted: serviceLine.ServiceStarted,
						}
					}
					return cosModel.ServiceLineEndedInput{
						ContractId:     contractId,
						Description:    serviceLine.Description,
						BillingCycle:   serviceLine.BillingCycle,
						Price:          serviceLine.Price,
						Quantity:       serviceLine.Quantity,
						ServiceStarted: serviceLine.ServiceStarted,
						ServiceEnded:   serviceLine.ServiceEnded,
					}
				}()
				serviceLineId, err := t.services.CustomerOsClient.CreateServiceLine(tenant, username, serviceLineInput)
				if err != nil {
					return err
				}

				if serviceLineId == "" {
					return errors.New("serviceLineId is nil")
				}

				waitForServiceLineToExist(t.services, contractId, serviceLineId)
			}

			invoiceId, err := t.services.CustomerOsClient.DryRunNextInvoiceForContractInput(tenant, username, contractId)
			if err != nil {
				return err
			}
			if invoiceId == "" {
				return errors.New("invoiceId is nil")
			}
		}

		//create people in org
		for _, people := range organization.People {
			var contactId string
			for _, contact := range contactIds {
				if contact.Email == people.Email {
					contactId = contact.Id
					break
				}
			}

			if contactId == "" {
				return errors.New("contactId is nil")
			}

			err := t.services.CustomerOsClient.AddContactToOrganization(tenant, username, contactId, organizationId, people.JobRole, people.Description)
			if err != nil {
				return err
			}
		}

		//create emails
		for _, email := range organization.Emails {
			sig, _ := uuid.NewUUID()
			sigs := sig.String()

			channelValue := "EMAIL"
			appSource := appSource
			sessionStatus := "ACTIVE"
			sessionType := "THREAD"
			sessionOpts := []InteractionSessionBuilderOption{
				WithSessionIdentifier(&sigs),
				WithSessionChannel(&channelValue),
				WithSessionName(&email.Subject),
				WithSessionAppSource(&appSource),
				WithSessionStatus(&sessionStatus),
				WithSessionType(&sessionType),
			}

			sessionId, err := t.services.CustomerOsClient.CreateInteractionSession(tenant, username, sessionOpts...)
			if sessionId == nil {
				return errors.New("sessionId is nil")
			}

			participantTypeTO, participantTypeCC, participantTypeBCC := "TO", "CC", "BCC"
			participantsTO := toParticipantInputArr(email.To, &participantTypeTO)
			participantsCC := toParticipantInputArr(email.Cc, &participantTypeCC)
			participantsBCC := toParticipantInputArr(email.Bcc, &participantTypeBCC)
			sentTo := append(append(participantsTO, participantsCC...), participantsBCC...)
			sentBy := toParticipantInputArr([]string{email.From}, nil)

			emailChannelData, err := buildEmailChannelData(email.Subject, err)
			if err != nil {
				return err
			}

			iig, err := uuid.NewUUID()
			if err != nil {
				return err
			}
			iigs := iig.String()
			eventOpts := []InteractionEventBuilderOption{
				WithCreatedAt(&email.Date),
				WithSessionId(sessionId),
				WithEventIdentifier(iigs),
				WithChannel(&channelValue),
				WithChannelData(emailChannelData),
				WithContent(&email.Body),
				WithContentType(&email.ContentType),
				WithSentBy(sentBy),
				WithSentTo(sentTo),
				WithAppSource(&appSource),
			}

			interactionEventId, err := t.services.CustomerOsClient.CreateInteractionEvent(tenant, username, eventOpts...)
			if err != nil {
				return err
			}

			if interactionEventId == nil {
				return errors.New("interactionEventId is nil")
			}
		}

		//TODO FIX DATA
		//create meetings
		for _, meeting := range organization.Meetings {
			var createdBy []*cosModel.MeetingParticipantInput
			createdBy = append(createdBy, getMeetingParticipantInput(meeting.CreatedBy, userIds, contactIds))

			var attendedBy []*cosModel.MeetingParticipantInput
			for _, attendee := range meeting.Attendees {
				attendedBy = append(attendedBy, getMeetingParticipantInput(attendee, userIds, contactIds))
			}

			contentType := "text/plain"
			noteInput := cosModel.NoteInput{Content: &meeting.Agenda, ContentType: &contentType, AppSource: &appSource}
			input := cosModel.MeetingInput{
				Name:       &meeting.Subject,
				CreatedAt:  &meeting.StartedAt,
				CreatedBy:  createdBy,
				AttendedBy: attendedBy,
				StartedAt:  &meeting.StartedAt,
				EndedAt:    &meeting.EndedAt,
				Note:       &noteInput,
				AppSource:  &appSource,
			}
			meetingId, err := t.services.CustomerOsClient.CreateMeeting(tenant, username, input)
			if err != nil {
				return err
			}

			if meetingId == "" {
				return errors.New("meetingId is nil")
			}

			eventType := "meeting"
			eventOpts := []InteractionEventBuilderOption{
				WithSentBy([]cosModel.InteractionEventParticipantInput{*getInteractionEventParticipantInput(meeting.CreatedBy, userIds, contactIds)}),
				WithSentTo(getInteractionEventParticipantInputList(meeting.Attendees, userIds, contactIds)),
				WithMeetingId(&meetingId),
				WithCreatedAt(&meeting.StartedAt),
				WithEventType(&eventType),
				WithAppSource(&appSource),
			}

			interactionEventId, err := t.services.CustomerOsClient.CreateInteractionEvent(tenant, username, eventOpts...)
			if err != nil {
				return err
			}

			if interactionEventId == nil {
				return errors.New("interactionEventId is nil")
			}
		}

		//log entries
		for _, logEntry := range organization.LogEntries {

			interactionEventId, err := t.services.CustomerOsClient.CreateLogEntry(tenant, username, organizationId, logEntry.CreatedBy, logEntry.Content, logEntry.ContentType, logEntry.Date)
			if err != nil {
				return err
			}

			if interactionEventId == nil {
				return errors.New("interactionEventId is nil")
			}
		}

		//issues
		for index, issue := range organization.Issues {
			issueGrpcRequest := issuepb.UpsertIssueGrpcRequest{
				Tenant:      tenant,
				Subject:     issue.Subject,
				Status:      issue.Status,
				Priority:    issue.Priority,
				Description: issue.Description,
				CreatedAt:   timestamppb.New(issue.CreatedAt),
				UpdatedAt:   timestamppb.New(issue.CreatedAt),
				SourceFields: &commonpb.SourceFields{
					Source:    "zendesk_support",
					AppSource: appSource,
				},
				ExternalSystemFields: &commonpb.ExternalSystemFields{
					ExternalSystemId: "zendesk_support",
					ExternalId:       "random-thing-" + fmt.Sprintf("%d", index),
					ExternalUrl:      "https://random-thing.zendesk.com/agent/tickets/" + fmt.Sprintf("%d", index),
					SyncDate:         timestamppb.New(issue.CreatedAt),
				},
			}

			issueGrpcRequest.ReportedByOrganizationId = &organizationId

			for _, userWithId := range userIds {
				if userWithId.Email == issue.CreatedBy {
					issueGrpcRequest.SubmittedByUserId = &userWithId.Id
					break
				}
			}

			_, err := t.services.GrpcClients.IssueClient.UpsertIssue(context, &issueGrpcRequest)
			if err != nil {
				return err
			}
		}

		//TODO FIX DATA
		//slack
		//for _, slackThread := range organization.Slack {
		//
		//	sig, err := uuid.NewUUID()
		//	if err != nil {
		//		return err
		//	}
		//	sigs := sig.String()
		//
		//	channelValue := "CHAT"
		//	appSource := appSource
		//	sessionStatus := "ACTIVE"
		//	sessionType := "THREAD"
		//	sessionName := slackThread[0].Message
		//	sessionOpts := []InteractionSessionBuilderOption{
		//		WithSessionIdentifier(&sigs),
		//		WithSessionChannel(&channelValue),
		//		WithSessionName(&sessionName),
		//		WithSessionAppSource(&appSource),
		//		WithSessionStatus(&sessionStatus),
		//		WithSessionType(&sessionType),
		//	}
		//
		//	sessionId, err := t.services.CustomerOsClient.CreateInteractionSession(tenant, username, sessionOpts...)
		//	if sessionId == nil {
		//		return errors.New("sessionId is nil")
		//	}
		//
		//	for _, slackMessage := range slackThread {
		//
		//		sentBy := toParticipantInputArr([]string{slackMessage.CreatedBy}, nil)
		//
		//		iig, err := uuid.NewUUID()
		//		if err != nil {
		//			return err
		//		}
		//		iigs := iig.String()
		//		eventType := "MESSAGE"
		//		contentType := "text/plain"
		//		eventOpts := []InteractionEventBuilderOption{
		//			WithCreatedAt(&slackMessage.CreatedAt),
		//			WithSessionId(sessionId),
		//			WithEventIdentifier(iigs),
		//			WithExternalId(iigs),
		//			WithExternalSystemId("slack"),
		//			WithChannel(&channelValue),
		//			WithEventType(&eventType),
		//			WithContent(&slackMessage.Message),
		//			WithContentType(&contentType),
		//			WithSentBy(sentBy),
		//			WithAppSource(&appSource),
		//		}
		//
		//		interactionEventId, err := t.services.CustomerOsClient.CreateInteractionEvent(tenant, username, eventOpts...)
		//		if err != nil {
		//			return err
		//		}
		//
		//		if interactionEventId == nil {
		//			return errors.New("interactionEventId is nil")
		//		}
		//
		//	}
		//
		//}

		//intercom
		for _, intercomThread := range organization.Intercom {

			sig, err := uuid.NewUUID()
			if err != nil {
				return err
			}
			sigs := sig.String()

			channelValue := "CHAT"
			appSource := appSource
			sessionStatus := "ACTIVE"
			sessionType := "THREAD"
			sessionName := intercomThread[0].Message
			sessionOpts := []InteractionSessionBuilderOption{
				WithSessionIdentifier(&sigs),
				WithSessionChannel(&channelValue),
				WithSessionName(&sessionName),
				WithSessionAppSource(&appSource),
				WithSessionStatus(&sessionStatus),
				WithSessionType(&sessionType),
			}

			sessionId, err := t.services.CustomerOsClient.CreateInteractionSession(tenant, username, sessionOpts...)
			if sessionId == nil {
				return errors.New("sessionId is nil")
			}

			//TODO FIX DATA
			//for _, intercomMessage := range intercomThread {

			//sentById := ""
			//for _, contactWithId := range contactIds {
			//	if contactWithId.Email == intercomMessage.CreatedBy {
			//		sentById = contactWithId.Id
			//		break
			//	}
			//}
			//sentBy := toContactParticipantInputArr([]string{sentById})
			//
			//iig, err := uuid.NewUUID()
			//if err != nil {
			//	return err
			//}
			//iigs := iig.String()
			//eventType := "MESSAGE"
			//contentType := "text/html"
			//eventOpts := []InteractionEventBuilderOption{
			//	WithCreatedAt(&intercomMessage.CreatedAt),
			//	WithSessionId(sessionId),
			//	WithEventIdentifier(iigs),
			//	WithExternalId(iigs),
			//	WithExternalSystemId("intercom"),
			//	WithChannel(&channelValue),
			//	WithEventType(&eventType),
			//	WithContent(&intercomMessage.Message),
			//	WithContentType(&contentType),
			//	WithSentBy(sentBy),
			//	WithAppSource(&appSource),
			//}
			//interactionEventId, err := t.services.CustomerOsClient.CreateInteractionEvent(tenant, username, eventOpts...)
			//if err != nil {
			//	return err
			//}
			//
			//if interactionEventId == nil {
			//	return errors.New("interactionEventId is nil")
			//}

			//}

		}
	}

	for _, masterPlan := range sourceData.MasterPlans {
		masterPlanId, err := t.services.CustomerOsClient.CreateMasterPlan(tenant, username, masterPlan.Name)
		if err != nil {
			return err
		}
		if masterPlanId == "" {
			return errors.New("masterPlanId is nil")
		}
		for _, milestone := range masterPlan.Milestones {
			masterPlanMilestoneInput := cosModel.MasterPlanMilestoneInput{
				MasterPlanId:  masterPlanId,
				Name:          milestone.Name,
				Order:         milestone.Order,
				DurationHours: milestone.DurationHours,
				Optional:      milestone.Optional,
				Items:         milestone.Items,
			}
			masterPlanMilestoneId, err := t.services.CustomerOsClient.CreateMasterPlanMilestone(tenant, username, masterPlanMilestoneInput)
			if err != nil {
				return err
			}
			if masterPlanMilestoneId == "" {
				return errors.New("masterPlanMilestoneId is nil")
			}
		}
	}

	return nil
}

func waitForContractToExist(services *Services, tenant string, contractId string) {
	var _ *dbtype.Node
	var maxAttempts = 5
	var attempt = 0

	for attempt < maxAttempts {
		var err error
		_, err = services.CustomerOsClient.GetContractById(tenant, contractId)
		if err != nil {
			attempt++
			time.Sleep(time.Second * 2)
			if attempt == maxAttempts {
				fmt.Println("Failed to create contracts.")
				os.Exit(1)
			}
		} else {
			break
		}
	}
}

func waitForServiceLineToExist(services *Services, contractId, serviceLineId string) {
	var _ *dbtype.Node
	var maxAttempts = 5
	var attempt = 0

	for attempt < maxAttempts {
		var err error
		_, err = services.CustomerOsClient.GetServiceLine(contractId, serviceLineId)
		if err != nil {
			attempt++
			time.Sleep(time.Second * 2)
			if attempt == maxAttempts {
				fmt.Println("Failed to create service line.")
				os.Exit(1)
			}
		} else {
			break
		}
	}
}

func toParticipantInputArr(from []string, participantType *string) []cosModel.InteractionEventParticipantInput {
	var to []cosModel.InteractionEventParticipantInput
	for _, a := range from {
		participantInput := cosModel.InteractionEventParticipantInput{
			Email: &a,
			Type:  participantType,
		}
		to = append(to, participantInput)
	}
	return to
}

func toContactParticipantInputArr(from []string) []cosModel.InteractionEventParticipantInput {
	var to []cosModel.InteractionEventParticipantInput
	for _, a := range from {
		participantInput := cosModel.InteractionEventParticipantInput{
			ContactID: &a,
		}
		to = append(to, participantInput)
	}
	return to
}

func buildEmailChannelData(subject string, err error) (*string, error) {
	emailContent := cosModel.EmailChannelData{
		Subject: subject,
		//InReplyTo: utils.EnsureEmailRfcIds(email.InReplyTo),
		//Reference: utils.EnsureEmailRfcIds(email.References),
	}
	jsonContent, err := json.Marshal(emailContent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal email content: %v", err)
	}
	jsonContentString := string(jsonContent)

	return &jsonContentString, nil
}

func getMeetingParticipantInput(emailAddress string, userIds, contactIds []EmailAddressWithId) *cosModel.MeetingParticipantInput {
	for _, userWithId := range userIds {
		if userWithId.Email == emailAddress {
			return &cosModel.MeetingParticipantInput{UserID: &userWithId.Id}
		}
	}

	for _, contactWithId := range contactIds {
		if contactWithId.Email == emailAddress {
			return &cosModel.MeetingParticipantInput{ContactID: &contactWithId.Id}
		}
	}

	return nil
}

func getInteractionEventParticipantInput(emailAddress string, userIds, contactIds []EmailAddressWithId) *cosModel.InteractionEventParticipantInput {
	for _, userWithId := range userIds {
		if userWithId.Email == emailAddress {
			return &cosModel.InteractionEventParticipantInput{UserID: &userWithId.Id}
		}
	}

	for _, contactWithId := range contactIds {
		if contactWithId.Email == emailAddress {
			return &cosModel.InteractionEventParticipantInput{ContactID: &contactWithId.Id}
		}
	}

	return nil
}

func getInteractionEventParticipantInputList(emailAddresses []string, userIds, contactIds []EmailAddressWithId) []cosModel.InteractionEventParticipantInput {
	var interactionEventParticipantInputList []cosModel.InteractionEventParticipantInput
	for _, emailAddress := range emailAddresses {
		interactionEventParticipantInputList = append(interactionEventParticipantInputList, *getInteractionEventParticipantInput(emailAddress, userIds, contactIds))
	}
	return interactionEventParticipantInputList
}

type EmailAddressWithId struct {
	Email string
	Id    string
}