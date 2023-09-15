package gophermart

type Balance struct {
	UserID    uint64
	Current   uint64
	Withdrawn uint64
}

type BalanceProxy struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type balances struct {
	linker *GopherMart
}

func newBalance(linker *GopherMart) *balances {
	return &balances{
		linker: linker,
	}
}

func (bs *balances) Get(userID uint64) (Balance, error) {
	return bs.linker.storage.GetBalance(userID)
}
