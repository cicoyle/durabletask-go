package backend

import (
	"fmt"
	"time"

	"github.com/dapr/durabletask-go/api"
	"github.com/dapr/durabletask-go/api/protos"
)

type WorkItem interface {
	fmt.Stringer
	IsWorkItem() bool
}

type OrchestrationWorkItem struct {
	InstanceID api.InstanceID
	NewEvents  []*HistoryEvent
	LockedBy   string
	RetryCount int32
	State      *protos.OrchestrationRuntimeState
	Properties map[string]interface{}
}

// String implements core.WorkItem and fmt.Stringer
func (wi OrchestrationWorkItem) String() string {
	return fmt.Sprintf("%s (%d event(s))", wi.InstanceID, len(wi.NewEvents))
}

// IsWorkItem implements core.WorkItem
func (wi OrchestrationWorkItem) IsWorkItem() bool {
	return true
}

func (wi *OrchestrationWorkItem) GetAbandonDelay() time.Duration {
	switch {
	case wi.RetryCount == 0:
		return time.Duration(0) // no delay
	case wi.RetryCount > 100:
		return 5 * time.Minute // max delay
	default:
		return time.Duration(wi.RetryCount) * time.Second // linear backoff
	}
}

type ActivityWorkItem struct {
	SequenceNumber    int64
	InstanceID        api.InstanceID
	NewEvent          *HistoryEvent
	Result            *HistoryEvent
	LockedBy          string
	Properties        map[string]interface{}
	OrchestratorAppId *string // The appId of the orchestrator that scheduled this activity
}

// String implements core.WorkItem and fmt.Stringer
func (wi ActivityWorkItem) String() string {
	name := wi.NewEvent.GetTaskScheduled().GetName()
	taskID := wi.NewEvent.EventId
	return fmt.Sprintf("%s/%s#%d", wi.InstanceID, name, taskID)
}

// IsWorkItem implements core.WorkItem
func (wi ActivityWorkItem) IsWorkItem() bool {
	return true
}
