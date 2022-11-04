package temporal_server

import (
	"context"
	"github.com/ovargas/wizapp/sdk/app"
	"github.com/ovargas/wizapp/sdk/logger"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/interceptor"
	temporal_log "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"time"
)

const (
	ServiceName = "temporal-worker"
)

var (
	log = logger.Log()

	registry           func(w Worker)
	onFatalError       func(error)
	workerInterceptors []WorkerInterceptor
	temporalLogger     Logger
	metricHandler      MetricHandler
	identity           string
	dataConverter      DataConverter
	contextPropagators []ContextPropagator
)

func init() {
	app.RegisterServer(ServiceName, createWorker)
}

type (
	MetricHandler     = client.MetricsHandler
	WorkerInterceptor = interceptor.WorkerInterceptor
	DataConverter     = converter.DataConverter
	ContextPropagator = workflow.ContextPropagator
	Logger            = temporal_log.Logger

	server struct {
		app.UnimplementedServer
		worker    worker.Worker
		isStarted bool
	}

	WorkerConfig struct {
		MaxConcurrentActivityExecutionSize      int                        `mapstructure:"max_concurrent_activity_execution_size"`
		WorkerActivitiesPerSecond               float64                    `mapstructure:"worker_activities_per_second"`
		MaxConcurrentLocalActivityExecutionSize int                        `mapstructure:"max_concurrent_local_activity_execution_size"`
		WorkerLocalActivitiesPerSecond          float64                    `mapstructure:"worker_local_activities_per_second"`
		TaskQueueActivitiesPerSecond            float64                    `mapstructure:"task_queue_activities_per_second"`
		MaxConcurrentActivityTaskPollers        int                        `mapstructure:"max_concurrent_activity_task_pollers"`
		MaxConcurrentWorkflowTaskExecutionSize  int                        `mapstructure:"max_concurrent_workflow_task_execution_size"`
		MaxConcurrentWorkflowTaskPollers        int                        `mapstructure:"max_concurrent_workflow_task_pollers"`
		EnableLoggingInReplay                   bool                       `mapstructure:"enable_logging_in_replay"`
		StickyScheduleToStartTimeout            time.Duration              `mapstructure:"sticky_schedule_to_start_timeout"`
		BackgroundActivityContext               context.Context            `mapstructure:"background_activity_context"`
		WorkflowPanicPolicy                     worker.WorkflowPanicPolicy `mapstructure:"workflow_panic_policy"`
		WorkerStopTimeout                       time.Duration              `mapstructure:"worker_stop_timeout"`
		EnableSessionWorker                     bool                       `mapstructure:"enable_session_worker"`
		MaxConcurrentSessionExecutionSize       int                        `mapstructure:"max_concurrent_session_execution_size"`
		DisableWorkflowWorker                   bool                       `mapstructure:"disable_workflow_worker"`
		LocalActivityWorkerOnly                 bool                       `mapstructure:"local_activity_worker_only"`
		Identity                                string                     `mapstructure:"identity"`
		DeadlockDetectionTimeout                time.Duration              `mapstructure:"deadlock_detection_timeout"`
		MaxHeartbeatThrottleInterval            time.Duration              `mapstructure:"max_heartbeat_throttle_interval"`
		DefaultHeartbeatThrottleInterval        time.Duration              `mapstructure:"default_heartbeat_throttle_interval"`
		DisableEagerActivities                  bool                       `mapstructure:"disable_eager_activities"`
		MaxConcurrentEagerActivityExecutionSize int                        `mapstructure:"max_concurrent_eager_activity_execution_size"`
		DisableRegistrationAliasing             bool                       `mapstructure:"disable_registration_aliasing"`
	}

	Config struct {
		HostPort  string       `mapstructure:"host_port"`
		Namespace string       `mapstructure:"namespace"`
		TaskQueue string       `mapstructure:"task_queue"`
		Worker    WorkerConfig `mapstructure:"worker"`
	}

	Worker interface {
		worker.WorkflowRegistry
		worker.ActivityRegistry
	}
)

func SetLogger(logger Logger) {
	temporalLogger = logger
}

func SetMetricHandler(handler MetricHandler) {
	metricHandler = handler
}

func SetIdentity(name string) {
	identity = name
}

