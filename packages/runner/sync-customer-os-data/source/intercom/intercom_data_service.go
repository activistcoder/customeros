package intercom

import (
	"context"
	"github.com/openline-ai/openline-customer-os/packages/runner/sync-customer-os-data/common"
	"github.com/openline-ai/openline-customer-os/packages/runner/sync-customer-os-data/config"
	"github.com/openline-ai/openline-customer-os/packages/runner/sync-customer-os-data/entity"
	"github.com/openline-ai/openline-customer-os/packages/runner/sync-customer-os-data/logger"
	"github.com/openline-ai/openline-customer-os/packages/runner/sync-customer-os-data/source"
	source_entity "github.com/openline-ai/openline-customer-os/packages/runner/sync-customer-os-data/source/entity"
	"github.com/openline-ai/openline-customer-os/packages/runner/sync-customer-os-data/source/repository"
	"github.com/openline-ai/openline-customer-os/packages/runner/sync-customer-os-data/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"gorm.io/gorm"
)

const (
	AdminsTableSuffix            = "admins"
	TeamsTableSuffix             = "teams"
	ContactsTableSuffix          = "contacts"
	ConversationsTableSuffix     = "conversations"
	ConversationPartsTableSuffix = "conversation_parts"
)

var sourceTableSuffixByDataType = map[string][]string{
	string(common.USERS):              {AdminsTableSuffix, TeamsTableSuffix},
	string(common.ORGANIZATIONS):      {ContactsTableSuffix},
	string(common.CONTACTS):           {ContactsTableSuffix},
	string(common.INTERACTION_EVENTS): {ConversationsTableSuffix, ConversationPartsTableSuffix},
}

type intercomDataService struct {
	airbyteStoreDb *config.RawDataStoreDB
	tenant         string
	instance       string
	processingIds  map[string]source.ProcessingEntity
	dataFuncs      map[common.SyncedEntityType]func(context.Context, int, string) []any
	log            logger.Logger
}

func NewIntercomDataService(airbyteStoreDb *config.RawDataStoreDB, tenant string, log logger.Logger) source.SourceDataService {
	dataService := intercomDataService{
		airbyteStoreDb: airbyteStoreDb,
		tenant:         tenant,
		processingIds:  map[string]source.ProcessingEntity{},
		log:            log,
	}
	dataService.dataFuncs = map[common.SyncedEntityType]func(context.Context, int, string) []any{}
	dataService.dataFuncs[common.USERS] = dataService.GetUsersForSync
	dataService.dataFuncs[common.ORGANIZATIONS] = dataService.GetOrganizationsForSync
	dataService.dataFuncs[common.CONTACTS] = dataService.GetContactsForSync
	dataService.dataFuncs[common.INTERACTION_EVENTS] = dataService.GetInteractionEventsForSync
	return &dataService
}

func (s *intercomDataService) GetDataForSync(ctx context.Context, dataType common.SyncedEntityType, batchSize int, runId string) []interface{} {
	span, ctx := opentracing.StartSpanFromContext(ctx, "IntercomDataService.GetDataForSync")
	defer span.Finish()
	tracing.SetDefaultSyncServiceSpanTags(ctx, span)
	span.LogFields(log.String("dataType", string(dataType)), log.Int("batchSize", batchSize))

	if ok := s.dataFuncs[dataType]; ok != nil {
		return s.dataFuncs[dataType](ctx, batchSize, runId)
	} else {
		s.log.Infof("No %s data function for %s", s.SourceId(), dataType)
		return nil
	}
}

func (s *intercomDataService) Init() {
	err := s.getDb().AutoMigrate(&source_entity.SyncStatusForAirbyte{})
	if err != nil {
		s.log.Error(err)
	}
}

func (s *intercomDataService) getDb() *gorm.DB {
	//schemaName := s.SourceId()
	//
	//if len(s.instance) > 0 {
	//	schemaName = schemaName + "_" + s.instance
	//}
	//schemaName = schemaName + "_" + s.tenant
	return s.airbyteStoreDb.GetDBHandler(&config.Context{
		Schema: "airbyte_internal",
	})
}

func (s *intercomDataService) SourceId() string {
	return string(entity.AirbyteSourceIntercom)
}

func (s *intercomDataService) Close() {
	s.processingIds = make(map[string]source.ProcessingEntity)
}

func (s *intercomDataService) GetUsersForSync(ctx context.Context, batchSize int, runId string) []any {
	s.processingIds = make(map[string]source.ProcessingEntity)
	currentEntity := string(common.USERS)
	var users []any
	for _, sourceTableSuffix := range sourceTableSuffixByDataType[currentEntity] {
		airbyteRecords, err := repository.GetAirbyteUnprocessedRawRecords(ctx, s.getDb(), batchSize, runId, currentEntity, sourceTableSuffix, s.tenant, s.SourceId())
		if err != nil {
			s.log.Error(err)
			return nil
		}
		for _, v := range airbyteRecords {
			if len(users) >= batchSize {
				break
			}
			outputJSON, err := MapUser(v.AirbyteData)
			user, err := source.MapJsonToUser(outputJSON, v.AirbyteRawId, s.SourceId())
			if err != nil {
				user = entity.UserData{
					BaseData: entity.BaseData{
						SyncId: v.AirbyteRawId,
					},
				}
			}
			s.processingIds[v.AirbyteRawId] = source.ProcessingEntity{
				ExternalId:  user.ExternalId,
				Entity:      currentEntity,
				TableSuffix: sourceTableSuffix,
			}
			users = append(users, user)
		}
	}
	return users
}

