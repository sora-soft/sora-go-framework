package types

import (
	"fmt"

	"github.com/sora-soft/sora-go-framework.git/pkg/utility"
)

type WorkerState int

const (
	WorkerStateInit     WorkerState = 1
	WorkerStatePending  WorkerState = 2
	WorkerStateReady    WorkerState = 3
	WorkerStateStopping WorkerState = 4
	WorkerStateStopped  WorkerState = 5
	WorkerStateError    WorkerState = 100
)

func (s WorkerState) String() string {
	switch s {
	case WorkerStateInit:
		return "Init"
	case WorkerStatePending:
		return "Pending"
	case WorkerStateReady:
		return "Ready"
	case WorkerStateStopping:
		return "Stopping"
	case WorkerStateStopped:
		return "Stopped"
	case WorkerStateError:
		return "Error"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

type WorkerMetaData struct {
	Name      string         `json:"name,omitempty"`
	Alias     *string        `json:"alias,omitempty"`
	State     WorkerState    `json:"state,omitempty"`
	Id        string         `json:"id,omitempty"`
	StartTime int64          `json:"startTime,omitempty"`
	Labels    utility.Labels `json:"labels,omitempty"`
}
