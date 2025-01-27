package contract

import (
	"context"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/utils"
	neo4jentity "github.com/openline-ai/openline-customer-os/packages/server/customer-os-neo4j-repository/entity"
	neo4jenum "github.com/openline-ai/openline-customer-os/packages/server/customer-os-neo4j-repository/enum"
	neo4jtest "github.com/openline-ai/openline-customer-os/packages/server/customer-os-neo4j-repository/test"
	"github.com/openline-ai/openline-customer-os/packages/server/events-processing-platform-subscribers/constants"
	"github.com/openline-ai/openline-customer-os/packages/server/events-processing-platform-subscribers/test/mocked_grpc"
	opportunitypb "github.com/openline-ai/openline-customer-os/packages/server/events-processing-proto/gen/proto/go/api/grpc/v1/opportunity"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestContractEventHandler_UpdateRenewalNextCycleDate_CreateRenewalOpportunityWhenMissing(t *testing.T) {
	ctx := context.Background()
	defer tearDownTestCase(ctx, testDatabase)(t)

	// prepare neo4j data
	neo4jtest.CreateTenant(ctx, testDatabase.Driver, tenantName)
	serviceStartedAt, _ := utils.UnmarshalDateTime("2021-01-01T00:00:00Z")
	orgId := neo4jtest.CreateOrganization(ctx, testDatabase.Driver, tenantName, neo4jentity.OrganizationEntity{})
	contractId := neo4jtest.CreateContractForOrganization(ctx, testDatabase.Driver, tenantName, orgId, neo4jentity.ContractEntity{
		ServiceStartedAt: serviceStartedAt,
		LengthInMonths:   1,
	})

	// prepare grpc client
	calledEventsPlatformToCreateRenewalOpportunity := false
	opportunityCallbacks := mocked_grpc.MockOpportunityServiceCallbacks{
		CreateRenewalOpportunity: func(context context.Context, op *opportunitypb.CreateRenewalOpportunityGrpcRequest) (*opportunitypb.OpportunityIdGrpcResponse, error) {
			require.Equal(t, tenantName, op.Tenant)
			require.Equal(t, "", op.LoggedInUserId)
			require.Equal(t, contractId, op.ContractId)
			require.Nil(t, op.CreatedAt)
			require.Nil(t, op.UpdatedAt)
			require.Equal(t, opportunitypb.RenewalLikelihood_HIGH_RENEWAL, op.RenewalLikelihood)
			require.Equal(t, constants.AppSourceEventProcessingPlatformSubscribers, op.SourceFields.AppSource)
			calledEventsPlatformToCreateRenewalOpportunity = true
			return &opportunitypb.OpportunityIdGrpcResponse{
				Id: "some-opportunity-id",
			}, nil
		},
	}
	mocked_grpc.SetOpportunityCallbacks(&opportunityCallbacks)

	// prepare event handler
	handler := NewContractHandler(testLogger, testDatabase.Services, testMockedGrpcClient)

	// EXECUTE
	err := handler.UpdateActiveRenewalOpportunityRenewDateAndArr(ctx, tenantName, contractId)
	require.Nil(t, err)

	// Check
	require.True(t, calledEventsPlatformToCreateRenewalOpportunity)
}

func TestContractEventHandler_UpdateRenewalNextCycleDate_MonthlyContract(t *testing.T) {
	ctx := context.Background()
	defer tearDownTestCase(ctx, testDatabase)(t)

	// prepare neo4j data
	neo4jtest.CreateTenant(ctx, testDatabase.Driver, tenantName)
	serviceStartedAt, _ := utils.UnmarshalDateTime("2021-01-01T00:00:00Z")
	orgId := neo4jtest.CreateOrganization(ctx, testDatabase.Driver, tenantName, neo4jentity.OrganizationEntity{})
	contractId := neo4jtest.CreateContractForOrganization(ctx, testDatabase.Driver, tenantName, orgId, neo4jentity.ContractEntity{
		ServiceStartedAt: serviceStartedAt,
		LengthInMonths:   1,
		AutoRenew:        true,
	})
	opportunityId := neo4jtest.CreateOpportunityForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.OpportunityEntity{
		InternalType:  neo4jenum.OpportunityInternalTypeRenewal,
		InternalStage: neo4jenum.OpportunityInternalStageOpen,
	})

	calledEventsPlatformToUpdateRenewalOpportunityNextCycleDate := false
	opportunityCallbacks := mocked_grpc.MockOpportunityServiceCallbacks{
		UpdateRenewalOpportunityNextCycleDate: func(context context.Context, op *opportunitypb.UpdateRenewalOpportunityNextCycleDateGrpcRequest) (*opportunitypb.OpportunityIdGrpcResponse, error) {
			require.Equal(t, tenantName, op.Tenant)
			require.Equal(t, opportunityId, op.OpportunityId)
			require.Equal(t, constants.AppSourceEventProcessingPlatformSubscribers, op.AppSource)
			require.Equal(t, startOfNextMonth(utils.Now()), *utils.TimestampProtoToTimePtr(op.RenewedAt))
			calledEventsPlatformToUpdateRenewalOpportunityNextCycleDate = true
			return &opportunitypb.OpportunityIdGrpcResponse{
				Id: opportunityId,
			}, nil
		},
		UpdateOpportunity: func(context context.Context, op *opportunitypb.UpdateOpportunityGrpcRequest) (*opportunitypb.OpportunityIdGrpcResponse, error) {
			return &opportunitypb.OpportunityIdGrpcResponse{
				Id: opportunityId,
			}, nil
		},
	}
	mocked_grpc.SetOpportunityCallbacks(&opportunityCallbacks)

	// prepare event handler
	handler := NewContractHandler(testLogger, testDatabase.Services, testMockedGrpcClient)

	// EXECUTE
	err := handler.UpdateActiveRenewalOpportunityRenewDateAndArr(ctx, tenantName, contractId)
	require.Nil(t, err)

	// Check
	require.True(t, calledEventsPlatformToUpdateRenewalOpportunityNextCycleDate)
}

