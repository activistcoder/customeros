package main

import (
	"context"
	"github.com/openline-ai/openline-customer-os/packages/runner/sync-tracking/config"
	"github.com/openline-ai/openline-customer-os/packages/runner/sync-tracking/constants"
	localcron "github.com/openline-ai/openline-customer-os/packages/runner/sync-tracking/cron"
	"github.com/openline-ai/openline-customer-os/packages/runner/sync-tracking/logger"
	"github.com/openline-ai/openline-customer-os/packages/runner/sync-tracking/service"
	commconf "github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/config"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/grpc_client"
	"github.com/openline-ai/openline-customer-os/packages/server/customer-os-common-module/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/robfig/cron"
	"io"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Config
	cfg := config.Load()

	// Logging
	appLogger := initLogger(cfg)

	// Tracing
	tracingCloser := initTracing(cfg, appLogger)
	if tracingCloser != nil {
		defer tracingCloser.Close()
	}

	ctx := context.Background()

	// Initialize postgres db
	postgresDb, _ := InitDB(cfg, appLogger)
	defer postgresDb.SqlDB.Close()

	// Neo4j DB
	neo4jDriver, errNeo4j := commconf.NewNeo4jDriver(cfg.Neo4j)
	if errNeo4j != nil {
		appLogger.Fatalf("failed opening connection to neo4j: %v", errNeo4j.Error())
	}
	defer (neo4jDriver).Close(ctx)

	// Events processing
	df := grpc_client.NewDialFactory(&cfg.GrpcClientConfig)
	gRPCconn, err := df.GetEventsProcessingPlatformConn()
	defer df.Close(gRPCconn)
	if err != nil {
		appLogger.Fatalf("Failed to connect: %v", err)
	}
	grpcClient := grpc_client.InitClients(gRPCconn)

	services := service.InitServices(cfg, &neo4jDriver, postgresDb.GormDB, grpcClient)

	cronJub := localcron.StartCron(cfg, services)

	if err := run(appLogger, cronJub); err != nil {
		appLogger.Fatal(err)
	}

	// Flush logs and exit
	appLogger.Sync()
}

func run(log logger.Logger, cron *cron.Cron) error {
	defer cron.Stop()

	// Shutdown handling
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	sig := <-shutdown
	log.Infof("Received shutdown signal %v", sig)

	// Gracefully stop
	if err := localcron.StopCron(log, cron); err != nil {
		return err
	}
	log.Info("Graceful shutdown complete")

	return nil
}

func initLogger(cfg *config.Config) logger.Logger {
	appLogger := logger.NewExtendedAppLogger(&cfg.Logger)
	appLogger.InitLogger()
	appLogger.WithName(constants.ServiceName)
	return appLogger
}

func initTracing(cfg *config.Config, appLogger logger.Logger) io.Closer {
	tracer, closer, err := tracing.NewJaegerTracer(&cfg.Jaeger, appLogger)
	if err != nil {
		appLogger.Fatalf("Could not initialize jaeger tracer: %v", err.Error())
	}
	opentracing.SetGlobalTracer(tracer)
	return closer
}

func InitDB(cfg *config.Config, log logger.Logger) (db *commconf.StorageDB, err error) {
	if db, err = commconf.NewPostgresDBConn(cfg.Postgres); err != nil {
		log.Fatalf("Could not open db connection: %s", err.Error())
	}
	return
}