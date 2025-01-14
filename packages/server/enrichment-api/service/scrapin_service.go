package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/tracing"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/utils"
	postgresentity "github.com/openline-ai/openline-customer-os/packages/server/customer-os-postgres-repository/entity"
	"github.com/openline-ai/openline-customer-os/packages/server/enrichment-api/config"
	"github.com/openline-ai/openline-customer-os/packages/server/enrichment-api/logger"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"time"
)

type ScrapinService interface {
	ScrapInPersonProfile(ctx context.Context, linkedInUrl string) (uint64, *postgresentity.ScrapInResponseBody, error)
	ScrapInSearchPerson(ctx context.Context, email, fistName, lastName, domain string) (uint64, *postgresentity.ScrapInResponseBody, error)
	ScrapInCompanyProfile(ctx context.Context, linkedInUrl string) (uint64, *postgresentity.ScrapInResponseBody, error)
	ScrapInSearchCompany(ctx context.Context, domain string) (uint64, *postgresentity.ScrapInResponseBody, error)
}

type ScrapInSearchRequestParams struct {
	FirstName     string `json:"firstName,omitempty"`
	LastName      string `json:"lastName,omitempty"`
	CompanyDomain string `json:"companyDomain,omitempty"`
	Email         string `json:"email,omitempty"`
	LinkedInUrl   string `json:"linkedInUrl,omitempty"`
}

type scrapinService struct {
	config   *config.Config
	services *Services
	log      logger.Logger
}

func NewScrapeInService(config *config.Config, services *Services, log logger.Logger) ScrapinService {
	return &scrapinService{
		config:   config,
		services: services,
		log:      log,
	}
}

func (s *scrapinService) ScrapInPersonProfile(ctx context.Context, linkedInUrl string) (uint64, *postgresentity.ScrapInResponseBody, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ScrapinService.ScrapInPersonProfile")
	defer span.Finish()
	span.LogFields(log.String("linkedInUrl", linkedInUrl))

	latestEnrichDetailsScrapInRecord, err := s.services.CommonServices.PostgresRepositories.EnrichDetailsScrapInRepository.GetLatestByParam1AndFlow(ctx, linkedInUrl, postgresentity.ScrapInFlowPersonProfile)
	if err != nil {
		tracing.TraceErr(span, errors.Wrap(err, "failed to get scrapin data"))
		return 0, nil, err
	}

	var data *postgresentity.ScrapInResponseBody
	var recordId uint64

	callScrapInNow := false

	if latestEnrichDetailsScrapInRecord == nil || latestEnrichDetailsScrapInRecord.UpdatedAt.AddDate(0, 0, s.config.ScrapinConfig.TtlDays).Before(utils.Now()) {
		// if no cached data found, or cache data is older than ttl (90 days), call scrapin
		callScrapInNow = true
	} else if latestEnrichDetailsScrapInRecord.PersonFound == false {
		// if last attempt was > 1 day ago, call scrapin
		if latestEnrichDetailsScrapInRecord.UpdatedAt.AddDate(0, 0, 1).Before(utils.Now()) {
			callScrapInNow = true
		}
	}

	if callScrapInNow {
		// get data from scrapin
		if data, err = s.callScrapinPersonProfile(ctx, linkedInUrl); err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to call scrapin"))
			return 0, nil, err
		}

		traceIfCreditsDepleting(data, span)

		// save to db
		dataAsString, err := json.Marshal(data)
		if err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to marshal data"))
			return 0, nil, err
		}
		paramsAsString, err := json.Marshal(ScrapInSearchRequestParams{
			LinkedInUrl: linkedInUrl,
		})
		if err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to marshal params"))
		}
		dbRecord, err := s.services.CommonServices.PostgresRepositories.EnrichDetailsScrapInRepository.Create(ctx, postgresentity.EnrichDetailsScrapIn{
			Param1:        linkedInUrl,
			Flow:          postgresentity.ScrapInFlowPersonProfile,
			AllParamsJson: string(paramsAsString),
			Data:          string(dataAsString),
			PersonFound:   data.Person != nil,
			CompanyFound:  data.Company != nil,
			Success:       data.Success,
		})
		if err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to save scrapin data in db"))
		}
		if dbRecord != nil {
			recordId = dbRecord.ID
		}
	} else {
		recordId = latestEnrichDetailsScrapInRecord.ID
		// unmarshal cached data
		unmarshalledData := postgresentity.ScrapInResponseBody{}
		if err = json.Unmarshal([]byte(latestEnrichDetailsScrapInRecord.Data), &unmarshalledData); err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to unmarshal scrapin cached data"))
			return 0, nil, err
		}
		data = &unmarshalledData
	}

	// if fresh data not found, check most recent cached data with person found
	if data == nil || data.Person == nil {
		latestEnrichDetailsScrapInRecordWithPersonFound, err := s.services.CommonServices.PostgresRepositories.EnrichDetailsScrapInRepository.GetLatestByParam1AndFlowWithPersonFound(ctx, linkedInUrl, postgresentity.ScrapInFlowPersonProfile)
		if err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to extract cached scrapin data"))
			return 0, nil, err
		}
		if latestEnrichDetailsScrapInRecordWithPersonFound != nil {
			unmarshalledData := postgresentity.ScrapInResponseBody{}
			if err = json.Unmarshal([]byte(latestEnrichDetailsScrapInRecordWithPersonFound.Data), &unmarshalledData); err != nil {
				tracing.TraceErr(span, errors.Wrap(err, "failed to unmarshal scrapin cached data"))
				return 0, nil, err
			}
			data = &unmarshalledData
			return latestEnrichDetailsScrapInRecordWithPersonFound.ID, data, nil
		}
	}

	return recordId, data, nil
}