func TestContractEventHandler_UpdateRenewalNextCycleDate_MonthlyContract_NotAutoRenewed(t *testing.T) {
	ctx := context.Background()
	defer tearDownTestCase(ctx, testDatabase)(t)

	// prepare neo4j data
	neo4jtest.CreateTenant(ctx, testDatabase.Driver, tenantName)
	serviceStartedAt, _ := utils.UnmarshalDateTime("2021-01-01T00:00:00Z")
	expectedRenewalDate, _ := utils.UnmarshalDateTime("2021-02-01T00:00:00Z")
	orgId := neo4jtest.CreateOrganization(ctx, testDatabase.Driver, tenantName, neo4jentity.OrganizationEntity{})
	contractId := neo4jtest.CreateContractForOrganization(ctx, testDatabase.Driver, tenantName, orgId, neo4jentity.ContractEntity{
		ServiceStartedAt: serviceStartedAt,
		LengthInMonths:   1,
		AutoRenew:        false,
	})
	opportunityId := neo4jtest.CreateOpportunityForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.OpportunityEntity{
		InternalType:  neo4jenum.OpportunityInternalTypeRenewal,
		InternalStage: neo4jenum.OpportunityInternalStageOpen,
	})

	calledEventsPlatformToUpdateRenewalOpportunityNextCycleDate := false
	opportunityCallbacks := mocked_grpc.MockOpportunityServiceCallbacks{
		UpdateRenewalOpportunityNextCycleDate: func(context context.Context, op *opportunitypb.UpdateRenewalOpportunityNextCycleDateGrpcRequest) (*opportunitypb.OpportunityIdGrpcResponse, error) {
			require.Equal(t, tenantName, op.Tenant)
			require.Equal(t, opportunityId, op.OpportunityId)
			require.Equal(t, constants.AppSourceEventProcessingPlatformSubscribers, op.AppSource)
			require.Equal(t, expectedRenewalDate, utils.TimestampProtoToTimePtr(op.RenewedAt))
			calledEventsPlatformToUpdateRenewalOpportunityNextCycleDate = true
			return &opportunitypb.OpportunityIdGrpcResponse{
				Id: opportunityId,
			}, nil
		},
		UpdateOpportunity: func(context context.Context, op *opportunitypb.UpdateOpportunityGrpcRequest) (*opportunitypb.OpportunityIdGrpcResponse, error) {
			return &opportunitypb.OpportunityIdGrpcResponse{
				Id: opportunityId,
			}, nil
		},
	}
	mocked_grpc.SetOpportunityCallbacks(&opportunityCallbacks)

	// prepare event handler
	handler := NewContractHandler(testLogger, testDatabase.Services, testMockedGrpcClient)

	// EXECUTE
	err := handler.UpdateActiveRenewalOpportunityRenewDateAndArr(ctx, tenantName, contractId)
	require.Nil(t, err)

	// Check
	require.True(t, calledEventsPlatformToUpdateRenewalOpportunityNextCycleDate)
}

func TestContractEventHandler_UpdateRenewalNextCycleDate_QuarterlyContract(t *testing.T) {
	ctx := context.Background()
	defer tearDownTestCase(ctx, testDatabase)(t)

	// prepare neo4j data
	neo4jtest.CreateTenant(ctx, testDatabase.Driver, tenantName)
	yesterday := utils.Now().AddDate(0, 0, -1)
	orgId := neo4jtest.CreateOrganization(ctx, testDatabase.Driver, tenantName, neo4jentity.OrganizationEntity{})
	contractId := neo4jtest.CreateContractForOrganization(ctx, testDatabase.Driver, tenantName, orgId, neo4jentity.ContractEntity{
		ServiceStartedAt: &yesterday,
		LengthInMonths:   3,
		AutoRenew:        true,
	})
	opportunityId := neo4jtest.CreateOpportunityForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.OpportunityEntity{
		InternalType:  neo4jenum.OpportunityInternalTypeRenewal,
		InternalStage: neo4jenum.OpportunityInternalStageOpen,
	})

	calledEventsPlatformToUpdateRenewalOpportunityNextCycleDate := false
	opportunityCallbacks := mocked_grpc.MockOpportunityServiceCallbacks{
		UpdateRenewalOpportunityNextCycleDate: func(context context.Context, op *opportunitypb.UpdateRenewalOpportunityNextCycleDateGrpcRequest) (*opportunitypb.OpportunityIdGrpcResponse, error) {
			require.Equal(t, tenantName, op.Tenant)
			require.Equal(t, opportunityId, op.OpportunityId)
			require.Equal(t, constants.AppSourceEventProcessingPlatformSubscribers, op.AppSource)
			in1Quarter := yesterday.AddDate(0, 3, 0)
			require.Equal(t, utils.ToDate(in1Quarter), *utils.TimestampProtoToTimePtr(op.RenewedAt))
			calledEventsPlatformToUpdateRenewalOpportunityNextCycleDate = true
			return &opportunitypb.OpportunityIdGrpcResponse{
				Id: opportunityId,
			}, nil
		},
		UpdateOpportunity: func(context context.Context, op *opportunitypb.UpdateOpportunityGrpcRequest) (*opportunitypb.OpportunityIdGrpcResponse, error) {
			return &opportunitypb.OpportunityIdGrpcResponse{
				Id: opportunityId,
			}, nil
		},
	}
	mocked_grpc.SetOpportunityCallbacks(&opportunityCallbacks)

	// prepare event handler
	handler := NewContractHandler(testLogger, testDatabase.Services, testMockedGrpcClient)

	// EXECUTE
	err := handler.UpdateActiveRenewalOpportunityRenewDateAndArr(ctx, tenantName, contractId)
	require.Nil(t, err)

	// Check
	require.True(t, calledEventsPlatformToUpdateRenewalOpportunityNextCycleDate)
}

