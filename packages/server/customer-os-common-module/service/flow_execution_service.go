package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/common"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/model"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/tracing"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/utils"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-neo4j-repository/entity"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-neo4j-repository/mapper"
	"github.com/opentracing/opentracing-go"
	"time"
)

const numberOfEmailsPerDay = 2
const minMinutesBetweenEmails = int(5)
const maxMinutesBetweenEmails = int(10)

type FlowExecutionService interface {
	GetFlowActionExecutions(ctx context.Context, flowId, contactId string) ([]*entity.FlowActionExecutionEntity, error)

	ScheduleFlow(ctx context.Context, tx *neo4j.ManagedTransaction, flowId, contactId string) error
}

type flowExecutionService struct {
	services *Services
}

func NewFlowExecutionService(services *Services) FlowExecutionService {
	return &flowExecutionService{
		services: services,
	}
}

func (s *flowExecutionService) GetFlowActionExecutions(ctx context.Context, flowId, entityId string) ([]*entity.FlowActionExecutionEntity, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "FlowExecutionService.GetForContact")
	defer span.Finish()
	tracing.SetDefaultServiceSpanTags(ctx, span)

	//get executions for contact
	nodes, err := s.services.Neo4jRepositories.FlowActionExecutionReadRepository.GetByContact(ctx, flowId, entityId)
	if err != nil {
		tracing.TraceErr(span, err)
		return nil, err
	}

	entities := make([]*entity.FlowActionExecutionEntity, 0)
	for _, node := range nodes {
		entities = append(entities, mapper.MapDbNodeToFlowActionExecutionEntity(node))
	}

	return entities, nil
}

func (s *flowExecutionService) ScheduleFlow(ctx context.Context, tx *neo4j.ManagedTransaction, flowId, entityId string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "FlowExecutionService.ScheduleFlow")
	defer span.Finish()
	tracing.SetDefaultServiceSpanTags(ctx, span)

	now := utils.Now()

	flowExecutions, err := s.GetFlowActionExecutions(ctx, flowId, entityId)
	if err != nil {
		tracing.TraceErr(span, err)
		return err
	}

	if len(flowExecutions) == 0 {
		startAction, err := s.services.FlowService.FlowActionGetStart(ctx, flowId)
		if err != nil {
			tracing.TraceErr(span, err)
			return err
		}

		nextActions, err := s.services.FlowService.FlowActionGetNext(ctx, startAction.Id)
		if err != nil {
			tracing.TraceErr(span, err)
			return err
		}

		for _, nextAction := range nextActions {

			scheduleAt := now.Add(time.Duration(nextAction.Data.WaitBefore) * time.Minute)

			err := s.scheduleNextAction(ctx, tx, flowId, entityId, scheduleAt, *nextAction)
			if err != nil {
				tracing.TraceErr(span, err)
				return err
			}
		}

	} else {
		lastActionExecution := flowExecutions[len(flowExecutions)-1]
		lastActionExecutedAt := *lastActionExecution.ExecutedAt

		lastAction, err := s.services.FlowService.FlowActionGetById(ctx, lastActionExecution.ActionId)
		if err != nil {
			tracing.TraceErr(span, err)
			return err
		}

		nextActions, err := s.services.FlowService.FlowActionGetNext(ctx, lastAction.Id)
		if err != nil {
			tracing.TraceErr(span, err)
			return err
		}

		for _, nextAction := range nextActions {

			scheduleAt := lastActionExecutedAt.Add(time.Duration(nextAction.Data.WaitBefore) * time.Minute)

			err := s.scheduleNextAction(ctx, tx, flowId, entityId, scheduleAt, *nextAction)
			if err != nil {
				tracing.TraceErr(span, err)
				return err
			}
		}
	}

	return nil
}