func (s *scrapinService) callScrapinPersonProfile(ctx context.Context, linkedInUrl string) (*postgresentity.ScrapInResponseBody, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ScrapInService.callScrapinPersonProfile")
	defer span.Finish()
	span.LogFields(log.String("linkedInUrl", linkedInUrl))

	baseUrl := s.config.ScrapinConfig.Url
	if baseUrl == "" {
		err := errors.New("ScrapIn URL not set")
		tracing.TraceErr(span, err)
		s.log.Errorf("ScrapIn URL not set")
		return &postgresentity.ScrapInResponseBody{}, err
	}
	scrapInApiKey := s.config.ScrapinConfig.ApiKey
	if scrapInApiKey == "" {
		err := errors.New("Scrapin Api key not set")
		tracing.TraceErr(span, err)
		s.log.Errorf("Scrapin Api key not set")
		return &postgresentity.ScrapInResponseBody{}, err
	}

	params := url.Values{}
	params.Add("apikey", scrapInApiKey)
	params.Add("linkedInUrl", linkedInUrl)

	scrapinStatusCode, body, err := makeScrapInHTTPRequest(baseUrl + "/enrichment/profile" + "?" + params.Encode())
	if err != nil {
		tracing.TraceErr(span, errors.Wrap(err, "makeScrapInHTTPRequest"))
		s.log.Errorf("Error making scrapin HTTP request: %s", err.Error())
		return &postgresentity.ScrapInResponseBody{}, err
	}
	span.LogFields(log.Int("scrapin.statusCode", scrapinStatusCode))

	var scrapinResponse postgresentity.ScrapInResponseBody
	err = json.Unmarshal(body, &scrapinResponse)
	if err != nil {
		tracing.TraceErr(span, errors.Wrap(err, "json.Unmarshal"))
		span.LogFields(log.String("response.body", string(body)))
		s.log.Errorf("Error unmarshalling scrapin response: %s", err.Error())
		return &postgresentity.ScrapInResponseBody{}, err
	}

	return &scrapinResponse, nil
}