func TestContractEventHandler_UpdateRenewalNextCycleDate_AnnualContract(t *testing.T) {
	ctx := context.Background()
	defer tearDownTestCase(ctx, testDatabase)(t)

	// prepare neo4j data
	neo4jtest.CreateTenant(ctx, testDatabase.Driver, tenantName)
	serviceStartedAt, _ := utils.UnmarshalDateTime("2021-01-01T00:00:00Z")
	orgId := neo4jtest.CreateOrganization(ctx, testDatabase.Driver, tenantName, neo4jentity.OrganizationEntity{})
	contractId := neo4jtest.CreateContractForOrganization(ctx, testDatabase.Driver, tenantName, orgId, neo4jentity.ContractEntity{
		ServiceStartedAt: serviceStartedAt,
		LengthInMonths:   12,
		AutoRenew:        true,
	})
	opportunityId := neo4jtest.CreateOpportunityForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.OpportunityEntity{
		InternalType:  neo4jenum.OpportunityInternalTypeRenewal,
		InternalStage: neo4jenum.OpportunityInternalStageOpen,
	})

	// prepare grpc client
	calledEventsPlatformToUpdateRenewalOpportunityNextCycleDate := false
	opportunityCallbacks := mocked_grpc.MockOpportunityServiceCallbacks{
		UpdateRenewalOpportunityNextCycleDate: func(context context.Context, op *opportunitypb.UpdateRenewalOpportunityNextCycleDateGrpcRequest) (*opportunitypb.OpportunityIdGrpcResponse, error) {
			require.Equal(t, tenantName, op.Tenant)
			require.Equal(t, opportunityId, op.OpportunityId)
			require.Equal(t, constants.AppSourceEventProcessingPlatformSubscribers, op.AppSource)
			require.Equal(t, startOfNextYear(utils.Now()), *utils.TimestampProtoToTimePtr(op.RenewedAt))
			calledEventsPlatformToUpdateRenewalOpportunityNextCycleDate = true
			return &opportunitypb.OpportunityIdGrpcResponse{
				Id: opportunityId,
			}, nil
		},
		UpdateOpportunity: func(context context.Context, op *opportunitypb.UpdateOpportunityGrpcRequest) (*opportunitypb.OpportunityIdGrpcResponse, error) {
			return &opportunitypb.OpportunityIdGrpcResponse{
				Id: opportunityId,
			}, nil
		},
	}
	mocked_grpc.SetOpportunityCallbacks(&opportunityCallbacks)

	// prepare event handler
	handler := NewContractHandler(testLogger, testDatabase.Services, testMockedGrpcClient)

	// EXECUTE
	err := handler.UpdateActiveRenewalOpportunityRenewDateAndArr(ctx, tenantName, contractId)
	require.Nil(t, err)

	// Check
	require.True(t, calledEventsPlatformToUpdateRenewalOpportunityNextCycleDate)
}

func TestContractEventHandler_UpdateRenewalNextCycleDate_MultiAnnualContract(t *testing.T) {
	ctx := context.Background()
	defer tearDownTestCase(ctx, testDatabase)(t)

	// prepare neo4j data
	neo4jtest.CreateTenant(ctx, testDatabase.Driver, tenantName)
	yesterday := utils.Now().AddDate(0, 0, -1)
	orgId := neo4jtest.CreateOrganization(ctx, testDatabase.Driver, tenantName, neo4jentity.OrganizationEntity{})
	contractId := neo4jtest.CreateContractForOrganization(ctx, testDatabase.Driver, tenantName, orgId, neo4jentity.ContractEntity{
		ServiceStartedAt: &yesterday,
		LengthInMonths:   int64(120),
		AutoRenew:        true,
	})
	opportunityId := neo4jtest.CreateOpportunityForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.OpportunityEntity{
		InternalType:  neo4jenum.OpportunityInternalTypeRenewal,
		InternalStage: neo4jenum.OpportunityInternalStageOpen,
	})

	// prepare grpc client
	calledEventsPlatformToUpdateRenewalOpportunityNextCycleDate := false
	opportunityCallbacks := mocked_grpc.MockOpportunityServiceCallbacks{
		UpdateRenewalOpportunityNextCycleDate: func(context context.Context, op *opportunitypb.UpdateRenewalOpportunityNextCycleDateGrpcRequest) (*opportunitypb.OpportunityIdGrpcResponse, error) {
			require.Equal(t, tenantName, op.Tenant)
			require.Equal(t, opportunityId, op.OpportunityId)
			require.Equal(t, constants.AppSourceEventProcessingPlatformSubscribers, op.AppSource)
			in10Years := yesterday.AddDate(10, 0, 0)
			require.Equal(t, utils.ToDate(in10Years), *utils.TimestampProtoToTimePtr(op.RenewedAt))
			calledEventsPlatformToUpdateRenewalOpportunityNextCycleDate = true
			return &opportunitypb.OpportunityIdGrpcResponse{
				Id: opportunityId,
			}, nil
		},
		UpdateOpportunity: func(context context.Context, op *opportunitypb.UpdateOpportunityGrpcRequest) (*opportunitypb.OpportunityIdGrpcResponse, error) {
			return &opportunitypb.OpportunityIdGrpcResponse{
				Id: opportunityId,
			}, nil
		},
	}
	mocked_grpc.SetOpportunityCallbacks(&opportunityCallbacks)

	// prepare event handler
	handler := NewContractHandler(testLogger, testDatabase.Services, testMockedGrpcClient)

	// EXECUTE
	err := handler.UpdateActiveRenewalOpportunityRenewDateAndArr(ctx, tenantName, contractId)
	require.Nil(t, err)

	// Check
	require.True(t, calledEventsPlatformToUpdateRenewalOpportunityNextCycleDate)
}

func startOfNextMonth(current time.Time) time.Time {
	year, month, _ := current.Date()
	firstOfNextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, current.Location())

	// Handle December to January transition
	if month == time.December {
		firstOfNextMonth = time.Date(year+1, time.January, 1, 0, 0, 0, 0, current.Location())
	}
	return firstOfNextMonth
}

func startOfNextYear(current time.Time) time.Time {
	year, _, _ := current.Date()
	firstOfNextYear := time.Date(year+1, time.January, 1, 0, 0, 0, 0, current.Location())
	return firstOfNextYear
}

