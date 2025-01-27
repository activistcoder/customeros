package service

import (
	"context"
	mailsherpa "github.com/customeros/mailsherpa/mailvalidate"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/common"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/constants"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/dto"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/logger"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/model"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/tracing"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/utils"
	neo4jentity "github.com/openline-ai/openline-customer-os/packages/server/customer-os-neo4j-repository/entity"
	neo4jmapper "github.com/openline-ai/openline-customer-os/packages/server/customer-os-neo4j-repository/mapper"
	neo4jmodel "github.com/openline-ai/openline-customer-os/packages/server/customer-os-neo4j-repository/model"
	neo4jrepository "github.com/openline-ai/openline-customer-os/packages/server/customer-os-neo4j-repository/repository"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"time"
)

type ContactService interface {
	SaveContact(ctx context.Context, id *string, contactFields neo4jrepository.ContactFields, socialUrl string, externalSystem neo4jmodel.ExternalSystem) (string, error)
	HideContact(ctx context.Context, contactId string) error
	ShowContact(ctx context.Context, contactId string) error
	GetContactById(ctx context.Context, contactId string) (*neo4jentity.ContactEntity, error)
	LinkContactWithOrganization(ctx context.Context, contactId, organizationId, jobTitle, description, source string, primary bool, startedAt, endedAt *time.Time) error
	CheckContactExistsWithLinkedIn(ctx context.Context, url, alias, externalId string) (bool, string, error)
	CheckContactExistsWithEmail(ctx context.Context, email string) (bool, string, error)
}

type contactService struct {
	log      logger.Logger
	services *Services
}

func NewContactService(log logger.Logger, services *Services) ContactService {
	return &contactService{
		log:      log,
		services: services,
	}
}