func (s *scrapinService) ScrapInSearchPerson(ctx context.Context, email, firstName, lastName, domain string) (uint64, *postgresentity.ScrapInResponseBody, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ScrapinService.ScrapInSearchPerson")
	defer span.Finish()
	span.LogFields(log.String("email", email), log.String("firstName", firstName), log.String("lastName", lastName), log.String("domain", domain))

	latestEnrichDetailsScrapInRecord, err := s.services.CommonServices.PostgresRepositories.EnrichDetailsScrapInRepository.GetLatestByAllParamsAndFlow(ctx, email, firstName, lastName, domain, postgresentity.ScrapInFlowPersonSearch)
	if err != nil {
		tracing.TraceErr(span, errors.Wrap(err, "failed to get scrapin data"))
		return 0, nil, err
	}

	var data *postgresentity.ScrapInResponseBody
	var recordId uint64

	callScrapInNow := false

	if latestEnrichDetailsScrapInRecord == nil || latestEnrichDetailsScrapInRecord.UpdatedAt.AddDate(0, 0, s.config.ScrapinConfig.TtlDays).Before(utils.Now()) {
		callScrapInNow = true
	} else if latestEnrichDetailsScrapInRecord.PersonFound == false {
		// if last attempt was > 1 day ago, call scrapin
		if latestEnrichDetailsScrapInRecord.UpdatedAt.AddDate(0, 0, 1).Before(utils.Now()) {
			callScrapInNow = true
		}
	}

	// if cached data is missing or last time fetched > ttl refresh
	if callScrapInNow {
		// get data from scrapin
		if data, err = s.callScrapinPersonSearch(ctx, email, firstName, lastName, domain); err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to call scrapin"))
			return 0, nil, err
		}

		traceIfCreditsDepleting(data, span)

		// save to db
		dataAsString, err := json.Marshal(data)
		if err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to marshal data"))
			return 0, nil, err
		}
		paramsAsString, err := json.Marshal(ScrapInSearchRequestParams{
			Email:         email,
			FirstName:     firstName,
			LastName:      lastName,
			CompanyDomain: domain,
		})
		if err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to marshal params"))
		}
		dbRecord, err := s.services.CommonServices.PostgresRepositories.EnrichDetailsScrapInRepository.Create(ctx, postgresentity.EnrichDetailsScrapIn{
			Param1:        email,
			Param2:        firstName,
			Param3:        lastName,
			Param4:        domain,
			Flow:          postgresentity.ScrapInFlowPersonSearch,
			AllParamsJson: string(paramsAsString),
			Data:          string(dataAsString),
			PersonFound:   data.Person != nil,
			CompanyFound:  data.Company != nil,
			Success:       data.Success,
		})
		if err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to save scrapin data in db"))
		}
		if dbRecord != nil {
			recordId = dbRecord.ID
		}
	} else {
		recordId = latestEnrichDetailsScrapInRecord.ID
		// unmarshal cached data
		unmarshalledData := postgresentity.ScrapInResponseBody{}
		if err = json.Unmarshal([]byte(latestEnrichDetailsScrapInRecord.Data), &unmarshalledData); err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to unmarshal scrapin cached data"))
			return 0, nil, err
		}
		data = &unmarshalledData
	}

	// if fresh data not found, check most recent cached data with person found
	if data == nil || data.Person == nil {
		latestEnrichDetailsScrapInRecordWithPersonFound, err := s.services.CommonServices.PostgresRepositories.EnrichDetailsScrapInRepository.GetLatestByAllParamsAndFlowWithPersonFound(ctx, email, firstName, lastName, domain, postgresentity.ScrapInFlowPersonSearch)
		if err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to get scrapin data"))
			return 0, nil, err
		}
		if latestEnrichDetailsScrapInRecordWithPersonFound != nil {
			unmarshalledData := postgresentity.ScrapInResponseBody{}
			if err = json.Unmarshal([]byte(latestEnrichDetailsScrapInRecordWithPersonFound.Data), &unmarshalledData); err != nil {
				tracing.TraceErr(span, errors.Wrap(err, "failed to unmarshal scrapin cached data"))
				return 0, nil, err
			}
			data = &unmarshalledData
			return latestEnrichDetailsScrapInRecordWithPersonFound.ID, data, nil
		}
	}

	return recordId, data, nil
}