func TestContractEventHandler_UpdateRenewalArrForecast_OnlyOnceBilled(t *testing.T) {
	ctx := context.Background()
	defer tearDownTestCase(ctx, testDatabase)(t)

	// prepare neo4j data
	neo4jtest.CreateTenant(ctx, testDatabase.Driver, tenantName)
	serviceStartedAt, _ := utils.UnmarshalDateTime("2021-01-01T00:00:00Z")
	orgId := neo4jtest.CreateOrganization(ctx, testDatabase.Driver, tenantName, neo4jentity.OrganizationEntity{})
	contractId := neo4jtest.CreateContractForOrganization(ctx, testDatabase.Driver, tenantName, orgId, neo4jentity.ContractEntity{
		ServiceStartedAt: serviceStartedAt,
		LengthInMonths:   int64(12),
	})
	opportunityId := neo4jtest.CreateOpportunityForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.OpportunityEntity{
		InternalType:  neo4jenum.OpportunityInternalTypeRenewal,
		InternalStage: neo4jenum.OpportunityInternalStageOpen,
		RenewalDetails: neo4jentity.RenewalDetails{
			RenewalLikelihood: neo4jenum.RenewalLikelihoodHigh,
		},
	})
	neo4jtest.CreateServiceLineItemForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.ServiceLineItemEntity{
		Price:     float64(10),
		Billed:    neo4jenum.BilledTypeOnce,
		CreatedAt: utils.Now(),
	})

	// prepare grpc client
	calledEventsPlatformToUpdateOpportunity := false
	opportunityCallbacks := mocked_grpc.MockOpportunityServiceCallbacks{
		UpdateOpportunity: func(context context.Context, op *opportunitypb.UpdateOpportunityGrpcRequest) (*opportunitypb.OpportunityIdGrpcResponse, error) {
			require.Equal(t, tenantName, op.Tenant)
			require.Equal(t, opportunityId, op.Id)
			require.Equal(t, float64(0), op.Amount)
			require.Equal(t, float64(0), op.MaxAmount)
			require.Equal(t, []opportunitypb.OpportunityMaskField{opportunitypb.OpportunityMaskField_OPPORTUNITY_PROPERTY_AMOUNT,
				opportunitypb.OpportunityMaskField_OPPORTUNITY_PROPERTY_MAX_AMOUNT}, op.FieldsMask)
			require.Equal(t, constants.AppSourceEventProcessingPlatformSubscribers, op.SourceFields.AppSource)
			require.Equal(t, constants.SourceOpenline, op.SourceFields.Source)
			calledEventsPlatformToUpdateOpportunity = true
			return &opportunitypb.OpportunityIdGrpcResponse{
				Id: opportunityId,
			}, nil
		},
	}
	mocked_grpc.SetOpportunityCallbacks(&opportunityCallbacks)

	// prepare event handler
	handler := NewContractHandler(testLogger, testDatabase.Services, testMockedGrpcClient)

	// EXECUTE
	err := handler.UpdateActiveRenewalOpportunityArr(ctx, tenantName, contractId)
	require.Nil(t, err)

	// Check
	require.True(t, calledEventsPlatformToUpdateOpportunity)
}

func TestContractEventHandler_UpdateRenewalArrForecast_MultipleServices(t *testing.T) {
	ctx := context.Background()
	defer tearDownTestCase(ctx, testDatabase)(t)

	// prepare neo4j data
	neo4jtest.CreateTenant(ctx, testDatabase.Driver, tenantName)
	serviceStartedAt, _ := utils.UnmarshalDateTime("2021-01-01T00:00:00Z")
	orgId := neo4jtest.CreateOrganization(ctx, testDatabase.Driver, tenantName, neo4jentity.OrganizationEntity{})
	contractId := neo4jtest.CreateContractForOrganization(ctx, testDatabase.Driver, tenantName, orgId, neo4jentity.ContractEntity{
		ServiceStartedAt: serviceStartedAt,
		LengthInMonths:   1,
	})
	opportunityId := neo4jtest.CreateOpportunityForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.OpportunityEntity{
		InternalType:  neo4jenum.OpportunityInternalTypeRenewal,
		InternalStage: neo4jenum.OpportunityInternalStageOpen,
		RenewalDetails: neo4jentity.RenewalDetails{
			RenewalLikelihood:   neo4jenum.RenewalLikelihoodHigh,
			RenewalAdjustedRate: 100,
		},
	})
	neo4jtest.CreateServiceLineItemForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.ServiceLineItemEntity{
		Price:     float64(1000000),
		Billed:    neo4jenum.BilledTypeOnce,
		CreatedAt: utils.Now(),
	})
	neo4jtest.CreateServiceLineItemForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.ServiceLineItemEntity{
		Price:     float64(10),
		Quantity:  int64(5),
		Billed:    neo4jenum.BilledTypeMonthly,
		CreatedAt: utils.Now(),
	})
	neo4jtest.CreateServiceLineItemForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.ServiceLineItemEntity{
		Price:     float64(1000),
		Quantity:  int64(2),
		Billed:    neo4jenum.BilledTypeAnnually,
		CreatedAt: utils.Now(),
	})

	// prepare grpc client
	calledEventsPlatformToUpdateOpportunity := false
	opportunityCallbacks := mocked_grpc.MockOpportunityServiceCallbacks{
		UpdateOpportunity: func(context context.Context, op *opportunitypb.UpdateOpportunityGrpcRequest) (*opportunitypb.OpportunityIdGrpcResponse, error) {
			require.Equal(t, tenantName, op.Tenant)
			require.Equal(t, opportunityId, op.Id)
			require.Equal(t, float64(2600), op.Amount)
			require.Equal(t, float64(2600), op.MaxAmount)
			require.Equal(t, []opportunitypb.OpportunityMaskField{opportunitypb.OpportunityMaskField_OPPORTUNITY_PROPERTY_AMOUNT,
				opportunitypb.OpportunityMaskField_OPPORTUNITY_PROPERTY_MAX_AMOUNT}, op.FieldsMask)
			require.Equal(t, constants.AppSourceEventProcessingPlatformSubscribers, op.SourceFields.AppSource)
			require.Equal(t, constants.SourceOpenline, op.SourceFields.Source)
			calledEventsPlatformToUpdateOpportunity = true
			return &opportunitypb.OpportunityIdGrpcResponse{
				Id: opportunityId,
			}, nil
		},
	}
	mocked_grpc.SetOpportunityCallbacks(&opportunityCallbacks)

	// prepare event handler
	handler := NewContractHandler(testLogger, testDatabase.Services, testMockedGrpcClient)

	// EXECUTE
	err := handler.UpdateActiveRenewalOpportunityArr(ctx, tenantName, contractId)
	require.Nil(t, err)

	// Check
	require.True(t, calledEventsPlatformToUpdateOpportunity)
}