func SetDataConverter(converter DataConverter) {
	dataConverter = converter
}

func SetContextPropagators(propagators ...ContextPropagator) {
	contextPropagators = append(contextPropagators, propagators...)
}

func WorkerRegistry(fn func(w Worker)) {
	registry = fn
}

func OnFatalError(fn func(error)) {
	onFatalError = fn
}

func WorkerInterceptors(interceptors ...WorkerInterceptor) {
	workerInterceptors = append(interceptors, interceptors...)
}

func createWorker(config *app.ApplicationConfig) (app.Server, error) {
	var cfg Config

	if err := config.UnmarshalKey("temporal", &cfg); err != nil {
		return nil, err
	}

	dial, err := client.Dial(client.Options{
		HostPort:           cfg.HostPort,
		Namespace:          cfg.Namespace,
		Logger:             temporalLogger,
		MetricsHandler:     metricHandler,
		Identity:           identity,
		DataConverter:      dataConverter,
		ContextPropagators: contextPropagators,
	})

	if err != nil {
		return nil, err
	}

	w := worker.New(dial, cfg.TaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize:      cfg.Worker.MaxConcurrentActivityExecutionSize,
		WorkerActivitiesPerSecond:               cfg.Worker.WorkerActivitiesPerSecond,
		MaxConcurrentLocalActivityExecutionSize: cfg.Worker.MaxConcurrentLocalActivityExecutionSize,
		WorkerLocalActivitiesPerSecond:          cfg.Worker.WorkerLocalActivitiesPerSecond,
		TaskQueueActivitiesPerSecond:            cfg.Worker.TaskQueueActivitiesPerSecond,
		MaxConcurrentActivityTaskPollers:        cfg.Worker.MaxConcurrentActivityTaskPollers,
		MaxConcurrentWorkflowTaskExecutionSize:  cfg.Worker.MaxConcurrentWorkflowTaskExecutionSize,
		MaxConcurrentWorkflowTaskPollers:        cfg.Worker.MaxConcurrentWorkflowTaskPollers,
		EnableLoggingInReplay:                   cfg.Worker.EnableLoggingInReplay,
		DisableStickyExecution:                  false,
		StickyScheduleToStartTimeout:            cfg.Worker.StickyScheduleToStartTimeout,
		BackgroundActivityContext:               cfg.Worker.BackgroundActivityContext,
		WorkflowPanicPolicy:                     cfg.Worker.WorkflowPanicPolicy,
		WorkerStopTimeout:                       cfg.Worker.WorkerStopTimeout,
		EnableSessionWorker:                     cfg.Worker.EnableSessionWorker,
		MaxConcurrentSessionExecutionSize:       cfg.Worker.MaxConcurrentSessionExecutionSize,
		DisableWorkflowWorker:                   cfg.Worker.DisableWorkflowWorker,
		LocalActivityWorkerOnly:                 cfg.Worker.LocalActivityWorkerOnly,
		Identity:                                cfg.Worker.Identity,
		DeadlockDetectionTimeout:                cfg.Worker.DeadlockDetectionTimeout,
		MaxHeartbeatThrottleInterval:            cfg.Worker.MaxHeartbeatThrottleInterval,
		DefaultHeartbeatThrottleInterval:        cfg.Worker.DefaultHeartbeatThrottleInterval,
		Interceptors:                            workerInterceptors,
		OnFatalError:                            onFatalError,
		DisableEagerActivities:                  cfg.Worker.DisableEagerActivities,
		MaxConcurrentEagerActivityExecutionSize: cfg.Worker.MaxConcurrentEagerActivityExecutionSize,
		DisableRegistrationAliasing:             cfg.Worker.DisableRegistrationAliasing,
	})

	if registry != nil {
		registry(w)
	}

	return &server{
		worker: w,
	}, nil
}

func (w *server) Start() error {
	if w.isStarted {
		return nil
	}

	w.isStarted = true

	defer func() {
		w.isStarted = false
	}()

	if err := w.worker.Run(worker.InterruptCh()); err != nil {
		log.Fatalf("Unable to start temporal worker: %v", err)
	}

	return nil
}

func (w *server) Stop() error {
	if w.isStarted {
		w.worker.Stop()
		log.Infof("temporal worker stopped")
	}
	return nil
}