func (s *scrapinService) callScrapinPersonSearch(ctx context.Context, email, firstName, lastName, domain string) (*postgresentity.ScrapInResponseBody, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ScrapInService.callScrapinPersonSearch")
	defer span.Finish()
	span.LogFields(log.String("email", email), log.String("firstName", firstName), log.String("lastName", lastName), log.String("domain", domain))

	baseUrl := s.config.ScrapinConfig.Url
	if baseUrl == "" {
		err := errors.New("ScrapIn URL not set")
		tracing.TraceErr(span, err)
		s.log.Errorf("ScrapIn URL not set")
		return &postgresentity.ScrapInResponseBody{}, err
	}
	scrapInApiKey := s.config.ScrapinConfig.ApiKey
	if scrapInApiKey == "" {
		err := errors.New("Scrapin Api key not set")
		tracing.TraceErr(span, err)
		s.log.Errorf("Scrapin Api key not set")
		return &postgresentity.ScrapInResponseBody{}, err
	}

	params := url.Values{}
	params.Add("apikey", scrapInApiKey)
	params.Add("email", email)
	if firstName != "" {
		params.Add("firstName", firstName)
	}
	if lastName != "" {
		params.Add("lastName", lastName)
	}
	if domain != "" {
		params.Add("companyDomain", domain)
	}

	scrapinStatusCode, body, err := makeScrapInHTTPRequest(baseUrl + "/enrichment" + "?" + params.Encode())
	if err != nil {
		tracing.TraceErr(span, errors.Wrap(err, "makeScrapInHTTPRequest"))
		s.log.Errorf("Error making scrapin HTTP request: %s", err.Error())
		return &postgresentity.ScrapInResponseBody{}, err
	}
	span.LogFields(log.Int("scrapin.statusCode", scrapinStatusCode))

	var scrapinResponse postgresentity.ScrapInResponseBody
	err = json.Unmarshal(body, &scrapinResponse)
	if err != nil {
		tracing.TraceErr(span, errors.Wrap(err, "json.Unmarshal"))
		span.LogFields(log.String("response.body", string(body)))
		s.log.Errorf("Error unmarshalling scrapin response: %s", err.Error())
		return &postgresentity.ScrapInResponseBody{}, err
	}

	return &scrapinResponse, nil
}

