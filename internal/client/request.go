package client

import (
	"context"
	"fmt"
	"github.com/Osselnet/gophermart.git/internal/gophermart"
	"github.com/go-resty/resty/v2"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

type queueOrder struct {
	*Queue
	ctx   context.Context
	order *gophermart.Order
}

func (qo *queueOrder) Do() error {
	ctx, cancel := context.WithTimeout(qo.ctx, 60*time.Second)
	defer cancel()
	order := qo.order
	url := fmt.Sprintf("%s%d", qo.url, order.ID)
	log.Println("[DEBUG] Making request:", url)

	ao := &accrualOrder{}
	client := resty.New()
	resp, err := client.R().
		SetHeader("Accept", "*/*").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Length", "0").
		SetContext(ctx).
		SetResult(&ao).
		Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode() == http.StatusInternalServerError {
		return fmt.Errorf("internal server error, status code %d", resp.StatusCode())
	}

	if resp.StatusCode() == http.StatusTooManyRequests {
		retryAfter, err := strconv.Atoi(resp.Header().Get("Retry-After"))
		atomic.StoreUint32(&qo.sleep, uint32(retryAfter))
		if err != nil {
			return fmt.Errorf("[ERROR] Error failed to parse Retry-After value, err: %w", err)
		}
		fmt.Println("[WARNING] Too many requests detected", string(resp.Body()))
		return gophermart.ErrTooManyRequests
	}

	if resp.StatusCode() == http.StatusNoContent {
		log.Printf("[WARNING] No content for order %d\n", order.ID)
		return nil
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unknown status code %d", resp.StatusCode())
	}

	if fmt.Sprint(order.ID) != ao.Order {
		log.Printf("[WARNING] Order ID not match, want %d, got %s\n", order.ID, ao.Order)
		return nil
	}

	if order.Status == ao.Status && order.Status == gophermart.StatusProcessing {
		log.Printf("[DEBUG] Order %d already in processing\n", order.ID)
		return nil
	}

	if !gophermart.IsValidStatus(ao.Status) {
		log.Printf("[WARNING] Unknown status detected: %s\n", ao.Status)
		return nil
	}
	return qo.updateOrder(ao, qo)
}