func TestContractEventHandler_UpdateRenewalArrForecast_MediumLikelihood(t *testing.T) {
	ctx := context.Background()
	defer tearDownTestCase(ctx, testDatabase)(t)

	// prepare neo4j data
	neo4jtest.CreateTenant(ctx, testDatabase.Driver, tenantName)
	serviceStartedAt, _ := utils.UnmarshalDateTime("2021-01-01T00:00:00Z")
	orgId := neo4jtest.CreateOrganization(ctx, testDatabase.Driver, tenantName, neo4jentity.OrganizationEntity{})
	contractId := neo4jtest.CreateContractForOrganization(ctx, testDatabase.Driver, tenantName, orgId, neo4jentity.ContractEntity{
		ServiceStartedAt: serviceStartedAt,
		LengthInMonths:   1,
	})
	opportunityId := neo4jtest.CreateOpportunityForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.OpportunityEntity{
		InternalType:  neo4jenum.OpportunityInternalTypeRenewal,
		InternalStage: neo4jenum.OpportunityInternalStageOpen,
		RenewalDetails: neo4jentity.RenewalDetails{
			RenewalLikelihood:   neo4jenum.RenewalLikelihoodMedium,
			RenewalAdjustedRate: 50,
		},
	})
	neo4jtest.CreateServiceLineItemForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.ServiceLineItemEntity{
		Price:     float64(1000),
		Quantity:  int64(4),
		Billed:    neo4jenum.BilledTypeAnnually,
		CreatedAt: utils.Now(),
	})

	// prepare grpc client
	calledEventsPlatformToUpdateOpportunity := false
	opportunityCallbacks := mocked_grpc.MockOpportunityServiceCallbacks{
		UpdateOpportunity: func(context context.Context, op *opportunitypb.UpdateOpportunityGrpcRequest) (*opportunitypb.OpportunityIdGrpcResponse, error) {
			require.Equal(t, tenantName, op.Tenant)
			require.Equal(t, opportunityId, op.Id)
			require.Equal(t, float64(2000), op.Amount)
			require.Equal(t, float64(4000), op.MaxAmount)
			require.Equal(t, []opportunitypb.OpportunityMaskField{opportunitypb.OpportunityMaskField_OPPORTUNITY_PROPERTY_AMOUNT,
				opportunitypb.OpportunityMaskField_OPPORTUNITY_PROPERTY_MAX_AMOUNT}, op.FieldsMask)
			require.Equal(t, constants.AppSourceEventProcessingPlatformSubscribers, op.SourceFields.AppSource)
			require.Equal(t, constants.SourceOpenline, op.SourceFields.Source)
			calledEventsPlatformToUpdateOpportunity = true
			return &opportunitypb.OpportunityIdGrpcResponse{
				Id: opportunityId,
			}, nil
		},
	}
	mocked_grpc.SetOpportunityCallbacks(&opportunityCallbacks)

	// prepare event handler
	handler := NewContractHandler(testLogger, testDatabase.Services, testMockedGrpcClient)

	// EXECUTE
	err := handler.UpdateActiveRenewalOpportunityArr(ctx, tenantName, contractId)
	require.Nil(t, err)

	// Check
	require.True(t, calledEventsPlatformToUpdateOpportunity)
}

func TestContractEventHandler_UpdateRenewalArrForecast_ContractEndsBeforeNextRenewal(t *testing.T) {
	ctx := context.Background()
	defer tearDownTestCase(ctx, testDatabase)(t)

	tomorrow := utils.Now().AddDate(0, 0, 1)
	afterTomorrow := utils.Now().AddDate(0, 0, 2)

	// prepare neo4j data
	neo4jtest.CreateTenant(ctx, testDatabase.Driver, tenantName)
	serviceStartedAt, _ := utils.UnmarshalDateTime("2021-01-01T00:00:00Z")
	orgId := neo4jtest.CreateOrganization(ctx, testDatabase.Driver, tenantName, neo4jentity.OrganizationEntity{})
	contractId := neo4jtest.CreateContractForOrganization(ctx, testDatabase.Driver, tenantName, orgId, neo4jentity.ContractEntity{
		ServiceStartedAt: serviceStartedAt,
		LengthInMonths:   12,
		EndedAt:          utils.TimePtr(tomorrow),
	})
	opportunityId := neo4jtest.CreateOpportunityForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.OpportunityEntity{
		InternalType:  neo4jenum.OpportunityInternalTypeRenewal,
		InternalStage: neo4jenum.OpportunityInternalStageOpen,
		RenewalDetails: neo4jentity.RenewalDetails{
			RenewalLikelihood: neo4jenum.RenewalLikelihoodMedium,
			RenewedAt:         utils.TimePtr(afterTomorrow),
		},
	})
	neo4jtest.CreateServiceLineItemForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.ServiceLineItemEntity{
		Price:     float64(1000),
		Quantity:  int64(1),
		Billed:    neo4jenum.BilledTypeAnnually,
		CreatedAt: utils.Now(),
	})

	// prepare grpc client
	calledEventsPlatformToUpdateOpportunity := false
	opportunityCallbacks := mocked_grpc.MockOpportunityServiceCallbacks{
		UpdateOpportunity: func(context context.Context, op *opportunitypb.UpdateOpportunityGrpcRequest) (*opportunitypb.OpportunityIdGrpcResponse, error) {
			require.Equal(t, tenantName, op.Tenant)
			require.Equal(t, opportunityId, op.Id)
			require.Equal(t, float64(0), op.Amount)
			require.Equal(t, float64(0), op.MaxAmount)
			require.Equal(t, []opportunitypb.OpportunityMaskField{opportunitypb.OpportunityMaskField_OPPORTUNITY_PROPERTY_AMOUNT,
				opportunitypb.OpportunityMaskField_OPPORTUNITY_PROPERTY_MAX_AMOUNT}, op.FieldsMask)
			require.Equal(t, constants.AppSourceEventProcessingPlatformSubscribers, op.SourceFields.AppSource)
			require.Equal(t, constants.SourceOpenline, op.SourceFields.Source)
			calledEventsPlatformToUpdateOpportunity = true
			return &opportunitypb.OpportunityIdGrpcResponse{
				Id: opportunityId,
			}, nil
		},
	}
	mocked_grpc.SetOpportunityCallbacks(&opportunityCallbacks)

	// prepare event handler
	handler := NewContractHandler(testLogger, testDatabase.Services, testMockedGrpcClient)

	// EXECUTE
	err := handler.UpdateActiveRenewalOpportunityArr(ctx, tenantName, contractId)
	require.Nil(t, err)

	// Check
	require.True(t, calledEventsPlatformToUpdateOpportunity)
}