func (s *scrapinService) ScrapInCompanyProfile(ctx context.Context, linkedInUrl string) (uint64, *postgresentity.ScrapInResponseBody, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ScrapinService.ScrapInCompanyProfile")
	defer span.Finish()
	span.LogFields(log.String("linkedInUrl", linkedInUrl))

	latestEnrichDetailsScrapInRecord, err := s.services.CommonServices.PostgresRepositories.EnrichDetailsScrapInRepository.GetLatestByParam1AndFlow(ctx, linkedInUrl, postgresentity.ScrapInFlowCompanyProfile)
	if err != nil {
		tracing.TraceErr(span, errors.Wrap(err, "failed to get scrapin data"))
		return 0, nil, err
	}

	var data *postgresentity.ScrapInResponseBody
	var recordId uint64

	callScrapInNow := false

	if latestEnrichDetailsScrapInRecord == nil || latestEnrichDetailsScrapInRecord.UpdatedAt.AddDate(0, 0, s.config.ScrapinConfig.TtlDays).Before(utils.Now()) {
		callScrapInNow = true
	} else if latestEnrichDetailsScrapInRecord.CompanyFound == false {
		// if last attempt was > 1 day ago, call scrapin
		if latestEnrichDetailsScrapInRecord.UpdatedAt.AddDate(0, 0, 1).Before(utils.Now()) {
			callScrapInNow = true
		}
	}

	// if cached data is missing or last time fetched > ttl refresh
	if callScrapInNow {
		// get data from scrapin
		if data, err = s.callScrapinCompanyProfile(ctx, linkedInUrl); err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to call scrapin"))
			return 0, nil, err
		}

		traceIfCreditsDepleting(data, span)

		// save to db
		dataAsString, err := json.Marshal(data)
		if err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to marshal data"))
			return 0, nil, err
		}
		paramsAsString, err := json.Marshal(ScrapInSearchRequestParams{
			LinkedInUrl: linkedInUrl,
		})
		if err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to marshal params"))
		}
		dbRecord, err := s.services.CommonServices.PostgresRepositories.EnrichDetailsScrapInRepository.Create(ctx, postgresentity.EnrichDetailsScrapIn{
			Param1:        linkedInUrl,
			Flow:          postgresentity.ScrapInFlowCompanyProfile,
			AllParamsJson: string(paramsAsString),
			Data:          string(dataAsString),
			PersonFound:   false,
			CompanyFound:  data.Company != nil,
			Success:       data.Success,
		})
		if err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to save scrapin data in db"))
		}
		if dbRecord != nil {
			recordId = dbRecord.ID
		}
	} else {
		recordId = latestEnrichDetailsScrapInRecord.ID
		// unmarshal cached data
		unmarshalledData := postgresentity.ScrapInResponseBody{}
		if err = json.Unmarshal([]byte(latestEnrichDetailsScrapInRecord.Data), &unmarshalledData); err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to unmarshal scrapin cached data"))
			return 0, nil, err
		}
		data = &unmarshalledData
	}

	// if fresh data not found, check most recent cached data with company found
	if data == nil || data.Company == nil {
		latestEnrichDetailsScrapInRecordWithCompanyFound, err := s.services.CommonServices.PostgresRepositories.EnrichDetailsScrapInRepository.GetLatestByParam1AndFlowWithCompanyFound(ctx, linkedInUrl, postgresentity.ScrapInFlowCompanyProfile)
		if err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to get scrapin data"))
			return 0, nil, err
		}
		if latestEnrichDetailsScrapInRecordWithCompanyFound != nil {
			unmarshalledData := postgresentity.ScrapInResponseBody{}
			if err = json.Unmarshal([]byte(latestEnrichDetailsScrapInRecordWithCompanyFound.Data), &unmarshalledData); err != nil {
				tracing.TraceErr(span, errors.Wrap(err, "failed to unmarshal scrapin cached data"))
				return 0, nil, err
			}
			data = &unmarshalledData
			return latestEnrichDetailsScrapInRecordWithCompanyFound.ID, data, nil
		}
	}

	return recordId, data, nil
}

func (s *scrapinService) callScrapinCompanyProfile(ctx context.Context, linkedInUrl string) (*postgresentity.ScrapInResponseBody, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ScrapInService.callScrapinCompanyProfile")
	defer span.Finish()
	span.LogFields(log.String("linkedInUrl", linkedInUrl))

	baseUrl := s.config.ScrapinConfig.Url
	if baseUrl == "" {
		err := errors.New("ScrapIn URL not set")
		tracing.TraceErr(span, err)
		s.log.Errorf("ScrapIn URL not set")
		return &postgresentity.ScrapInResponseBody{}, err
	}
	scrapInApiKey := s.config.ScrapinConfig.ApiKey
	if scrapInApiKey == "" {
		err := errors.New("Scrapin Api key not set")
		tracing.TraceErr(span, err)
		s.log.Errorf("Scrapin Api key not set")
		return &postgresentity.ScrapInResponseBody{}, err
	}

	params := url.Values{}
	params.Add("apikey", scrapInApiKey)
	params.Add("linkedInUrl", linkedInUrl)

	scrapinStatusCode, body, err := makeScrapInHTTPRequest(baseUrl + "/enrichment/company" + "?" + params.Encode())
	if err != nil {
		tracing.TraceErr(span, errors.Wrap(err, "makeScrapInHTTPRequest"))
		s.log.Errorf("Error making scrapin HTTP request: %s", err.Error())
		return &postgresentity.ScrapInResponseBody{}, err
	}
	span.LogFields(log.Int("scrapin.statusCode", scrapinStatusCode))

	var scrapinResponse postgresentity.ScrapInResponseBody
	err = json.Unmarshal(body, &scrapinResponse)
	if err != nil {
		tracing.TraceErr(span, errors.Wrap(err, "json.Unmarshal"))
		span.LogFields(log.String("response.body", string(body)))
		s.log.Errorf("Error unmarshalling scrapin response: %s", err.Error())
		return &postgresentity.ScrapInResponseBody{}, err
	}

	return &scrapinResponse, nil
}

