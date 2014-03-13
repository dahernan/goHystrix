package goHystrix

//type CircuitBreaker struct {
//}

type Circuit interface {
	IsOpen()
}