func (s *contactService) SaveContact(ctx context.Context, id *string, contactFields neo4jrepository.ContactFields, socialUrl string, externalSystem neo4jmodel.ExternalSystem) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ContactService.SaveContact")
	defer span.Finish()
	tracing.SetDefaultServiceSpanTags(ctx, span)
	tracing.LogObjectAsJson(span, "contactFields", contactFields)
	tracing.LogObjectAsJson(span, "externalSystem", externalSystem)
	span.LogKV("socialUrl", socialUrl)

	// validate tenant
	err := common.ValidateTenant(ctx)
	if err != nil {
		tracing.TraceErr(span, err)
		return "", err
	}
	tenant := common.GetTenantFromContext(ctx)

	// set default values
	if contactFields.SourceFields.AppSource != "" {
		common.SetAppSourceInContext(ctx, contactFields.SourceFields.AppSource)
	}

	createFlow := false
	contactId := ""

	if id == nil || *id == "" {
		createFlow = true

		// Reject contact creation if linked-in url is already used by another contact
		if (neo4jentity.SocialEntity{Url: socialUrl}).IsLinkedin() {
			linkedInUsed, existingContactId, err := s.services.ContactService.CheckContactExistsWithLinkedIn(ctx, socialUrl, "", "")
			if err != nil {
				tracing.TraceErr(span, err)
				return "", err
			}
			if linkedInUsed {
				err = errors.Errorf("linkedin url %s already used by contact %s", socialUrl, existingContactId)
				return "", err
			}
		}

		span.LogKV("flow", "create")
		contactId, err = s.services.Neo4jRepositories.CommonReadRepository.GenerateId(ctx, tenant, model.NodeLabelContact)
		if err != nil {
			tracing.TraceErr(span, err)
			return "", err
		}
		// prepare missing fields
		contactFields.CreatedAt = utils.NowIfZero(contactFields.CreatedAt)
		if contactFields.SourceFields.Source == "" {
			contactFields.SourceFields.Source = neo4jentity.DataSourceOpenline.String()
		}
		if contactFields.SourceFields.AppSource == "" {
			contactFields.SourceFields.AppSource = common.GetAppSourceFromContext(ctx)
		}
	} else {
		span.LogKV("flow", "update")
		contactId = *id

		// validate contact exists
		exists, err := s.services.Neo4jRepositories.CommonReadRepository.ExistsById(ctx, tenant, contactId, model.NodeLabelContact)
		if err != nil || !exists {
			err = errors.New("contact not found")
			tracing.TraceErr(span, err)
			return "", err
		}
	}
	tracing.TagEntity(span, contactId)

	// Clean and update contact names if not updated manually
	if common.GetAppSourceFromContext(ctx) != constants.AppSourceCustomerOsApi {
		if contactFields.UpdateName {
			contactFields.Name = utils.CleanName(contactFields.Name)
		}
		if contactFields.UpdateFirstName {
			contactFields.FirstName = utils.CleanName(contactFields.FirstName)
		}
		if contactFields.UpdateLastName {
			contactFields.LastName = utils.CleanName(contactFields.LastName)
		}
	}

	_, err = utils.ExecuteWriteInTransaction(ctx, s.services.Neo4jRepositories.Neo4jDriver, s.services.Neo4jRepositories.Database, nil, func(tx neo4j.ManagedTransaction) (any, error) {
		innerErr := s.services.Neo4jRepositories.ContactWriteRepository.SaveContactInTx(ctx, &tx, tenant, contactId, contactFields)
		if innerErr != nil {
			s.log.Errorf("Error while saving contact %s: %s", contactId, err.Error())
			return nil, innerErr
		}
		if externalSystem.Available() {
			innerErr = s.services.Neo4jRepositories.ExternalSystemWriteRepository.LinkWithEntityInTx(ctx, tx, tenant, contactId, model.NodeLabelContact, externalSystem)
			if err != nil {
				s.log.Errorf("Error while link contact %s with external system %s: %s", contactId, externalSystem.ExternalSystemId, err.Error())
				return nil, innerErr
			}
		}
		return nil, nil
	})
	if err != nil {
		tracing.TraceErr(span, err)
		return "", err
	}

	if createFlow {
		err = s.services.RabbitMQService.PublishEvent(ctx, contactId, model.CONTACT, dto.New_CreateContact_From_ContactFields(contactFields, externalSystem))
		if err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "unable to publish message CreateContact"))
		}
		utils.EventCompleted(ctx, tenant, model.CONTACT.String(), contactId, s.services.GrpcClients, utils.NewEventCompletedDetails().WithCreate())
	} else {
		err = s.services.RabbitMQService.PublishEvent(ctx, contactId, model.CONTACT, dto.New_UpdateContact_From_ContactFields(contactFields, externalSystem))
		if err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "unable to publish message UpdateContact"))
		}
		if contactFields.SourceFields.AppSource != constants.AppSourceCustomerOsApi {
			utils.EventCompleted(ctx, tenant, model.CONTACT.String(), contactId, s.services.GrpcClients, utils.NewEventCompletedDetails().WithUpdate())
		}
	}

	if createFlow && socialUrl != "" {
		_, err := s.services.SocialService.AddSocialToEntity(ctx,
			LinkWith{
				Id:   contactId,
				Type: model.CONTACT,
			},
			neo4jentity.SocialEntity{
				Url:       socialUrl,
				Source:    neo4jentity.DecodeDataSource(contactFields.SourceFields.Source),
				AppSource: contactFields.SourceFields.AppSource,
			})
		if err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to merge social with contact"))
		}
	}

	if createFlow {
		span.LogFields(log.Bool("response.contactCreated", true))
	} else {
		span.LogFields(log.Bool("response.contactUpdated", true))
	}
	return contactId, nil
}