func (s *flowExecutionService) scheduleNextAction(ctx context.Context, tx *neo4j.ManagedTransaction, flowId, entityId string, scheduleAt time.Time, nextAction entity.FlowActionEntity) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "FlowExecutionService.scheduleNextAction")
	defer span.Finish()
	tracing.SetDefaultServiceSpanTags(ctx, span)

	switch nextAction.Data.Action {
	case entity.FlowActionTypeEmailNew, entity.FlowActionTypeEmailReply:
		return s.ScheduleEmailAction(ctx, tx, flowId, entityId, scheduleAt, nextAction)
	default:
		tracing.TraceErr(span, fmt.Errorf("Unsupported action type %s", nextAction.Data.Action))
		return errors.New("Unsupported action type")
	}
}

func (s *flowExecutionService) ScheduleEmailAction(ctx context.Context, tx *neo4j.ManagedTransaction, flowId, entityId string, scheduleAt time.Time, nextAction entity.FlowActionEntity) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "FlowExecutionService.ScheduleEmailAction")
	defer span.Finish()
	tracing.SetDefaultServiceSpanTags(ctx, span)

	tenant := common.GetTenantFromContext(ctx)

	// 1. Get the mailbox for contact or associate the best available mailbox
	flowExecutionSettings, err := s.getMailboxForContact(ctx, flowId, entityId)
	if err != nil {
		tracing.TraceErr(span, err)
		return err
	}

	if flowExecutionSettings == nil || flowExecutionSettings.Mailbox == nil {
		//compute the best available mailbox and associate

		// 1. get all available mailboxes
		// 2. select the mailbox with the fastest response time
		flowActionSenders, err := s.services.FlowService.FlowActionSenderGetList(ctx, []string{nextAction.Id})
		if err != nil {
			tracing.TraceErr(span, err)
			return err
		}

		availableMailboxes := []string{}
		for _, flowActionSender := range *flowActionSenders {
			if flowActionSender.Mailbox != nil {
				availableMailboxes = append(availableMailboxes, *flowActionSender.Mailbox)
			}
		}

		fastestMailbox, err := s.services.Neo4jRepositories.FlowActionExecutionReadRepository.GetFastestMailboxAvailable(ctx, availableMailboxes)
		if err != nil {
			tracing.TraceErr(span, err)
			return err
		}

		if fastestMailbox == nil {
			fastestMailbox = &availableMailboxes[0]
		}

		id, err := s.services.Neo4jRepositories.CommonReadRepository.GenerateId(ctx, tenant, model.NodeLabelFlowExecutionSettings)
		if err != nil {
			tracing.TraceErr(span, err)
			return err
		}
		flowExecutionSettings = &entity.FlowExecutionSettingsEntity{
			Id:       id,
			FlowId:   flowId,
			EntityId: entityId,
			Mailbox:  fastestMailbox,
		}

		node, err := s.services.Neo4jRepositories.FlowExecutionSettingsWriteRepository.Merge(ctx, tx, flowExecutionSettings)
		if err != nil {
			tracing.TraceErr(span, err)
			return err
		}
		flowExecutionSettings = mapper.MapDbNodeToFlowExecutionSettingsEntity(node)
	}

	// 2. Schedule the email action
	actualScheduleAt, err := s.getFirstAvailableSlotForMailbox(ctx, tx, *flowExecutionSettings.Mailbox, scheduleAt)
	if err != nil {
		tracing.TraceErr(span, err)
		return err
	}

	err = s.StoreNextActionExecutionEntity(ctx, tx, flowId, nextAction.Id, entityId, flowExecutionSettings.Mailbox, *actualScheduleAt)
	if err != nil {
		tracing.TraceErr(span, err)
		return err
	}

	return nil
}

