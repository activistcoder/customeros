package service

import (
	"context"
	"fmt"
	"github.com/machinebox/graphql"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-api-sdk/graph/model"

	"time"
)

type CustomerOSApiClient interface {
	CreateExternalSystem(tenant string, username *string, input model.ExternalSystemInput) (string, error)
}

type customerOSApiClient struct {
	customerOSApiKey string
	graphqlClient    *graphql.Client
}

func NewCustomerOsClient(customerOSApiPath, customerOSApiKey string) CustomerOSApiClient {
	return &customerOSApiClient{
		customerOSApiKey: customerOSApiKey,
		graphqlClient:    graphql.NewClient(customerOSApiPath),
	}
}

func (s *customerOSApiClient) CreateExternalSystem(tenant string, username *string, input model.ExternalSystemInput) (string, error) {
	graphqlRequest := graphql.NewRequest(
		`mutation ExternalSystemCreate($input: ExternalSystemInput!) {
				externalSystem_Create(input: $input)
				}`)
	graphqlRequest.Var("input", input)

	err := s.addHeadersToGraphRequest(graphqlRequest, &tenant, username)
	if err != nil {
		return "", err
	}
	ctx, cancel, err := s.contextWithTimeout()
	if err != nil {
		return "", err
	}
	defer cancel()

	var graphqlResponse map[string]string
	if err := s.graphqlClient.Run(ctx, graphqlRequest, &graphqlResponse); err != nil {
		return "", fmt.Errorf("externalSystem_Create: %w", err)
	}
	return graphqlResponse["externalSystem_Create"], nil
}

func (s *customerOSApiClient) addHeadersToGraphRequest(req *graphql.Request, tenant, username *string) error {
	req.Header.Add("X-Openline-API-KEY", s.customerOSApiKey)

	if tenant != nil {
		req.Header.Add("X-Openline-TENANT", *tenant)
	}

	if username != nil {
		req.Header.Add("X-Openline-USERNAME", *username)
	}

	return nil
}

func (s *customerOSApiClient) contextWithTimeout() (context.Context, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	return ctx, cancel, nil
}