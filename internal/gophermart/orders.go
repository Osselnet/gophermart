package gophermart

import (
	"fmt"
	"github.com/Osselnet/gophermart.git/pkg/tool"
	"strconv"
	"sync"
	"time"
)

const (
	StatusNew        = "NEW"
	StatusProcessing = "PROCESSING"
	StatusInvalid    = "INVALID"
	StatusProcessed  = "PROCESSED"
)

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
	mu     sync.RWMutex
	byID   map[uint64]*Order
}

func newOrders(linker *GopherMart) *orders {
	return &orders{
		linker: linker,
		byID:   make(map[uint64]*Order),
	}
}

func (os *orders) Add(orderID, userID uint64) error {
	strOrderID := strconv.Itoa(int(orderID))
	if !tool.IsValid(strOrderID) {
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

	os.mu.Lock()
	os.byID[orderID] = order
	os.mu.Unlock()

	return nil
}

func (os *orders) Get(orderID uint64) (*Order, error) {
	var err error

	os.mu.RLock()
	o, ok := os.byID[orderID]
	os.mu.RUnlock()
	if !ok {
		o, err = os.linker.storage.GetOrder(orderID)
		if err != nil {
			return nil, err
		}

		os.mu.Lock()
		os.byID[orderID] = o
		os.mu.Unlock()
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