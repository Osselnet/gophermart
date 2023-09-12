package test

import (
	"github.com/Osselnet/gophermart.git/internal/gophermart"
	"github.com/Osselnet/gophermart.git/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGopherMart_GetOrders(t *testing.T) {
	rTime := time.Now()

	tests := []struct {
		name    string
		userID  uint64
		ors     []*gophermart.Order
		want    []*gophermart.OrderProxy
		wantErr bool
	}{
		{
			name:   "status Ok",
			userID: 189,
			ors: []*gophermart.Order{
				{
					ID:         6767584380420,
					UserID:     189,
					Status:     "PROCESSED",
					Accrual:    79998,
					UploadedAt: rTime,
				},
			},
			want: []*gophermart.OrderProxy{
				{
					Number:     "6767584380420",
					Status:     "PROCESSED",
					Accrual:    799.98,
					UploadedAt: rTime.Format(time.RFC3339),
				},
			},
			wantErr: true,
		},
		{
			name:   "status NotOk",
			userID: 189,
			ors: []*gophermart.Order{
				{
					ID:         6767584380420,
					UserID:     189,
					Status:     "PROCESSED",
					Accrual:    79998,
					UploadedAt: rTime,
				},
			},
			want: []*gophermart.OrderProxy{
				{
					Number:     "6767584380420",
					Status:     "PROCESSED",
					Accrual:    79.98,
					UploadedAt: rTime.Format(time.RFC3339),
				},
			},
			wantErr: false,
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorer(ctrl)
	gm := gophermart.New(m)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.EXPECT().GetUserOrders(tt.userID).Return(tt.ors, nil)
			got, err := gm.GetOrders(tt.userID)
			assert.NoError(t, err)
			if tt.wantErr {
				assert.Equal(t, tt.want, got)
			} else {
				assert.NotEqual(t, tt.want, got)
			}
		})
	}
}

func TestGopherMart_PostWithdraw(t *testing.T) {
	tests := []struct {
		name    string
		wpr     *gophermart.WithdrawProxy
		bw      *gophermart.Withdraw
		wantErr bool
	}{
		{
			name: "status Ok",
			bw: &gophermart.Withdraw{
				OrderID: 303653406,
				UserID:  173,
				Sum:     26061,
			},
			wpr: &gophermart.WithdrawProxy{
				Order:  "303653406",
				UserID: 173,
				Sum:    260.61,
			},
			wantErr: true,
		},
		{
			name: "status NotOk",
			bw: &gophermart.Withdraw{
				OrderID: 303656,
				UserID:  173,
				Sum:     26061,
			},
			wpr: &gophermart.WithdrawProxy{
				Order:  "303656",
				UserID: 173,
				Sum:    260.61,
			},
			wantErr: false,
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorer(ctrl)
	gm := gophermart.New(m)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				m.EXPECT().GetOrderWithdrawals(tt.bw.OrderID).Return(nil, nil)
				m.EXPECT().AddWithdraw(tt.bw).Return(nil)
			}
			err := gm.PostWithdraw(tt.wpr)
			if tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestGopherMart_GetWithdrawals(t *testing.T) {
	type args struct {
		userID uint64
	}
	rTime := time.Now()

	tests := []struct {
		name    string
		args    args
		wds     []*gophermart.Withdraw
		want    []*gophermart.WithdrawProxy
		wantErr bool
	}{
		{
			name: "status Ok",
			args: args{
				173,
			},
			wds: []*gophermart.Withdraw{
				{
					OrderID:     303653406,
					UserID:      173,
					Sum:         26061,
					ProcessedAt: rTime,
				},
			},
			want: []*gophermart.WithdrawProxy{
				{
					Order:       "303653406",
					Sum:         260.61,
					ProcessedAt: rTime.Format(time.RFC3339),
				},
			},
			wantErr: true,
		},
		{
			name: "status NotOk",
			args: args{
				173,
			},
			wds: []*gophermart.Withdraw{
				{
					OrderID:     67625566,
					UserID:      185,
					Sum:         62154,
					ProcessedAt: rTime,
				},
			},
			want: []*gophermart.WithdrawProxy{
				{
					Order:       "67625566",
					Sum:         260.61,
					ProcessedAt: rTime.Format(time.RFC3339),
				},
			},
			wantErr: false,
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorer(ctrl)
	gm := gophermart.New(m)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.EXPECT().GetUserWithdrawals(tt.args.userID).Return(tt.wds, nil)
			got, err := gm.GetWithdrawals(tt.args.userID)
			assert.NoError(t, err)
			if tt.wantErr {
				assert.Equal(t, tt.want, got)
			} else {
				assert.NotEqual(t, tt.want, got)
			}
		})
	}
}

func TestGopherMart_GetBalance(t *testing.T) {
	type args struct {
		userID uint64
	}
	tests := []struct {
		name    string
		args    args
		bal     gophermart.Balance
		want    *gophermart.BalanceProxy
		wantErr bool
	}{
		{
			name: "status Ok",
			args: args{
				173,
			},
			bal: gophermart.Balance{
				UserID:  173,
				Current: 145996,
			},
			want: &gophermart.BalanceProxy{
				Current: 1459.96,
			},
			wantErr: true,
		},
		{
			name: "status NotOk",
			args: args{
				173,
			},
			bal: gophermart.Balance{
				UserID:  173,
				Current: 145996,
			},
			want: &gophermart.BalanceProxy{
				Current:   1459.96,
				Withdrawn: 1.0,
			},
			wantErr: false,
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorer(ctrl)
	gm := gophermart.New(m)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.EXPECT().GetBalance(tt.args.userID).Return(tt.bal, nil)
			got, err := gm.GetBalance(tt.args.userID)
			assert.NoError(t, err)
			if tt.wantErr {
				assert.Equal(t, tt.want, got)
			} else {
				assert.NotEqual(t, tt.want, got)
			}
		})
	}
}
