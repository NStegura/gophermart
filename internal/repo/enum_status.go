package repo

type OrderStatus int

const (
	NEW OrderStatus = iota + 1
	PROCESSING
	INVALID
	PROCESSED
)

func (os OrderStatus) String() string {
	return [...]string{"NEW", "PROCESSING", "INVALID", "PROCESSED"}[os-1]
}

func (os OrderStatus) Index() int {
	return int(os)
}
