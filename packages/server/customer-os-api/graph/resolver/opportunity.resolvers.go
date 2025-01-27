package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.56

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-api/constants"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-api/dataloader"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-api/graph/generated"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-api/graph/model"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-api/mapper"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-api/tracing"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/common"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/utils"
	neo4jentity "github.com/openline-ai/openline-customer-os/packages/server/customer-os-neo4j-repository/entity"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

// OpportunitySave is the resolver for the opportunity_save field.
func (r *mutationResolver) OpportunitySave(ctx context.Context, input model.OpportunitySaveInput) (*model.Opportunity, error) {
	ctx, span := tracing.StartGraphQLTracerSpan(ctx, "MutationResolver.OpportunitySave", graphql.GetOperationContext(ctx))
	defer span.Finish()
	tracing.SetDefaultResolverSpanTags(ctx, span)
	tracing.LogObjectAsJson(span, "request.input", input)

	tenant := common.GetTenantFromContext(ctx)

	id, err := r.Services.CommonServices.OpportunityService.Save(ctx, nil, tenant, input.OrganizationID, input.OpportunityID, mapper.MapOpportunitySaveInputToEntity(input))
	if err != nil {
		tracing.TraceErr(span, err)
		graphql.AddErrorf(ctx, "Failed to save opportunity")
		return nil, err
	}

	e, err := r.Services.CommonServices.OpportunityService.GetById(ctx, tenant, *id)
	if err != nil {
		tracing.TraceErr(span, err)
		graphql.AddErrorf(ctx, "Failed to fetch opportunity details")
		return nil, err
	}

	return mapper.MapEntityToOpportunity(e), nil
}

// OpportunityArchive is the resolver for the opportunity_Archive field.
func (r *mutationResolver) OpportunityArchive(ctx context.Context, id string) (*model.ActionResponse, error) {
	ctx, span := tracing.StartGraphQLTracerSpan(ctx, "MutationResolver.OpportunityArchive", graphql.GetOperationContext(ctx))
	defer span.Finish()
	tracing.SetDefaultResolverSpanTags(ctx, span)
	span.LogFields(log.String("request.id", id))

	tenant := common.GetTenantFromContext(ctx)

	err := r.Services.CommonServices.OpportunityService.Archive(ctx, tenant, id)
	if err != nil {
		tracing.TraceErr(span, err)
		return &model.ActionResponse{Accepted: false}, err
	}

	return &model.ActionResponse{Accepted: true}, nil
}

// OpportunityRenewalUpdate is the resolver for the opportunityRenewalUpdate field.
func (r *mutationResolver) OpportunityRenewalUpdate(ctx context.Context, input model.OpportunityRenewalUpdateInput, ownerUserID *string) (*model.Opportunity, error) {
	ctx, span := tracing.StartGraphQLTracerSpan(ctx, "MutationResolver.OpportunityRenewalUpdate", graphql.GetOperationContext(ctx))
	defer span.Finish()
	tracing.SetDefaultResolverSpanTags(ctx, span)
	tracing.LogObjectAsJson(span, "request.input", input)

	tenant := common.GetTenantFromContext(ctx)

	err := r.Services.OpportunityService.UpdateRenewal(ctx, input.OpportunityID, mapper.MapOpportunityRenewalLikelihoodFromModel(input.RenewalLikelihood), input.Amount, input.Comments, input.OwnerUserID, input.RenewalAdjustedRate, utils.IfNotNilStringWithDefault(input.AppSource, constants.AppSourceCustomerOsApi))
	if err != nil {
		tracing.TraceErr(span, err)
		graphql.AddErrorf(ctx, "Failed to update opportunity renewal %s", input.OpportunityID)
		return &model.Opportunity{ID: input.OpportunityID}, nil
	}
	opportunityEntity, err := r.Services.CommonServices.OpportunityService.GetById(ctx, tenant, input.OpportunityID)
	if err != nil {
		tracing.TraceErr(span, err)
		graphql.AddErrorf(ctx, "Failed fetching opportunity details. Opportunity id: %s", input.OpportunityID)
		return &model.Opportunity{ID: input.OpportunityID}, nil
	}

	return mapper.MapEntityToOpportunity(opportunityEntity), nil
}

