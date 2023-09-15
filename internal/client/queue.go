package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/Osselnet/gophermart.git/internal/gophermart"
	"golang.org/x/sync/errgroup"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	limitDefault = 1000
	limitDelta   = 1
)

type accrualOrder struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type Queue struct {
	url       string
	storage   gophermart.Storer
	limit     uint32
	sleep     uint32
	needSleep int32
	pool      map[uint64]*gophermart.Order
}

func NewQueue(st gophermart.Storer, addr string) *Queue {

	return &Queue{
		limit:   limitDefault,
		url:     addr + "/api/orders/",
		storage: st,
	}
}

func (q *Queue) updatePool() {
	limit := atomic.LoadUint32(&q.limit)

	ors, err := q.storage.GetPullOrders(limit)
	if err != nil {
		log.Println("[ERROR] Failed to get orders for pool -", err)
		return
	}

	pool := make(map[uint64]*gophermart.Order, limit)

	for k, order := range ors {
		pool[k] = order
	}
	q.pool = pool
	log.Printf("[DEBUG] Orders pool updated, now in pool: %d", len(q.pool))
}

func (q *Queue) updateOrder(ao *accrualOrder, qo *queueOrder) error {
	qo.order.Status = ao.Status
	qo.order.Accrual = uint64(ao.Accrual * 100)

	if err := q.storage.UpdateOrder(qo.order); err != nil {
		return fmt.Errorf("failed to update order ID %d - %w", qo.order.ID, err)
	}
	log.Printf("[DEBUG] Order successfully updated: order %v\n", qo.order)

	return nil
}

func (q *Queue) processor(ctx context.Context) {
	for {
		q.updatePool()

		g, _ := errgroup.WithContext(ctx)
		for _, order := range q.pool {
			w := &queueOrder{Queue: q, ctx: ctx, order: order}
			g.Go(w.Do)
		}
		err := g.Wait()
		if err != nil {
			atomic.StoreInt32(&q.needSleep, 1)
			if !errors.Is(err, gophermart.ErrTooManyRequests) {
				atomic.StoreUint32(&q.limit, limitDefault)
			}
			log.Println("[ERROR] Accrual service request failed -", err)
		}

		sleep := 1 * time.Second
		if atomic.LoadInt32(&q.needSleep) == 0 {
			atomic.AddUint32(&q.limit, limitDelta)
		} else {
			sleep = time.Duration(atomic.LoadUint32(&q.sleep)) * time.Second
			atomic.StoreInt32(&q.needSleep, 0)
		}
		log.Println("[DEBUG] Got new limit:", atomic.LoadUint32(&q.limit))
		log.Printf("[DEBUG] Sleeping for %s seconds\n", sleep)

		select {
		case <-ctx.Done():
			return
		case <-time.After(sleep):
		}
	}
}

func (q *Queue) Start() {
	rand.Seed(time.Now().UnixNano())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-sig
		cancel()
	}()

	q.processor(ctx)
}
