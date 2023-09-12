package gophermart

import (
	"fmt"
	"github.com/Osselnet/gophermart.git/pkg/luhn"
	"strconv"
	"time"
)

type Withdraw struct {
	OrderID     uint64
	UserID      uint64
	Sum         uint64
	ProcessedAt time.Time
}

type WithdrawProxy struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	UserID      uint64  `json:"-"`
	ProcessedAt string  `json:"processed_at"`
}

type withdrawals struct {
	linker *GopherMart
}

func newWithdrawals(linker *GopherMart) *withdrawals {
	return &withdrawals{
		linker: linker,
	}
}

func (ws *withdrawals) GetWithdrawals(userID uint64) ([]*Withdraw, error) {
	wds, err := ws.linker.storage.GetUserWithdrawals(userID)
	if err != nil {
		return nil, err
	}

	if len(wds) == 0 {
		return nil, ErrNoContent
	}

	return wds, nil
}

func (ws *withdrawals) Add(withdraw *Withdraw) error {
	strOrderID := strconv.Itoa(int(withdraw.OrderID))
	if !luhn.IsValid(strOrderID) {
		return ErrOrderInvalidFormat
	}

	wds, _ := ws.linker.storage.GetOrderWithdrawals(withdraw.OrderID)
	if wds != nil {
		if wds.UserID == withdraw.UserID {
			return fmt.Errorf("withdraw already recorded by this user")
		}
		return fmt.Errorf("withdraw already recorded by another user")
	}

	err := ws.linker.storage.AddWithdraw(withdraw)
	if err != nil {
		return err
	}

	return nil
}
