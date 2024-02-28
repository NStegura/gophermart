package models

type AccrualStatus int

const (
	REGISTERED AccrualStatus = iota + 1
	PROCESSING
	INVALID
	PROCESSED
)

func statuses() [4]string {
	return [4]string{"REGISTERED", "PROCESSING", "INVALID", "PROCESSED"}
}

func (as AccrualStatus) String() string {
	return statuses()[as-1]
}

func (as AccrualStatus) Index() int {
	return int(as)
}

type OrderAccrual struct {
	OrderID int64   `json:"order,string"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

func (oa OrderAccrual) IsValid() bool {
	for _, s := range statuses() {
		if oa.Status == s {
			return true
		}
	}
	return false
}
