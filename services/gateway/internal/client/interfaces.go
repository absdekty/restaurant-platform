package client

type CircuitBreaker interface {
	Execute(fn func() (interface{}, error)) (interface{}, error)
}
