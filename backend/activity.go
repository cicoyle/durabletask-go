package backend

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/dapr/durabletask-go/api"
	"github.com/dapr/durabletask-go/api/helpers"
	"github.com/dapr/durabletask-go/api/protos"
)

type activityProcessor struct {
	be       Backend
	executor ActivityExecutor
}

type ActivityExecutor interface {
	ExecuteActivity(context.Context, api.InstanceID, *protos.HistoryEvent) (*protos.HistoryEvent, error)
}

func NewActivityTaskWorker(be Backend, executor ActivityExecutor, logger Logger, opts ...NewTaskWorkerOptions) TaskWorker[*ActivityWorkItem] {
	processor := newActivityProcessor(be, executor)
	return NewTaskWorker(processor, logger, opts...)
}

func newActivityProcessor(be Backend, executor ActivityExecutor) TaskProcessor[*ActivityWorkItem] {
	return &activityProcessor{
		be:       be,
		executor: executor,
	}
}

// Name implements TaskProcessor
func (*activityProcessor) Name() string {
	return "activity-processor"
}

// NextWorkItem implements TaskDispatcher
func (ap *activityProcessor) NextWorkItem(ctx context.Context) (*ActivityWorkItem, error) {
	return ap.be.NextActivityWorkItem(ctx)
}

// ProcessWorkItem implements TaskDispatcher
func (p *activityProcessor) ProcessWorkItem(ctx context.Context, awi *ActivityWorkItem) error {
	ts := awi.NewEvent.GetTaskScheduled()
	fmt.Printf("%v: processing work item: %s\n", p.Name(), awi)
	fmt.Printf("%v: processing ts: %+v\n", ts)

	if ts == nil {
		return fmt.Errorf("%v: invalid TaskScheduled event", awi.InstanceID)
	}

	// Store orchestrator's AppId from TaskScheduled event for later completion event routing
	awi.OrchestratorAppId = ts.OrchestrationAppId

	// Create span as child of spanContext found in TaskScheduledEvent
	ctx, err := helpers.ContextFromTraceContext(ctx, ts.ParentTraceContext)
	if err != nil {
		return fmt.Errorf("%v: failed to parse activity trace context: %w", awi.InstanceID, err)
	}
	var span trace.Span
	ctx, span = helpers.StartNewActivitySpan(ctx, ts.Name, ts.Version.GetValue(), string(awi.InstanceID), awi.NewEvent.EventId)
	if span != nil {
		defer func() {
			if r := recover(); r != nil {
				span.SetStatus(codes.Error, fmt.Sprintf("%v", r))
			}
			span.End()
		}()
	}

	fmt.Printf("%v: processing work item: %+v\n", *awi.OrchestratorAppId, ts)
	fmt.Println("cassie, I think here it should just create a pending activity if the task is for calling another appID activity but we see we see")
	// Execute the activity and get its result
	result, err := p.executor.ExecuteActivity(ctx, awi.InstanceID, awi.NewEvent)
	if err != nil {
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	if result != nil && awi.OrchestratorAppId != nil {
		result.OrchestrationAppID = awi.OrchestratorAppId
	}

	awi.Result = result
	return nil
}

// CompleteWorkItem implements TaskDispatcher
func (ap *activityProcessor) CompleteWorkItem(ctx context.Context, awi *ActivityWorkItem) error {
	fmt.Printf("[DURABLETASK] %v: completing work item: %s\n", ap.Name(), awi)
	if awi.Result == nil {
		return fmt.Errorf("can't complete work item '%s' with nil result", awi)
	}
	if awi.Result.GetTaskCompleted() == nil && awi.Result.GetTaskFailed() == nil {
		return fmt.Errorf("can't complete work item '%s', which isn't TaskCompleted or TaskFailed", awi)
	}

	return ap.be.CompleteActivityWorkItem(ctx, awi)
}

// AbandonWorkItem implements TaskDispatcher
func (ap *activityProcessor) AbandonWorkItem(ctx context.Context, awi *ActivityWorkItem) error {
	return ap.be.AbandonActivityWorkItem(ctx, awi)
}
