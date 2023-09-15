package gophermart

import (
	"fmt"
	"github.com/Osselnet/gophermart.git/pkg/luhn"
	"strconv"
	"time"
)

const (
	StatusNew        = "NEW"
	StatusProcessing = "PROCESSING"
	StatusInvalid    = "INVALID"
	StatusProcessed  = "PROCESSED"
)

func IsValidStatus(status string) bool {
	switch status {
	case StatusNew:
	case StatusProcessing:
	case StatusProcessed:
	case StatusInvalid:
	default:
		return false
	}

	return true
}

type Order struct {
	ID         uint64
	UserID     uint64
	Status     string
	Accrual    uint64
	UploadedAt time.Time
}

type OrderProxy struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

func (op *OrderProxy) String() string {
	return fmt.Sprintf("%#v\n", op)
}

type orders struct {
	linker *GopherMart
}

func newOrders(linker *GopherMart) *orders {
	return &orders{
		linker: linker,
	}
}

func (os *orders) Add(orderID, userID uint64) error {
	strOrderID := strconv.Itoa(int(orderID))
	if !luhn.IsValid(strOrderID) {
		return ErrOrderInvalidFormat
	}

	order, _ := os.Get(orderID)
	if order != nil {
		if order.UserID == userID {
			return ErrOrderAlreadyLoadedByUser
		}
		return ErrOrderAlreadyLoadedByAnotherUser
	}

	order = &Order{
		ID:         orderID,
		UserID:     userID,
		Status:     StatusNew,
		UploadedAt: time.Now(),
	}
	err := os.linker.storage.AddOrder(order)
	if err != nil {
		return err
	}

	return nil
}

func (os *orders) Get(orderID uint64) (*Order, error) {
	o, err := os.linker.storage.GetOrder(orderID)
	if err != nil {
		return nil, err
	}

	return o, nil
}

func (os *orders) GetUserOrders(userID uint64) ([]*Order, error) {
	ors, err := os.linker.storage.GetUserOrders(userID)
	if err != nil {
		return nil, err
	}

	return ors, nil
}
