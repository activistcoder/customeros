package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.44

import (
	"context"
	"fmt"
	"time"

	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-api-sdk/graph/generated"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-api-sdk/graph/model"
)

// LastTouchPointTimelineEvent is the resolver for the lastTouchPointTimelineEvent field.
func (r *lastTouchpointResolver) LastTouchPointTimelineEvent(ctx context.Context, obj *model.LastTouchpoint) (model.TimelineEvent, error) {
	panic(fmt.Errorf("not implemented: LastTouchPointTimelineEvent - lastTouchPointTimelineEvent"))
}

// OrganizationCreate is the resolver for the organization_Create field.
func (r *mutationResolver) OrganizationCreate(ctx context.Context, input model.OrganizationInput) (*model.Organization, error) {
	panic(fmt.Errorf("not implemented: OrganizationCreate - organization_Create"))
}

// OrganizationUpdate is the resolver for the organization_Update field.
func (r *mutationResolver) OrganizationUpdate(ctx context.Context, input model.OrganizationUpdateInput) (*model.Organization, error) {
	panic(fmt.Errorf("not implemented: OrganizationUpdate - organization_Update"))
}

// OrganizationArchive is the resolver for the organization_Archive field.
func (r *mutationResolver) OrganizationArchive(ctx context.Context, id string) (*model.Result, error) {
	panic(fmt.Errorf("not implemented: OrganizationArchive - organization_Archive"))
}

// OrganizationArchiveAll is the resolver for the organization_ArchiveAll field.
func (r *mutationResolver) OrganizationArchiveAll(ctx context.Context, ids []string) (*model.Result, error) {
	panic(fmt.Errorf("not implemented: OrganizationArchiveAll - organization_ArchiveAll"))
}

// OrganizationHide is the resolver for the organization_Hide field.
func (r *mutationResolver) OrganizationHide(ctx context.Context, id string) (string, error) {
	panic(fmt.Errorf("not implemented: OrganizationHide - organization_Hide"))
}

// OrganizationHideAll is the resolver for the organization_HideAll field.
func (r *mutationResolver) OrganizationHideAll(ctx context.Context, ids []string) (*model.Result, error) {
	panic(fmt.Errorf("not implemented: OrganizationHideAll - organization_HideAll"))
}

// OrganizationShow is the resolver for the organization_Show field.
func (r *mutationResolver) OrganizationShow(ctx context.Context, id string) (string, error) {
	panic(fmt.Errorf("not implemented: OrganizationShow - organization_Show"))
}

// OrganizationShowAll is the resolver for the organization_ShowAll field.
func (r *mutationResolver) OrganizationShowAll(ctx context.Context, ids []string) (*model.Result, error) {
	panic(fmt.Errorf("not implemented: OrganizationShowAll - organization_ShowAll"))
}

// OrganizationMerge is the resolver for the organization_Merge field.
func (r *mutationResolver) OrganizationMerge(ctx context.Context, primaryOrganizationID string, mergedOrganizationIds []string) (*model.Organization, error) {
	panic(fmt.Errorf("not implemented: OrganizationMerge - organization_Merge"))
}

// OrganizationAddSubsidiary is the resolver for the organization_AddSubsidiary field.
func (r *mutationResolver) OrganizationAddSubsidiary(ctx context.Context, input model.LinkOrganizationsInput) (*model.Organization, error) {
	panic(fmt.Errorf("not implemented: OrganizationAddSubsidiary - organization_AddSubsidiary"))
}

// OrganizationRemoveSubsidiary is the resolver for the organization_RemoveSubsidiary field.
func (r *mutationResolver) OrganizationRemoveSubsidiary(ctx context.Context, organizationID string, subsidiaryID string) (*model.Organization, error) {
	panic(fmt.Errorf("not implemented: OrganizationRemoveSubsidiary - organization_RemoveSubsidiary"))
}

// OrganizationAddNewLocation is the resolver for the organization_AddNewLocation field.
func (r *mutationResolver) OrganizationAddNewLocation(ctx context.Context, organizationID string) (*model.Location, error) {
	panic(fmt.Errorf("not implemented: OrganizationAddNewLocation - organization_AddNewLocation"))
}

// OrganizationAddSocial is the resolver for the organization_AddSocial field.
func (r *mutationResolver) OrganizationAddSocial(ctx context.Context, organizationID string, input model.SocialInput) (*model.Social, error) {
	panic(fmt.Errorf("not implemented: OrganizationAddSocial - organization_AddSocial"))
}