func (s *contactService) HideContact(ctx context.Context, contactId string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ContactService.HideContact")
	defer span.Finish()
	tracing.SetDefaultServiceSpanTags(ctx, span)
	tracing.TagEntity(span, contactId)

	// validate tenant
	err := common.ValidateTenant(ctx)
	if err != nil {
		tracing.TraceErr(span, err)
		return err
	}
	tenant := common.GetTenantFromContext(ctx)

	err = s.services.Neo4jRepositories.ContactWriteRepository.UpdateAnyProperty(ctx, tenant, contactId, neo4jentity.ContactPropertyHide, true)
	if err != nil {
		tracing.TraceErr(span, err)
		s.log.Errorf("error while hiding contact %s: %s", contactId, err.Error())
	}
	err = s.services.Neo4jRepositories.CommonWriteRepository.UpdateTimeProperty(ctx, tenant, model.NodeLabelContact, contactId, string(neo4jentity.ContactPropertyHiddenAt), utils.NowPtr())
	if err != nil {
		tracing.TraceErr(span, err)
		s.log.Errorf("error while updating hidden at property for contact %s: %s", contactId, err.Error())
	}

	err = s.services.RabbitMQService.PublishEvent(ctx, contactId, model.CONTACT, dto.HideContact{})
	if err != nil {
		tracing.TraceErr(span, errors.Wrap(err, "unable to publish message HideContact"))
	}

	utils.EventCompleted(ctx, tenant, model.CONTACT.String(), contactId, s.services.GrpcClients, utils.NewEventCompletedDetails().WithDelete())

	return nil
}

func (s *contactService) ShowContact(ctx context.Context, contactId string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ContactService.ShowContact")
	defer span.Finish()
	tracing.SetDefaultServiceSpanTags(ctx, span)
	tracing.TagEntity(span, contactId)

	// validate tenant
	err := common.ValidateTenant(ctx)
	if err != nil {
		tracing.TraceErr(span, err)
		return err
	}
	tenant := common.GetTenantFromContext(ctx)

	err = s.services.Neo4jRepositories.ContactWriteRepository.UpdateAnyProperty(ctx, tenant, contactId, neo4jentity.ContactPropertyHide, false)
	if err != nil {
		tracing.TraceErr(span, err)
		s.log.Errorf("error while showing contact %s: %s", contactId, err.Error())
		return err
	}

	err = s.services.RabbitMQService.PublishEvent(ctx, contactId, model.CONTACT, dto.ShowContact{})
	if err != nil {
		tracing.TraceErr(span, errors.Wrap(err, "unable to publish message ShowContact"))
	}

	utils.EventCompleted(ctx, tenant, model.CONTACT.String(), contactId, s.services.GrpcClients, utils.NewEventCompletedDetails().WithCreate())

	return nil
}

func (s *contactService) GetContactById(ctx context.Context, contactId string) (*neo4jentity.ContactEntity, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ContactService.GetContactById")
	defer span.Finish()
	tracing.SetDefaultServiceSpanTags(ctx, span)
	tracing.TagEntity(span, contactId)

	// validate tenant
	err := common.ValidateTenant(ctx)
	if err != nil {
		tracing.TraceErr(span, err)
		return nil, err
	}
	tenant := common.GetTenantFromContext(ctx)

	contactDbNode, err := s.services.Neo4jRepositories.ContactReadRepository.GetContact(ctx, tenant, contactId)
	if err != nil {
		tracing.TraceErr(span, err)
		s.log.Errorf("error while getting contact %s: %s", contactId, err.Error())
		return nil, err
	}

	return neo4jmapper.MapDbNodeToContactEntity(contactDbNode), nil
}