func (s *flowExecutionService) getFirstAvailableSlotForMailbox(ctx context.Context, tx *neo4j.ManagedTransaction, mailbox string, scheduleAt time.Time) (*time.Time, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "FlowExecutionService.getFirstAvailableSlotForMailbox")
	defer span.Finish()
	tracing.SetDefaultServiceSpanTags(ctx, span)

	startDate := utils.StartOfDayInUTC(scheduleAt)
	endDate := utils.EndOfDayInUTC(scheduleAt)

	emailsScheduledAlready, err := s.services.Neo4jRepositories.FlowActionExecutionReadRepository.CountEmailsPerMailboxPerDay(ctx, tx, mailbox, startDate, endDate)
	if err != nil {
		tracing.TraceErr(span, err)
		return nil, err
	}

	if emailsScheduledAlready >= numberOfEmailsPerDay {
		// If the mailbox has already reached the daily limit, schedule the next email for the next day
		scheduleAt = startDate.Add(24 * time.Hour)
		// todo break recursive based on index or smth
		return s.getFirstAvailableSlotForMailbox(ctx, tx, mailbox, scheduleAt)
	} else {
		lastScheduledExecutionNode, err := s.services.Neo4jRepositories.FlowActionExecutionReadRepository.GetLastScheduledForMailbox(ctx, mailbox)
		if err != nil {
			tracing.TraceErr(span, err)
			return nil, err
		}
		lastScheduledExecution := mapper.MapDbNodeToFlowActionExecutionEntity(lastScheduledExecutionNode)

		//this is a bit wrong as will ocupy all the space in between the last scheduled email and the current schedule time
		if lastScheduledExecution != nil {
			if lastScheduledExecution.ScheduledAt.After(scheduleAt) {
				// If the last scheduled email is after the current schedule time, use the last scheduled time
				scheduleAt = scheduleAt.Add(time.Duration(utils.GenerateRandomInt(minMinutesBetweenEmails, maxMinutesBetweenEmails)) * time.Minute)
				return &scheduleAt, nil
			} else {
				// If the last scheduled email is before the current schedule time, use the current schedule time
				return &scheduleAt, nil
			}
		} else {
			scheduleAt = scheduleAt.Add(time.Duration(utils.GenerateRandomInt(minMinutesBetweenEmails, maxMinutesBetweenEmails)) * time.Minute)
			return &scheduleAt, nil
		}
	}
}

func (s *flowExecutionService) getMailboxForContact(ctx context.Context, flowId, contactId string) (*entity.FlowExecutionSettingsEntity, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "FlowExecutionService.getMailboxForContact")
	defer span.Finish()
	tracing.SetDefaultServiceSpanTags(ctx, span)

	node, err := s.services.Neo4jRepositories.FlowExecutionSettingsReadRepository.GetByEntityId(ctx, flowId, contactId)
	if err != nil {
		tracing.TraceErr(span, err)
		return nil, err
	}

	return mapper.MapDbNodeToFlowExecutionSettingsEntity(node), nil
}

func (s *flowExecutionService) StoreNextActionExecutionEntity(ctx context.Context, tx *neo4j.ManagedTransaction, flowId, actionId, contactId string, mailbox *string, executionTime time.Time) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "FlowExecutionService.StoreNextActionExecutionEntity")
	defer span.Finish()
	tracing.SetDefaultServiceSpanTags(ctx, span)

	tenant := common.GetTenantFromContext(ctx)

	id, err := s.services.Neo4jRepositories.CommonReadRepository.GenerateId(ctx, tenant, model.NodeLabelFlowActionExecution)
	if err != nil {
		tracing.TraceErr(span, err)
		return err
	}

	// After the wait duration, the next action will be executed
	actionExecution := entity.FlowActionExecutionEntity{
		Id:          id,
		FlowId:      flowId,
		ActionId:    actionId,
		ContactId:   contactId,
		Mailbox:     mailbox,
		ScheduledAt: executionTime,
		Status:      entity.FlowActionExecutionStatusPending,
	}

	_, err = s.services.Neo4jRepositories.FlowActionExecutionWriteRepository.Merge(ctx, tx, &actionExecution)
	if err != nil {
		tracing.TraceErr(span, err)
		return err
	}

	return nil
}