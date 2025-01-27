package main

import (
	"context"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	commonConfig "github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/config"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/dto"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/grpc_client"
	commonService "github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/service"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/tracing"
	"github.com/openline-ai/openline-customer-os/packages/server/events-subscribers/config"
	"github.com/openline-ai/openline-customer-os/packages/server/events-subscribers/listeners"
	"github.com/openline-ai/openline-customer-os/packages/server/events-subscribers/logger"
	"github.com/opentracing/opentracing-go"
	"io"
	"log"
)

const (
	AppName = "events-subscribers"
)

func InitDB(cfg *config.Config, appLogger logger.Logger) (db *config.StorageDB, err error) {
	if db, err = config.NewDBConn(cfg); err != nil {
		appLogger.Fatalf("Coud not open db connection: %s", err.Error())
	}
	return
}

func main() {
	cfg := loadConfiguration()

	appLogger := logger.NewExtendedAppLogger(&cfg.Logger)
	appLogger.InitLogger()
	appLogger.WithName(AppName)

	// Initialize Tracing
	tracingCloser := initTracing(cfg, appLogger)
	if tracingCloser != nil {
		defer tracingCloser.Close()
	}

	db, _ := InitDB(cfg, appLogger)
	defer db.SqlDB.Close()

	neo4jDriver, err := config.NewDriver(cfg)
	if err != nil {
		appLogger.Fatalf("Could not establish connection with neo4j at: %v, error: %v", cfg.Neo4j.Target, err.Error())
	}
	ctx := context.Background()
	defer neo4jDriver.Close(ctx)

	// Events processing
	var eventsProcessingGrpcClient *grpc_client.Clients
	if cfg.GrpcClientConfig.EventsProcessingPlatformEnabled {
		df := grpc_client.NewDialFactory(&cfg.GrpcClientConfig)
		gRPCconn, err := df.GetEventsProcessingPlatformConn()
		defer df.Close(gRPCconn)
		if err != nil {
			appLogger.Fatalf("Failed to connect: %v", err)
		}
		eventsProcessingGrpcClient = grpc_client.InitClients(gRPCconn)
	}

	commonServices := commonService.InitServices(&commonConfig.GlobalConfig{
		RabbitMQConfig: &cfg.RabbitMQ,
		NovuConfig:     &cfg.NovuConfig,
		InternalServices: commonConfig.InternalServices{
			EnrichmentApiConfig: cfg.InternalServices.EnrichmentApi,
			AiApiConfig:         cfg.InternalServices.AiApi,
		},
	}, db.GormDB, &neo4jDriver, cfg.Neo4j.Database, eventsProcessingGrpcClient, appLogger)

	//Register listeners
	commonServices.RabbitMQService.RegisterHandler(dto.FlowSchedule{}, listeners.Handle_FlowSchedule)
	commonServices.RabbitMQService.RegisterHandler(dto.FlowParticipantSchedule{}, listeners.Handle_FlowParticipantSchedule)
	commonServices.RabbitMQService.RegisterHandler(dto.FlowComputeParticipantsRequirements{}, listeners.Handle_FlowComputeParticipantsRequirements)
	commonServices.RabbitMQService.RegisterHandler(dto.FlowParticipantGoalAchieved{}, listeners.Handle_FlowParticipantGoalAchieved)
	commonServices.RabbitMQService.RegisterHandler(dto.AddSocialToContact{}, listeners.OnSocialAddedToContact)
	commonServices.RabbitMQService.RegisterHandler(dto.RequestEnrichContact{}, listeners.OnRequestedEnrichContact)
	commonServices.RabbitMQService.RegisterHandler(dto.RequestRefreshLastTouchpoint{}, listeners.OnRequestLastTouchpointRefresh)

	// Listen for messages
	commonServices.RabbitMQService.ListenQueue(commonService.EventsQueueName, false)
	commonServices.RabbitMQService.ListenQueue(commonService.EventsFlowParticipantScheduleQueueName, true)

	// Block the main thread from exiting
	forever := make(chan bool)
	log.Println("Waiting for messages. To exit press CTRL+C")
	<-forever

}

func loadConfiguration() *config.Config {
	if err := godotenv.Load(); err != nil {
		log.Println("[WARNING] Error loading .env file")
	}

	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Printf("%+v\n", err)
	}

	return &cfg
}

func initTracing(cfg *config.Config, appLogger logger.Logger) io.Closer {
	if cfg.Jaeger.Enabled {
		tracer, closer, err := tracing.NewJaegerTracer(&cfg.Jaeger, appLogger)
		if err != nil {
			appLogger.Fatalf("Could not initialize jaeger tracer: %v", err.Error())
		}
		opentracing.SetGlobalTracer(tracer)
		return closer
	}
	return nil
}