func (s *intercomDataService) GetOrganizationsForSync(ctx context.Context, batchSize int, runId string) []any {
	s.processingIds = make(map[string]source.ProcessingEntity)
	currentEntity := string(common.ORGANIZATIONS)

	var organizations []any
	for _, sourceTableSuffix := range sourceTableSuffixByDataType[currentEntity] {
		airbyteRecords, err := repository.GetAirbyteUnprocessedRawRecords(ctx, s.getDb(), batchSize, runId, currentEntity, sourceTableSuffix, s.tenant, s.SourceId())
		if err != nil {
			s.log.Error(err)
			return nil
		}
		for _, v := range airbyteRecords {
			if len(organizations) >= batchSize {
				break
			}
			outputJSON, err := MapOrganization(v.AirbyteData)
			organization, err := source.MapJsonToOrganization(outputJSON, v.AirbyteRawId, s.SourceId())
			if err != nil {
				organization = entity.OrganizationData{
					BaseData: entity.BaseData{
						SyncId: v.AirbyteRawId,
					},
				}
			}

			s.processingIds[v.AirbyteRawId] = source.ProcessingEntity{
				ExternalId:  organization.ExternalId,
				Entity:      currentEntity,
				TableSuffix: sourceTableSuffix,
			}
			organizations = append(organizations, organization)
		}
	}
	return organizations
}

func (s *intercomDataService) GetContactsForSync(ctx context.Context, batchSize int, runId string) []any {
	s.processingIds = make(map[string]source.ProcessingEntity)
	currentEntity := string(common.CONTACTS)

	var contacts []any
	for _, sourceTableSuffix := range sourceTableSuffixByDataType[currentEntity] {
		airbyteRecords, err := repository.GetAirbyteUnprocessedRawRecords(ctx, s.getDb(), batchSize, runId, currentEntity, sourceTableSuffix, s.tenant, s.SourceId())
		if err != nil {
			s.log.Error(err)
			return nil
		}
		for _, v := range airbyteRecords {
			if len(contacts) >= batchSize {
				break
			}
			outputJSON, err := MapContact(v.AirbyteData)
			contact, err := source.MapJsonToContact(outputJSON, v.AirbyteRawId, s.SourceId())
			if err != nil {
				contact = entity.ContactData{
					BaseData: entity.BaseData{
						SyncId: v.AirbyteRawId,
					},
				}
			}

			s.processingIds[v.AirbyteRawId] = source.ProcessingEntity{
				ExternalId:  contact.ExternalId,
				Entity:      currentEntity,
				TableSuffix: sourceTableSuffix,
			}
			contacts = append(contacts, contact)
		}
	}
	return contacts
}

func (s *intercomDataService) GetInteractionEventsForSync(ctx context.Context, batchSize int, runId string) []any {
	s.processingIds = make(map[string]source.ProcessingEntity)
	currentEntity := string(common.INTERACTION_EVENTS)

	var interactionEvents []any
	for _, sourceTableSuffix := range sourceTableSuffixByDataType[currentEntity] {
		airbyteRecords, err := repository.GetAirbyteUnprocessedRawRecords(ctx, s.getDb(), batchSize, runId, currentEntity, sourceTableSuffix, s.tenant, s.SourceId())
		if err != nil {
			s.log.Error(err)
			return nil
		}
		for _, v := range airbyteRecords {
			if len(interactionEvents) >= batchSize {
				break
			}
			outputJSON, err := MapInteractionEvent(v.AirbyteData)
			interactionEvent, err := source.MapJsonToInteractionEvent(outputJSON, v.AirbyteRawId, s.SourceId())
			if err != nil {
				interactionEvent = entity.InteractionEventData{
					BaseData: entity.BaseData{
						SyncId: v.AirbyteRawId,
					},
				}
			}

			s.processingIds[v.AirbyteRawId] = source.ProcessingEntity{
				ExternalId:  interactionEvent.ExternalId,
				Entity:      currentEntity,
				TableSuffix: sourceTableSuffix,
			}
			interactionEvents = append(interactionEvents, interactionEvent)
		}
	}
	return interactionEvents
}

func (s *intercomDataService) MarkProcessed(ctx context.Context, syncId, runId string, synced, skipped bool, reason string) error {
	v, ok := s.processingIds[syncId]
	if ok {
		err := repository.MarkAirbyteRawRecordProcessed(ctx, s.getDb(), s.tenant, v.Entity, v.TableSuffix, syncId, synced, skipped, runId, v.ExternalId, reason)
		if err != nil {
			s.log.Errorf("error while marking %s with external reference %s as synced for %s", v.Entity, v.ExternalId, s.SourceId())
		}
		return err
	}
	return nil
}