// OpportunityRenewalUpdateAllForOrganization is the resolver for the opportunityRenewal_UpdateAllForOrganization field.
func (r *mutationResolver) OpportunityRenewalUpdateAllForOrganization(ctx context.Context, input model.OpportunityRenewalUpdateAllForOrganizationInput) (*model.Organization, error) {
	ctx, span := tracing.StartGraphQLTracerSpan(ctx, "MutationResolver.OpportunityRenewalUpdate", graphql.GetOperationContext(ctx))
	defer span.Finish()
	tracing.SetDefaultResolverSpanTags(ctx, span)
	tracing.LogObjectAsJson(span, "request.input", input)

	tenant := common.GetTenantFromContext(ctx)

	if input.RenewalLikelihood != nil {
		err := r.Services.OpportunityService.UpdateRenewalsForOrganization(ctx, input.OrganizationID, mapper.MapOpportunityRenewalLikelihoodFromModel(input.RenewalLikelihood), input.RenewalAdjustedRate)
		if err != nil {
			tracing.TraceErr(span, err)
			graphql.AddErrorf(ctx, "Failed to update renewal opportunities for organization %s", input.OrganizationID)
			return nil, nil
		}
	}
	organizationEntity, err := r.Services.CommonServices.OrganizationService.GetById(ctx, tenant, input.OrganizationID)
	if err != nil {
		tracing.TraceErr(span, err)
		graphql.AddErrorf(ctx, "Failed fetching organization details. Organization id: %s", input.OrganizationID)
		return nil, nil
	}

	return mapper.MapEntityToOrganization(organizationEntity), nil
}

// Organization is the resolver for the organization field.
func (r *opportunityResolver) Organization(ctx context.Context, obj *model.Opportunity) (*model.Organization, error) {
	ctx = tracing.EnrichCtxWithSpanCtxForGraphQL(ctx, graphql.GetOperationContext(ctx))

	organizationEntityNillable, err := dataloader.For(ctx).GetOrganizationForOpportunityOptional(ctx, obj.Metadata.ID)
	if err != nil {
		tracing.TraceErr(opentracing.SpanFromContext(ctx), err)
		r.log.Errorf("error fetching organization for opportunity %s: %s", obj.Metadata.ID, err.Error())
		graphql.AddErrorf(ctx, "error fetching organization for opportunity %s", obj.Metadata.ID)
		return nil, nil
	}
	return mapper.MapEntityToOrganization(organizationEntityNillable), nil
}

// CreatedBy is the resolver for the createdBy field.
func (r *opportunityResolver) CreatedBy(ctx context.Context, obj *model.Opportunity) (*model.User, error) {
	ctx = tracing.EnrichCtxWithSpanCtxForGraphQL(ctx, graphql.GetOperationContext(ctx))

	userEntityNillable, err := dataloader.For(ctx).GetUserCreatorForOpportunity(ctx, obj.ID)
	if err != nil {
		tracing.TraceErr(opentracing.SpanFromContext(ctx), err)
		r.log.Errorf("error fetching user creator for opportunity %s: %s", obj.ID, err.Error())
		graphql.AddErrorf(ctx, "error fetching user creator for opportunity %s", obj.ID)
		return nil, nil
	}
	return mapper.MapEntityToUser(userEntityNillable), nil
}