func TestContractEventHandler_UpdateRenewalArrForecast_ContractEndsIn6Months_ProrateAmount(t *testing.T) {
	ctx := context.Background()
	defer tearDownTestCase(ctx, testDatabase)(t)

	in6Months := utils.Now().AddDate(0, 6, 0)
	in1Month := utils.Now().AddDate(0, 1, 0)

	// prepare neo4j data
	neo4jtest.CreateTenant(ctx, testDatabase.Driver, tenantName)
	serviceStartedAt, _ := utils.UnmarshalDateTime("2021-01-01T00:00:00Z")
	orgId := neo4jtest.CreateOrganization(ctx, testDatabase.Driver, tenantName, neo4jentity.OrganizationEntity{})
	contractId := neo4jtest.CreateContractForOrganization(ctx, testDatabase.Driver, tenantName, orgId, neo4jentity.ContractEntity{
		ServiceStartedAt: serviceStartedAt,
		LengthInMonths:   1,
		EndedAt:          utils.TimePtr(in6Months),
	})
	opportunityId := neo4jtest.CreateOpportunityForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.OpportunityEntity{
		InternalType:  neo4jenum.OpportunityInternalTypeRenewal,
		InternalStage: neo4jenum.OpportunityInternalStageOpen,
		RenewalDetails: neo4jentity.RenewalDetails{
			RenewalLikelihood:   neo4jenum.RenewalLikelihoodHigh,
			RenewedAt:           utils.TimePtr(in1Month),
			RenewalAdjustedRate: 100,
		},
	})
	neo4jtest.CreateServiceLineItemForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.ServiceLineItemEntity{
		Price:     float64(1000),
		Quantity:  int64(1),
		Billed:    neo4jenum.BilledTypeAnnually,
		CreatedAt: utils.Now(),
	})

	// prepare grpc client
	calledEventsPlatformToUpdateOpportunity := false
	opportunityCallbacks := mocked_grpc.MockOpportunityServiceCallbacks{
		UpdateOpportunity: func(context context.Context, op *opportunitypb.UpdateOpportunityGrpcRequest) (*opportunitypb.OpportunityIdGrpcResponse, error) {
			require.Equal(t, tenantName, op.Tenant)
			require.Equal(t, opportunityId, op.Id)
			require.Equal(t, float64(500), op.Amount)
			require.Equal(t, float64(500), op.MaxAmount)
			require.Equal(t, []opportunitypb.OpportunityMaskField{opportunitypb.OpportunityMaskField_OPPORTUNITY_PROPERTY_AMOUNT,
				opportunitypb.OpportunityMaskField_OPPORTUNITY_PROPERTY_MAX_AMOUNT}, op.FieldsMask)
			require.Equal(t, constants.AppSourceEventProcessingPlatformSubscribers, op.SourceFields.AppSource)
			require.Equal(t, constants.SourceOpenline, op.SourceFields.Source)
			calledEventsPlatformToUpdateOpportunity = true
			return &opportunitypb.OpportunityIdGrpcResponse{
				Id: opportunityId,
			}, nil
		},
	}
	mocked_grpc.SetOpportunityCallbacks(&opportunityCallbacks)

	// prepare event handler
	handler := NewContractHandler(testLogger, testDatabase.Services, testMockedGrpcClient)

	// EXECUTE
	err := handler.UpdateActiveRenewalOpportunityArr(ctx, tenantName, contractId)
	require.Nil(t, err)

	// Check
	require.True(t, calledEventsPlatformToUpdateOpportunity)
}