func (s *contactService) LinkContactWithOrganization(ctx context.Context, contactId, organizationId, jobTitle, description, source string, primary bool, startedAt, endedAt *time.Time) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ContactService.LinkContactWithOrganization")
	defer span.Finish()
	tracing.SetDefaultServiceSpanTags(ctx, span)
	tracing.TagEntity(span, contactId)
	span.LogFields(log.String("organizationId", organizationId), log.String("jobTitle", jobTitle), log.String("description", description), log.Bool("primary", primary))
	if startedAt != nil {
		span.LogFields(log.Object("startedAt", startedAt))
	}
	if endedAt != nil {
		span.LogFields(log.Object("endedAt", endedAt))
	}

	// validate tenant
	err := common.ValidateTenant(ctx)
	if err != nil {
		tracing.TraceErr(span, err)
		return err
	}
	tenant := common.GetTenantFromContext(ctx)

	// validate contact exists
	exists, err := s.services.Neo4jRepositories.CommonReadRepository.ExistsById(ctx, tenant, contactId, model.NodeLabelContact)
	if err != nil || !exists {
		err = errors.New("contact not found")
		tracing.TraceErr(span, err)
		return err
	}

	// validate organization exists
	exists, err = s.services.Neo4jRepositories.CommonReadRepository.ExistsById(ctx, tenant, organizationId, model.NodeLabelOrganization)
	if err != nil || !exists {
		err = errors.New("organization not found")
		tracing.TraceErr(span, err)
		return err
	}

	if startedAt != nil && startedAt.Before(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)) {
		startedAt = nil
	}
	if endedAt != nil && endedAt.Before(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)) {
		endedAt = nil
	}

	jobRoleData := neo4jrepository.JobRoleFields{
		Description: description,
		JobTitle:    jobTitle,
		Primary:     primary,
		StartedAt:   startedAt,
		EndedAt:     endedAt,
		SourceFields: neo4jmodel.SourceFields{
			Source:    neo4jmodel.GetSource(source),
			AppSource: common.GetAppSourceFromContext(ctx),
		},
	}

	err = s.services.Neo4jRepositories.JobRoleWriteRepository.LinkContactWithOrganization(ctx, tenant, contactId, organizationId, jobRoleData)
	if err != nil {
		tracing.TraceErr(span, errors.Wrap(err, "unable to link contact with organization"))
		return err
	}

	utils.EventCompleted(ctx, tenant, model.CONTACT.String(), contactId, s.services.GrpcClients, utils.NewEventCompletedDetails().WithUpdate())
	utils.EventCompleted(ctx, tenant, model.ORGANIZATION.String(), organizationId, s.services.GrpcClients, utils.NewEventCompletedDetails().WithUpdate())

	// send 2 events, for contact another for organization
	dtoData := dto.AddContactToOrganization{
		ContactId:      contactId,
		OrganizationId: organizationId,
		JobTitle:       jobTitle,
		Description:    description,
		Primary:        primary,
		StartedAt:      startedAt,
		EndedAt:        endedAt,
	}
	err = s.services.RabbitMQService.PublishEvent(ctx, contactId, model.CONTACT, dtoData)
	if err != nil {
		tracing.TraceErr(span, errors.Wrap(err, "unable to publish message AddContactToOrganization for contact"))
	}
	err = s.services.RabbitMQService.PublishEvent(ctx, organizationId, model.ORGANIZATION, dtoData)
	if err != nil {
		tracing.TraceErr(span, errors.Wrap(err, "unable to publish message AddContactToOrganization for organization"))
	}

	return nil
}

func (s *contactService) CheckContactExistsWithLinkedIn(ctx context.Context, url, alias, externalId string) (bool, string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ContactService.CheckContactExistsWithLinkedIn")
	defer span.Finish()

	if alias == "" {
		// use identifier as alias
		alias = neo4jentity.SocialEntity{Url: url}.ExtractLinkedinPersonIdentifierFromUrl()
	}
	contacts, err := s.services.Neo4jRepositories.ContactReadRepository.GetContactsByLinkedIn(ctx, common.GetTenantFromContext(ctx), url, alias, externalId)
	if err != nil {
		tracing.TraceErr(span, err)
		return false, "", err
	}
	contactId := ""
	if len(contacts) > 0 {
		contactId = contacts[0].Props["id"].(string)
	}
	return len(contacts) > 0, contactId, nil
}

func (s *contactService) CheckContactExistsWithEmail(ctx context.Context, email string) (bool, string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ContactService.CheckContactExistsWithEmail")
	defer span.Finish()

	if email == "" {
		return false, "", nil
	}
	// if not a valid email, return false
	syntaxValidation := mailsherpa.ValidateEmailSyntax(email)
	if !syntaxValidation.IsValid {
		return false, "", nil
	}

	contacts, err := s.services.Neo4jRepositories.ContactReadRepository.GetContactsWithEmail(ctx, common.GetTenantFromContext(ctx), email)
	if err != nil {
		tracing.TraceErr(span, err)
		return false, "", err
	}
	contactId := ""
	if len(contacts) > 0 {
		contactId = contacts[0].Props["id"].(string)
	}
	return len(contacts) > 0, contactId, nil
}
