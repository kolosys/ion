package circuit

import (
	"errors"
	"fmt"
)

// CircuitError represents circuit breaker specific errors with context
type CircuitError struct {
	Op          string // operation that failed
	CircuitName string // name of the circuit breaker
	State       string // current state of the circuit
	Err         error  // underlying error
}

func (e *CircuitError) Error() string {
	if e.CircuitName != "" {
		return fmt.Sprintf("ion: circuit %q %s (state: %s): %v", e.CircuitName, e.Op, e.State, e.Err)
	}
	return fmt.Sprintf("ion: circuit %s (state: %s): %v", e.Op, e.State, e.Err)
}

func (e *CircuitError) Unwrap() error {
	return e.Err
}

// IsCircuitOpen returns true if the error is due to an open circuit.
func (e *CircuitError) IsCircuitOpen() bool {
	return e.State == "Open"
}

// NewCircuitOpenError creates an error indicating the circuit is open
func NewCircuitOpenError(circuitName string) error {
	return &CircuitError{
		Op:          "execute",
		CircuitName: circuitName,
		State:       "Open",
		Err:         errors.New("circuit breaker is open"),
	}
}

// NewCircuitTimeoutError creates an error indicating a circuit operation timed out
func NewCircuitTimeoutError(circuitName string) error {
	return &CircuitError{
		Op:          "execute",
		CircuitName: circuitName,
		State:       "Unknown",
		Err:         errors.New("circuit breaker operation timeout"),
	}
}