func (s *scrapinService) ScrapInSearchCompany(ctx context.Context, domain string) (uint64, *postgresentity.ScrapInResponseBody, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ScrapinService.ScrapInSearchCompany")
	defer span.Finish()
	span.LogFields(log.String("domain", domain))

	latestEnrichDetailsScrapInRecord, err := s.services.CommonServices.PostgresRepositories.EnrichDetailsScrapInRepository.GetLatestByParam1AndFlow(ctx, domain, postgresentity.ScrapInFlowCompanySearch)
	if err != nil {
		tracing.TraceErr(span, errors.Wrap(err, "failed to get scrapin data"))
		return 0, nil, err
	}

	var data *postgresentity.ScrapInResponseBody
	var recordId uint64

	callScrapInNow := false

	if latestEnrichDetailsScrapInRecord == nil || latestEnrichDetailsScrapInRecord.UpdatedAt.AddDate(0, 0, s.config.ScrapinConfig.TtlDays).Before(utils.Now()) {
		callScrapInNow = true
	} else if latestEnrichDetailsScrapInRecord.CompanyFound == false {
		// if last attempt was > 1 day ago, call scrapin
		if latestEnrichDetailsScrapInRecord.UpdatedAt.AddDate(0, 0, 1).Before(utils.Now()) {
			callScrapInNow = true
		}
	}

	// if cached data is missing or last time fetched > ttl refresh
	if callScrapInNow {
		// get data from scrapin
		if data, err = s.callScrapinCompanySearch(ctx, domain); err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to call scrapin"))
			return 0, nil, err
		}

		traceIfCreditsDepleting(data, span)

		// save to db
		dataAsString, err := json.Marshal(data)
		if err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to marshal data"))
			return 0, nil, err
		}
		paramsAsString, err := json.Marshal(ScrapInSearchRequestParams{
			CompanyDomain: domain,
		})
		if err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to marshal params"))
		}
		dbRecord, err := s.services.CommonServices.PostgresRepositories.EnrichDetailsScrapInRepository.Create(ctx, postgresentity.EnrichDetailsScrapIn{
			Param1:        domain,
			Flow:          postgresentity.ScrapInFlowCompanySearch,
			AllParamsJson: string(paramsAsString),
			Data:          string(dataAsString),
			PersonFound:   false,
			CompanyFound:  data.Company != nil,
			Success:       data.Success,
		})
		if err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to save scrapin data in db"))
		}
		if dbRecord != nil {
			recordId = dbRecord.ID
		}
	} else {
		recordId = latestEnrichDetailsScrapInRecord.ID
		// unmarshal cached data
		unmarshalledData := postgresentity.ScrapInResponseBody{}
		if err = json.Unmarshal([]byte(latestEnrichDetailsScrapInRecord.Data), &unmarshalledData); err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to unmarshal scrapin cached data"))
			return 0, nil, err
		}
		data = &unmarshalledData
	}

	// if fresh data not found, check most recent cached data with company found
	if data == nil || data.Company == nil {
		latestEnrichDetailsScrapInRecordWithCompanyFound, err := s.services.CommonServices.PostgresRepositories.EnrichDetailsScrapInRepository.GetLatestByParam1AndFlowWithCompanyFound(ctx, domain, postgresentity.ScrapInFlowCompanySearch)
		if err != nil {
			tracing.TraceErr(span, errors.Wrap(err, "failed to get scrapin data"))
			return 0, nil, err
		}
		if latestEnrichDetailsScrapInRecordWithCompanyFound != nil {
			unmarshalledData := postgresentity.ScrapInResponseBody{}
			if err = json.Unmarshal([]byte(latestEnrichDetailsScrapInRecordWithCompanyFound.Data), &unmarshalledData); err != nil {
				tracing.TraceErr(span, errors.Wrap(err, "failed to unmarshal scrapin cached data"))
				return 0, nil, err
			}
			data = &unmarshalledData
			return latestEnrichDetailsScrapInRecordWithCompanyFound.ID, data, nil
		}
	}

	return recordId, data, nil
}

