package turnorchestrator

import "errors"

var (
	errDependencyPanic = errors.New("turn orchestrator dependency panicked")
	errStreamClosed    = errors.New("turn stream is closed")
)