func TestContractEventHandler_UpdateRenewalArrForecast_ContractEndsInMoreThan12Months_FullAmount(t *testing.T) {
	ctx := context.Background()
	defer tearDownTestCase(ctx, testDatabase)(t)

	in13Months := utils.Now().AddDate(0, 13, 0)
	in12Months := utils.Now().AddDate(1, 0, 0)

	// prepare neo4j data
	neo4jtest.CreateTenant(ctx, testDatabase.Driver, tenantName)
	serviceStartedAt, _ := utils.UnmarshalDateTime("2021-01-01T00:00:00Z")
	orgId := neo4jtest.CreateOrganization(ctx, testDatabase.Driver, tenantName, neo4jentity.OrganizationEntity{})
	contractId := neo4jtest.CreateContractForOrganization(ctx, testDatabase.Driver, tenantName, orgId, neo4jentity.ContractEntity{
		ServiceStartedAt: serviceStartedAt,
		LengthInMonths:   1,
		EndedAt:          utils.TimePtr(in13Months),
	})
	opportunityId := neo4jtest.CreateOpportunityForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.OpportunityEntity{
		InternalType:  neo4jenum.OpportunityInternalTypeRenewal,
		InternalStage: neo4jenum.OpportunityInternalStageOpen,
		RenewalDetails: neo4jentity.RenewalDetails{
			RenewalLikelihood:   neo4jenum.RenewalLikelihoodHigh,
			RenewedAt:           utils.TimePtr(in12Months),
			RenewalAdjustedRate: 100,
		},
	})
	neo4jtest.CreateServiceLineItemForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.ServiceLineItemEntity{
		Price:     float64(1000),
		Quantity:  int64(1),
		Billed:    neo4jenum.BilledTypeAnnually,
		CreatedAt: utils.Now(),
	})

	// prepare grpc client
	calledEventsPlatformToUpdateOpportunity := false
	opportunityCallbacks := mocked_grpc.MockOpportunityServiceCallbacks{
		UpdateOpportunity: func(context context.Context, op *opportunitypb.UpdateOpportunityGrpcRequest) (*opportunitypb.OpportunityIdGrpcResponse, error) {
			require.Equal(t, tenantName, op.Tenant)
			require.Equal(t, opportunityId, op.Id)
			require.Equal(t, float64(1000), op.Amount)
			require.Equal(t, float64(1000), op.MaxAmount)
			require.Equal(t, []opportunitypb.OpportunityMaskField{opportunitypb.OpportunityMaskField_OPPORTUNITY_PROPERTY_AMOUNT,
				opportunitypb.OpportunityMaskField_OPPORTUNITY_PROPERTY_MAX_AMOUNT}, op.FieldsMask)
			require.Equal(t, constants.AppSourceEventProcessingPlatformSubscribers, op.SourceFields.AppSource)
			require.Equal(t, constants.SourceOpenline, op.SourceFields.Source)
			calledEventsPlatformToUpdateOpportunity = true
			return &opportunitypb.OpportunityIdGrpcResponse{
				Id: opportunityId,
			}, nil
		},
	}
	mocked_grpc.SetOpportunityCallbacks(&opportunityCallbacks)

	// prepare event handler
	handler := NewContractHandler(testLogger, testDatabase.Services, testMockedGrpcClient)

	// EXECUTE
	err := handler.UpdateActiveRenewalOpportunityArr(ctx, tenantName, contractId)
	require.Nil(t, err)

	// Check
	require.True(t, calledEventsPlatformToUpdateOpportunity)
}

func TestContractEventHandler_UpdateActiveRenewalOpportunityLikelihood_EndedContract_UpdateLikelihood(t *testing.T) {
	ctx := context.Background()
	defer tearDownTestCase(ctx, testDatabase)(t)

	// prepare neo4j data
	neo4jtest.CreateTenant(ctx, testDatabase.Driver, tenantName)
	tomorrow := utils.Now().AddDate(0, 0, 1)
	afterTomorrow := utils.Now().AddDate(0, 0, 2)
	orgId := neo4jtest.CreateOrganization(ctx, testDatabase.Driver, tenantName, neo4jentity.OrganizationEntity{})
	contractId := neo4jtest.CreateContractForOrganization(ctx, testDatabase.Driver, tenantName, orgId, neo4jentity.ContractEntity{
		EndedAt:        &tomorrow,
		LengthInMonths: 12,
	})
	opportunityId := neo4jtest.CreateOpportunityForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.OpportunityEntity{
		InternalType:  neo4jenum.OpportunityInternalTypeRenewal,
		InternalStage: neo4jenum.OpportunityInternalStageOpen,
		RenewalDetails: neo4jentity.RenewalDetails{
			RenewalLikelihood:   neo4jenum.RenewalLikelihoodLow,
			RenewedAt:           &afterTomorrow,
			RenewalAdjustedRate: 25,
		},
	})

	// prepare grpc mock
	calledEventsPlatformToUpdateRenewalOpportunity := false
	opportunityCallbacks := mocked_grpc.MockOpportunityServiceCallbacks{
		UpdateRenewalOpportunity: func(context context.Context, op *opportunitypb.UpdateRenewalOpportunityGrpcRequest) (*opportunitypb.OpportunityIdGrpcResponse, error) {
			require.Equal(t, tenantName, op.Tenant)
			require.Equal(t, opportunityId, op.Id)
			require.Equal(t, constants.AppSourceEventProcessingPlatformSubscribers, op.SourceFields.AppSource)
			require.Equal(t, opportunitypb.RenewalLikelihood_ZERO_RENEWAL, op.RenewalLikelihood)
			require.Equal(t, int64(0), op.RenewalAdjustedRate)
			require.Equal(t, []opportunitypb.OpportunityMaskField{
				opportunitypb.OpportunityMaskField_OPPORTUNITY_PROPERTY_RENEWAL_LIKELIHOOD,
				opportunitypb.OpportunityMaskField_OPPORTUNITY_PROPERTY_ADJUSTED_RATE}, op.FieldsMask)
			calledEventsPlatformToUpdateRenewalOpportunity = true
			return &opportunitypb.OpportunityIdGrpcResponse{
				Id: opportunityId,
			}, nil
		},
	}
	mocked_grpc.SetOpportunityCallbacks(&opportunityCallbacks)

	// prepare event handler
	handler := NewContractHandler(testLogger, testDatabase.Services, testMockedGrpcClient)

	// EXECUTE
	err := handler.UpdateActiveRenewalOpportunityLikelihood(ctx, tenantName, contractId)
	require.Nil(t, err)

	// Check
	require.True(t, calledEventsPlatformToUpdateRenewalOpportunity)
}

