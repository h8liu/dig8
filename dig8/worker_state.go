package dig8

const (
	workerError = iota
	workerIdle
	workerPending
	workerBusy
)

// WorkerState is the state of a worker
type WorkerState struct {
	Worker string
	State  int
}
