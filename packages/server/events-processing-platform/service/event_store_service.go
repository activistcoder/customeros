package service

import (
	"context"
	"github.com/EventStore/EventStore-Client-Go/v3/esdb"
	"github.com/openline-ai/openline-customer-os/packages/server/events-processing-platform/constants"
	"github.com/openline-ai/openline-customer-os/packages/server/events-processing-platform/domain/common/aggregate"
	"github.com/openline-ai/openline-customer-os/packages/server/events-processing-platform/eventstore"
	grpcerr "github.com/openline-ai/openline-customer-os/packages/server/events-processing-platform/grpc_errors"
	"github.com/openline-ai/openline-customer-os/packages/server/events-processing-platform/logger"
	"github.com/openline-ai/openline-customer-os/packages/server/events-processing-platform/tracing"
	commentpb "github.com/openline-ai/openline-customer-os/packages/server/events-processing-proto/gen/proto/go/api/grpc/v1/comment"
	eventstorepb "github.com/openline-ai/openline-customer-os/packages/server/events-processing-proto/gen/proto/go/api/grpc/v1/event_store"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

type eventStoreService struct {
	commentpb.UnimplementedCommentGrpcServiceServer
	services       *Services
	log            logger.Logger
	aggregateStore eventstore.AggregateStore
}

func NewEventStoreService(services *Services, log logger.Logger, aggregateStore eventstore.AggregateStore) *eventStoreService {
	return &eventStoreService{
		services:       services,
		log:            log,
		aggregateStore: aggregateStore,
	}
}

func (s *eventStoreService) DeleteEventStoreStream(ctx context.Context, request *eventstorepb.DeleteEventStoreStreamRequest) (*emptypb.Empty, error) {
	ctx, span := tracing.StartGrpcServerTracerSpan(ctx, "EventStoreService.DeleteEventStoreStream")
	defer span.Finish()
	tracing.SetServiceSpanTags(ctx, span, request.Tenant, "")
	tracing.LogObjectAsJson(span, "request", request)

	if request.Tenant == "" {
		err := grpcerr.ErrMissingField("tenant")
		tracing.TraceErr(span, err)
		return nil, grpcerr.ErrResponse(err)
	} else if request.Type == "" {
		err := grpcerr.ErrMissingField("type")
		tracing.TraceErr(span, err)
		return nil, grpcerr.ErrResponse(err)
	} else if request.Id == "" {
		err := grpcerr.ErrMissingField("id")
		tracing.TraceErr(span, err)
		return nil, grpcerr.ErrResponse(err)
	}

	aggr := aggregate.NewCommonAggregateWithTenantAndId(eventstore.AggregateType(request.Type), request.Tenant, request.Id)
	// Check if aggregate exists
	err := s.aggregateStore.Exists(ctx, aggr.GetID())
	if err != nil {
		if !errors.Is(err, eventstore.ErrAggregateNotFound) {
			return &emptypb.Empty{}, grpcerr.ErrResponse(err)
		} else {
			return &emptypb.Empty{}, nil
		}
	}

	// 1 day in seconds
	maxAgeSeconds := constants.StreamMetadataMaxAgeSeconds
	if request.MinutesUntilDeletion > 0 {
		maxAgeSeconds = int(request.MinutesUntilDeletion * 60)
	}

	streamMetadata := esdb.StreamMetadata{}
	streamMetadata.SetMaxAge(time.Duration(maxAgeSeconds) * time.Second)

	err = s.aggregateStore.UpdateStreamMetadata(ctx, aggr.GetID(), streamMetadata)
	if err != nil {
		tracing.TraceErr(span, err)
		s.log.Errorf("Error while updating stream metadata: %s", err)
		return &emptypb.Empty{}, grpcerr.ErrResponse(err)
	}

	return &emptypb.Empty{}, nil
}

func (s *eventStoreService) errResponse(err error) error {
	return grpcerr.ErrResponse(err)
}