func TestContractEventHandler_UpdateActiveRenewalOpportunityLikelihood_EndedContract_LikelihoodAlreadyZero(t *testing.T) {
	ctx := context.Background()
	defer tearDownTestCase(ctx, testDatabase)(t)

	// prepare neo4j data
	neo4jtest.CreateTenant(ctx, testDatabase.Driver, tenantName)
	tomorrow := utils.Now().AddDate(0, 0, 1)
	afterTomorrow := utils.Now().AddDate(0, 0, 2)
	orgId := neo4jtest.CreateOrganization(ctx, testDatabase.Driver, tenantName, neo4jentity.OrganizationEntity{})
	contractId := neo4jtest.CreateContractForOrganization(ctx, testDatabase.Driver, tenantName, orgId, neo4jentity.ContractEntity{
		EndedAt:        &tomorrow,
		LengthInMonths: 12,
	})
	neo4jtest.CreateOpportunityForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.OpportunityEntity{
		InternalType:  neo4jenum.OpportunityInternalTypeRenewal,
		InternalStage: neo4jenum.OpportunityInternalStageOpen,
		RenewalDetails: neo4jentity.RenewalDetails{
			RenewalLikelihood:   neo4jenum.RenewalLikelihoodZero,
			RenewedAt:           &afterTomorrow,
			RenewalAdjustedRate: 0,
		},
	})

	// prepare event handler
	handler := NewContractHandler(testLogger, testDatabase.Services, testMockedGrpcClient)

	// EXECUTE
	err := handler.UpdateActiveRenewalOpportunityLikelihood(ctx, tenantName, contractId)
	require.Nil(t, err)
}

func TestContractEventHandler_UpdateActiveRenewalOpportunityLikelihood_ReinitiatedContract_UpdateLikelihood(t *testing.T) {
	ctx := context.Background()
	defer tearDownTestCase(ctx, testDatabase)(t)

	// prepare neo4j data
	neo4jtest.CreateTenant(ctx, testDatabase.Driver, tenantName)
	afterTomorrow := utils.Now().AddDate(0, 0, 2)
	orgId := neo4jtest.CreateOrganization(ctx, testDatabase.Driver, tenantName, neo4jentity.OrganizationEntity{})
	contractId := neo4jtest.CreateContractForOrganization(ctx, testDatabase.Driver, tenantName, orgId, neo4jentity.ContractEntity{
		LengthInMonths: 12,
	})
	opportunityId := neo4jtest.CreateOpportunityForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.OpportunityEntity{
		InternalType:  neo4jenum.OpportunityInternalTypeRenewal,
		InternalStage: neo4jenum.OpportunityInternalStageOpen,
		RenewalDetails: neo4jentity.RenewalDetails{
			RenewalLikelihood:   neo4jenum.RenewalLikelihoodZero,
			RenewedAt:           &afterTomorrow,
			RenewalAdjustedRate: 0,
		},
	})

	// prepare grpc mock
	calledEventsPlatformToUpdateRenewalOpportunity := false
	opportunityCallbacks := mocked_grpc.MockOpportunityServiceCallbacks{
		UpdateRenewalOpportunity: func(context context.Context, op *opportunitypb.UpdateRenewalOpportunityGrpcRequest) (*opportunitypb.OpportunityIdGrpcResponse, error) {
			require.Equal(t, tenantName, op.Tenant)
			require.Equal(t, opportunityId, op.Id)
			require.Equal(t, constants.AppSourceEventProcessingPlatformSubscribers, op.SourceFields.AppSource)
			require.Equal(t, opportunitypb.RenewalLikelihood_MEDIUM_RENEWAL, op.RenewalLikelihood)
			require.Equal(t, int64(50), op.RenewalAdjustedRate)
			require.Equal(t, []opportunitypb.OpportunityMaskField{
				opportunitypb.OpportunityMaskField_OPPORTUNITY_PROPERTY_RENEWAL_LIKELIHOOD,
				opportunitypb.OpportunityMaskField_OPPORTUNITY_PROPERTY_ADJUSTED_RATE,
			}, op.FieldsMask)
			calledEventsPlatformToUpdateRenewalOpportunity = true
			return &opportunitypb.OpportunityIdGrpcResponse{
				Id: opportunityId,
			}, nil
		},
	}
	mocked_grpc.SetOpportunityCallbacks(&opportunityCallbacks)

	// prepare event handler
	handler := NewContractHandler(testLogger, testDatabase.Services, testMockedGrpcClient)

	// EXECUTE
	err := handler.UpdateActiveRenewalOpportunityLikelihood(ctx, tenantName, contractId)
	require.Nil(t, err)

	// Check
	require.True(t, calledEventsPlatformToUpdateRenewalOpportunity)
}

func TestContractEventHandler_UpdateActiveRenewalOpportunityLikelihood_ReinitiatedContract_LikelihoodNotZero(t *testing.T) {
	ctx := context.Background()
	defer tearDownTestCase(ctx, testDatabase)(t)

	// prepare neo4j data
	neo4jtest.CreateTenant(ctx, testDatabase.Driver, tenantName)
	afterTomorrow := utils.Now().AddDate(0, 0, 2)
	orgId := neo4jtest.CreateOrganization(ctx, testDatabase.Driver, tenantName, neo4jentity.OrganizationEntity{})
	contractId := neo4jtest.CreateContractForOrganization(ctx, testDatabase.Driver, tenantName, orgId, neo4jentity.ContractEntity{
		LengthInMonths: 12,
	})
	neo4jtest.CreateOpportunityForContract(ctx, testDatabase.Driver, tenantName, contractId, neo4jentity.OpportunityEntity{
		InternalType:  neo4jenum.OpportunityInternalTypeRenewal,
		InternalStage: neo4jenum.OpportunityInternalStageOpen,
		RenewalDetails: neo4jentity.RenewalDetails{
			RenewalLikelihood:   neo4jenum.RenewalLikelihoodHigh,
			RenewedAt:           &afterTomorrow,
			RenewalAdjustedRate: 100,
		},
	})

	// prepare event handler
	handler := NewContractHandler(testLogger, testDatabase.Services, testMockedGrpcClient)

	// EXECUTE
	err := handler.UpdateActiveRenewalOpportunityLikelihood(ctx, tenantName, contractId)
	require.Nil(t, err)
}
