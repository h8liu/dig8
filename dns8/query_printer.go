package dns8

// QueryPrinter printing options
const (
	PrintAll = iota
	PrintReply
)

// QueryPrinter prints a query
type QueryPrinter struct {
	*Query

	Printer   *Printer
	PrintFlag int
}