// OrganizationSetOwner is the resolver for the organization_SetOwner field.
func (r *mutationResolver) OrganizationSetOwner(ctx context.Context, organizationID string, userID string) (*model.Organization, error) {
	panic(fmt.Errorf("not implemented: OrganizationSetOwner - organization_SetOwner"))
}

// OrganizationUnsetOwner is the resolver for the organization_UnsetOwner field.
func (r *mutationResolver) OrganizationUnsetOwner(ctx context.Context, organizationID string) (*model.Organization, error) {
	panic(fmt.Errorf("not implemented: OrganizationUnsetOwner - organization_UnsetOwner"))
}

// OrganizationUpdateOnboardingStatus is the resolver for the organization_UpdateOnboardingStatus field.
func (r *mutationResolver) OrganizationUpdateOnboardingStatus(ctx context.Context, input model.OnboardingStatusInput) (*model.Organization, error) {
	panic(fmt.Errorf("not implemented: OrganizationUpdateOnboardingStatus - organization_UpdateOnboardingStatus"))
}

// Contracts is the resolver for the contracts field.
func (r *organizationResolver) Contracts(ctx context.Context, obj *model.Organization) ([]*model.Contract, error) {
	panic(fmt.Errorf("not implemented: Contracts - contracts"))
}

// CustomFields is the resolver for the customFields field.
func (r *organizationResolver) CustomFields(ctx context.Context, obj *model.Organization) ([]*model.CustomField, error) {
	panic(fmt.Errorf("not implemented: CustomFields - customFields"))
}

// Domains is the resolver for the domains field.
func (r *organizationResolver) Domains(ctx context.Context, obj *model.Organization) ([]string, error) {
	panic(fmt.Errorf("not implemented: Domains - domains"))
}

// Locations is the resolver for the locations field.
func (r *organizationResolver) Locations(ctx context.Context, obj *model.Organization) ([]*model.Location, error) {
	panic(fmt.Errorf("not implemented: Locations - locations"))
}

// Owner is the resolver for the owner field.
func (r *organizationResolver) Owner(ctx context.Context, obj *model.Organization) (*model.User, error) {
	panic(fmt.Errorf("not implemented: Owner - owner"))
}

// ParentCompanies is the resolver for the parentCompanies field.
func (r *organizationResolver) ParentCompanies(ctx context.Context, obj *model.Organization) ([]*model.LinkedOrganization, error) {
	panic(fmt.Errorf("not implemented: ParentCompanies - parentCompanies"))
}

// SocialMedia is the resolver for the socialMedia field.
func (r *organizationResolver) SocialMedia(ctx context.Context, obj *model.Organization) ([]*model.Social, error) {
	panic(fmt.Errorf("not implemented: SocialMedia - socialMedia"))
}

// Subsidiaries is the resolver for the subsidiaries field.
func (r *organizationResolver) Subsidiaries(ctx context.Context, obj *model.Organization) ([]*model.LinkedOrganization, error) {
	panic(fmt.Errorf("not implemented: Subsidiaries - subsidiaries"))
}

// Tags is the resolver for the tags field.
func (r *organizationResolver) Tags(ctx context.Context, obj *model.Organization) ([]*model.Tag, error) {
	panic(fmt.Errorf("not implemented: Tags - tags"))
}

// TimelineEvents is the resolver for the timelineEvents field.
func (r *organizationResolver) TimelineEvents(ctx context.Context, obj *model.Organization, from *time.Time, size int, timelineEventTypes []model.TimelineEventType) ([]model.TimelineEvent, error) {
	panic(fmt.Errorf("not implemented: TimelineEvents - timelineEvents"))
}

// Contacts is the resolver for the contacts field.
func (r *organizationResolver) Contacts(ctx context.Context, obj *model.Organization, pagination *model.Pagination, where *model.Filter, sort []*model.SortBy) (*model.ContactsPage, error) {
	panic(fmt.Errorf("not implemented: Contacts - contacts"))
}

// JobRoles is the resolver for the jobRoles field.
func (r *organizationResolver) JobRoles(ctx context.Context, obj *model.Organization) ([]*model.JobRole, error) {
	panic(fmt.Errorf("not implemented: JobRoles - jobRoles"))
}

// Emails is the resolver for the emails field.
func (r *organizationResolver) Emails(ctx context.Context, obj *model.Organization) ([]*model.Email, error) {
	panic(fmt.Errorf("not implemented: Emails - emails"))
}

// PhoneNumbers is the resolver for the phoneNumbers field.
func (r *organizationResolver) PhoneNumbers(ctx context.Context, obj *model.Organization) ([]*model.PhoneNumber, error) {
	panic(fmt.Errorf("not implemented: PhoneNumbers - phoneNumbers"))
}

