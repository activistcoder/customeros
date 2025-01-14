package command_handler

import (
	"github.com/openline-ai/openline-customer-os/packages/server/events-processing-platform/config"
	"github.com/openline-ai/openline-customer-os/packages/server/events-processing-platform/logger"
	"github.com/openline-ai/openline-customer-os/packages/server/events/eventbuffer"
	"github.com/openline-ai/openline-customer-os/packages/server/events/eventstore"
)

// CommandHandlers acts as a container for all command handlers.
type CommandHandlers struct {
	UpsertOrganization       UpsertOrganizationCommandHandler
	LinkPhoneNumberCommand   LinkPhoneNumberCommandHandler
	LinkLocationCommand      LinkLocationCommandHandler
	ShowOrganizationCommand  ShowOrganizationCommandHandler
	UpsertCustomFieldCommand UpsertCustomFieldCommandHandler
	AddParentCommand         AddParentCommandHandler
	RemoveParentCommand      RemoveParentCommandHandler
	RefreshArr               RefreshArrCommandHandler
	UpdateOnboardingStatus   UpdateOnboardingStatusCommandHandler
	UpdateOrganizationOwner  UpdateOrganizationOwnerCommandHandler
}

func NewCommandHandlers(log logger.Logger, cfg *config.Config, es eventstore.AggregateStore, ebs *eventbuffer.EventBufferStoreService) *CommandHandlers {
	return &CommandHandlers{
		UpsertOrganization:       NewUpsertOrganizationCommandHandler(log, es),
		LinkPhoneNumberCommand:   NewLinkPhoneNumberCommandHandler(log, es),
		LinkLocationCommand:      NewLinkLocationCommandHandler(log, es),
		ShowOrganizationCommand:  NewShowOrganizationCommandHandler(log, es),
		UpsertCustomFieldCommand: NewUpsertCustomFieldCommandHandler(log, es),
		AddParentCommand:         NewAddParentCommandHandler(log, es),
		RemoveParentCommand:      NewRemoveParentCommandHandler(log, es),
		RefreshArr:               NewRefreshArrCommandHandler(log, es, cfg.Utils),
		UpdateOnboardingStatus:   NewUpdateOnboardingStatusCommandHandler(log, es, cfg.Utils),
		UpdateOrganizationOwner:  NewUpdateOrganizationOwnerCommandHandler(log, es, cfg.Utils, ebs),
	}
}