// Owner is the resolver for the owner field.
func (r *opportunityResolver) Owner(ctx context.Context, obj *model.Opportunity) (*model.User, error) {
	ctx = tracing.EnrichCtxWithSpanCtxForGraphQL(ctx, graphql.GetOperationContext(ctx))

	userEntityNillable, err := dataloader.For(ctx).GetUserOwnerForOpportunity(ctx, obj.ID)
	if err != nil {
		tracing.TraceErr(opentracing.SpanFromContext(ctx), err)
		r.log.Errorf("error fetching user owner for opportunity %s: %s", obj.ID, err.Error())
		graphql.AddErrorf(ctx, "error fetching user owner for opportunity %s", obj.ID)
		return nil, nil
	}
	return mapper.MapEntityToUser(userEntityNillable), nil
}

// ExternalLinks is the resolver for the externalLinks field.
func (r *opportunityResolver) ExternalLinks(ctx context.Context, obj *model.Opportunity) ([]*model.ExternalSystem, error) {
	ctx = tracing.EnrichCtxWithSpanCtxForGraphQL(ctx, graphql.GetOperationContext(ctx))

	entities, err := dataloader.For(ctx).GetExternalSystemsForOpportunity(ctx, obj.ID)
	if err != nil {
		tracing.TraceErr(opentracing.SpanFromContext(ctx), err)
		r.log.Errorf("Failed to get external system for opportunity %s: %s", obj.ID, err.Error())
		graphql.AddErrorf(ctx, "Failed to get external system for opportunity %s", obj.ID)
		return nil, nil
	}
	return mapper.MapEntitiesToExternalSystems(entities), nil
}

// Opportunity is the resolver for the opportunity field.
func (r *queryResolver) Opportunity(ctx context.Context, id string) (*model.Opportunity, error) {
	ctx, span := tracing.StartGraphQLTracerSpan(ctx, "QueryResolver.Opportunity", graphql.GetOperationContext(ctx))
	defer span.Finish()
	tracing.SetDefaultResolverSpanTags(ctx, span)
	span.LogFields(log.String("request.issueID", id))

	tenant := common.GetTenantFromContext(ctx)

	opportunityEntity, err := r.Services.CommonServices.OpportunityService.GetById(ctx, tenant, id)
	if err != nil || opportunityEntity == nil {
		tracing.TraceErr(span, err)
		graphql.AddErrorf(ctx, "Opportunity with id %s not found", id)
		return nil, err
	}
	return mapper.MapEntityToOpportunity(opportunityEntity), nil
}

// OpportunitiesLinkedToOrganizations is the resolver for the opportunities_LinkedToOrganizations field.
func (r *queryResolver) OpportunitiesLinkedToOrganizations(ctx context.Context, pagination *model.Pagination) (*model.OpportunityPage, error) {
	ctx, span := tracing.StartGraphQLTracerSpan(ctx, "QueryResolver.OpportunitiesLinkedToOrganizations", graphql.GetOperationContext(ctx))
	defer span.Finish()
	tracing.SetDefaultResolverSpanTags(ctx, span)
	tracing.LogObjectAsJson(span, "request.pagination", pagination)

	tenant := common.GetTenantFromContext(ctx)

	if pagination == nil {
		pagination = &model.Pagination{Page: 0, Limit: 0}
	}
	paginatedResult, err := r.Services.CommonServices.OpportunityService.GetPaginatedOrganizationOpportunities(ctx, tenant, pagination.Page, pagination.Limit)
	if err != nil {
		tracing.TraceErr(span, err)
		graphql.AddErrorf(ctx, "Error while fetching opportunities")
		return nil, err
	}
	return &model.OpportunityPage{
		Content:       mapper.MapEntitiesToOpportunities(paginatedResult.Rows.(*neo4jentity.OpportunityEntities)),
		TotalPages:    paginatedResult.TotalPages,
		TotalElements: paginatedResult.TotalRows,
	}, err
}

// Opportunity returns generated.OpportunityResolver implementation.
func (r *Resolver) Opportunity() generated.OpportunityResolver { return &opportunityResolver{r} }

type opportunityResolver struct{ *Resolver }