// SuggestedMergeTo is the resolver for the suggestedMergeTo field.
func (r *organizationResolver) SuggestedMergeTo(ctx context.Context, obj *model.Organization) ([]*model.SuggestedMergeOrganization, error) {
	panic(fmt.Errorf("not implemented: SuggestedMergeTo - suggestedMergeTo"))
}

// FieldSets is the resolver for the fieldSets field.
func (r *organizationResolver) FieldSets(ctx context.Context, obj *model.Organization) ([]*model.FieldSet, error) {
	panic(fmt.Errorf("not implemented: FieldSets - fieldSets"))
}

// EntityTemplate is the resolver for the entityTemplate field.
func (r *organizationResolver) EntityTemplate(ctx context.Context, obj *model.Organization) (*model.EntityTemplate, error) {
	panic(fmt.Errorf("not implemented: EntityTemplate - entityTemplate"))
}

// TimelineEventsTotalCount is the resolver for the timelineEventsTotalCount field.
func (r *organizationResolver) TimelineEventsTotalCount(ctx context.Context, obj *model.Organization, timelineEventTypes []model.TimelineEventType) (int64, error) {
	panic(fmt.Errorf("not implemented: TimelineEventsTotalCount - timelineEventsTotalCount"))
}

// ExternalLinks is the resolver for the externalLinks field.
func (r *organizationResolver) ExternalLinks(ctx context.Context, obj *model.Organization) ([]*model.ExternalSystem, error) {
	panic(fmt.Errorf("not implemented: ExternalLinks - externalLinks"))
}

// IssueSummaryByStatus is the resolver for the issueSummaryByStatus field.
func (r *organizationResolver) IssueSummaryByStatus(ctx context.Context, obj *model.Organization) ([]*model.IssueSummaryByStatus, error) {
	panic(fmt.Errorf("not implemented: IssueSummaryByStatus - issueSummaryByStatus"))
}

// Socials is the resolver for the socials field.
func (r *organizationResolver) Socials(ctx context.Context, obj *model.Organization) ([]*model.Social, error) {
	panic(fmt.Errorf("not implemented: Socials - socials"))
}

// LastTouchPointTimelineEvent is the resolver for the lastTouchPointTimelineEvent field.
func (r *organizationResolver) LastTouchPointTimelineEvent(ctx context.Context, obj *model.Organization) (model.TimelineEvent, error) {
	panic(fmt.Errorf("not implemented: LastTouchPointTimelineEvent - lastTouchPointTimelineEvent"))
}

// SubsidiaryOf is the resolver for the subsidiaryOf field.
func (r *organizationResolver) SubsidiaryOf(ctx context.Context, obj *model.Organization) ([]*model.LinkedOrganization, error) {
	panic(fmt.Errorf("not implemented: SubsidiaryOf - subsidiaryOf"))
}

// Organizations is the resolver for the organizations field.
func (r *queryResolver) Organizations(ctx context.Context, pagination *model.Pagination, where *model.Filter, sort []*model.SortBy) (*model.OrganizationPage, error) {
	panic(fmt.Errorf("not implemented: Organizations - organizations"))
}

// Organization is the resolver for the organization field.
func (r *queryResolver) Organization(ctx context.Context, id string) (*model.Organization, error) {
	panic(fmt.Errorf("not implemented: Organization - organization"))
}

// OrganizationByCustomerOsID is the resolver for the organization_ByCustomerOsId field.
func (r *queryResolver) OrganizationByCustomerOsID(ctx context.Context, customerOsID string) (*model.Organization, error) {
	panic(fmt.Errorf("not implemented: OrganizationByCustomerOsID - organization_ByCustomerOsId"))
}

// OrganizationDistinctOwners is the resolver for the organization_DistinctOwners field.
func (r *queryResolver) OrganizationDistinctOwners(ctx context.Context) ([]*model.User, error) {
	panic(fmt.Errorf("not implemented: OrganizationDistinctOwners - organization_DistinctOwners"))
}

// LastTouchpoint returns generated.LastTouchpointResolver implementation.
func (r *Resolver) LastTouchpoint() generated.LastTouchpointResolver {
	return &lastTouchpointResolver{r}
}

// Organization returns generated.OrganizationResolver implementation.
func (r *Resolver) Organization() generated.OrganizationResolver { return &organizationResolver{r} }

type lastTouchpointResolver struct{ *Resolver }
type organizationResolver struct{ *Resolver }