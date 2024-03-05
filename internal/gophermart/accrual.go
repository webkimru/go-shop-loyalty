package gophermart

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/api"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/logger"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/models"
	"net/http"
	"sync"
	"time"
)

func CheckAccrual(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	defer func() {
		logger.Log.Infoln("defer after ctrl + c")
	}()

	logger.Log.Infoln("Starting accrual checker")
	for {
		job, open := <-api.Repo.Jobs
		if !open {
			return
		}

		time.Sleep(time.Duration(app.AccrualPollInterval) * time.Second)

		logger.Log.Infoln("Checking accrual:", "job.UserID", job.UserID, "job.Number", job.Number)

		// Ходим в accrual service
		accrualResult, err := GetAccrual(ctx, job.Number)
		if err != nil {
			logger.Log.Errorln(err)
			continue
		}

		logger.Log.Infoln(
			"There is a accrual:",
			"accrual.Number", accrualResult.Number,
			"accrual.Status", accrualResult.Status,
			"accrual.Accrual", accrualResult.Accrual,
		)

		// update balance
		balance := models.Balance{}
		balance.Current = models.Money(accrualResult.Accrual.Set())
		err = api.Repo.Store.SetBalance(ctx, balance, job.UserID)
		if err != nil {
			logger.Log.Errorln("failed SetBalance()=", err)
		}

		// update order status
		order := models.Order{}
		order.UserID = job.UserID
		order.Number = accrualResult.Number
		order.Accrual = models.Money(accrualResult.Accrual.Set())
		order.Status = ConvertStatus(string(accrualResult.Status))
		err = api.Repo.Store.UpdateOrder(ctx, order)
		if err != nil {
			logger.Log.Errorln("failed UpdateOrder()=", err)
		}

		// если статусы нефинальные, то возвращаем в канал
		switch accrualResult.Status {
		case models.AccrualStateRegistered, models.AccrualStateProcessing:
			accrualRequest := models.AccrualRequest{
				Number: accrualResult.Number,
				UserID: job.UserID,
			}
			api.Repo.Jobs <- accrualRequest
		}
	}
}

func GetAccrual(ctx context.Context, order string) (*models.AccrualResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/orders/%s", app.AccrualSystemAddress, order), nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response code expected %d, but got %d", http.StatusOK, resp.StatusCode)
	}

	var accrualResponse models.AccrualResponse
	if err := json.NewDecoder(resp.Body).Decode(&accrualResponse); err != nil {
		return nil, err
	}

	return &accrualResponse, nil
}

func ConvertStatus(status string) models.OrderState {
	if status == string(models.AccrualStateRegistered) {
		return models.OrderStateNew
	}
	return models.OrderState(status)
}