func (s *scrapinService) callScrapinCompanySearch(ctx context.Context, domain string) (*postgresentity.ScrapInResponseBody, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ScrapInService.callScrapinCompanySearch")
	defer span.Finish()
	span.LogKV("domain", domain)

	baseUrl := s.config.ScrapinConfig.Url
	if baseUrl == "" {
		err := errors.New("ScrapIn URL not set")
		tracing.TraceErr(span, err)
		s.log.Errorf("ScrapIn URL not set")
		return &postgresentity.ScrapInResponseBody{}, err
	}

	scrapInApiKey := s.config.ScrapinConfig.ApiKey
	if scrapInApiKey == "" {
		err := errors.New("Scrapin Api key not set")
		tracing.TraceErr(span, err)
		s.log.Errorf("Scrapin Api key not set")
		return &postgresentity.ScrapInResponseBody{}, err
	}

	params := url.Values{}
	params.Add("apikey", scrapInApiKey)
	params.Add("domain", domain)
	scrapinUrl := baseUrl + "/enrichment/company/domain" + "?" + params.Encode()

	scrapinStatusCode, body, err := makeScrapInHTTPRequest(scrapinUrl)
	if err != nil {
		tracing.TraceErr(span, errors.Wrap(err, "makeScrapInHTTPRequest"))
		s.log.Errorf("Error making scrapin HTTP request: %s", err.Error())
		return &postgresentity.ScrapInResponseBody{}, err
	}
	span.LogFields(log.Int("scrapin.statusCode", scrapinStatusCode))

	var scrapinResponse postgresentity.ScrapInResponseBody
	err = json.Unmarshal(body, &scrapinResponse)
	if err != nil {
		tracing.TraceErr(span, errors.Wrap(err, "json.Unmarshal"))
		span.LogFields(log.String("response.body", string(body)))
		s.log.Errorf("Error unmarshalling scrapin response: %s", err.Error())
		return &postgresentity.ScrapInResponseBody{}, err
	}

	return &scrapinResponse, nil
}

func makeScrapInHTTPRequest(url string) (int, []byte, error) {
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")

	var body []byte
	var err error
	var res *http.Response
	var statusCode int

	// try few times to get data from scrapin
	for i := 0; i < 2; i++ {
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			return -1, nil, err
		}
		defer res.Body.Close()

		body, err = io.ReadAll(res.Body)
		statusCode = res.StatusCode

		if statusCode == http.StatusTooManyRequests {
			// sleep 1 second and try again
			time.Sleep(1 * time.Second)
			continue
		} else {
			return statusCode, body, err
		}
	}
	return statusCode, body, err
}

func traceIfCreditsDepleting(data *postgresentity.ScrapInResponseBody, span opentracing.Span) {
	if data != nil {
		if data.CreditsLeft > 0 && data.CreditsLeft < 50 {
			tracing.TraceErr(span, errors.New(fmt.Sprintf("ScrapIn credits are depleting, only %d credits left", data.CreditsLeft)))
		}
		if data.RateLimitLeft > 0 && data.RateLimitLeft < 50 {
			tracing.TraceErr(span, errors.New(fmt.Sprintf("ScrapIn rate limit is depleting, only %d requests left", data.RateLimitLeft)))
		}
	}
}
