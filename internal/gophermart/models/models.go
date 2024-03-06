package models

type User struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Password  string `json:"password"`
	CreatedAt string `json:"created_at"`
}

type OrderState string

const (
	OrderStateNew        OrderState = "NEW"        // заказ загружен в систему, но не попал в обработку
	OrderStateProcessing OrderState = "PROCESSING" // вознаграждение за заказ рассчитывается
	OrderStateInvalid    OrderState = "INVALID"    // система расчёта вознаграждений отказала в расчёте
	OrderStateProcessed  OrderState = "PROCESSED"  // данные по заказу проверены и информация о расчёте успешно получена
)

type Order struct {
	Number    string     `json:"number"`
	UserID    int64      `json:"-"`
	Accrual   Money      `json:"accrual,omitempty"`
	Status    OrderState `json:"status"`
	CreatedAt string     `json:"uploaded_at"`
}

func (o Order) IsValid() bool {
	// алгоритм Луна - https://ru.wikipedia.org/wiki/%D0%90%D0%BB%D0%B3%D0%BE%D1%80%D0%B8%D1%82%D0%BC_%D0%9B%D1%83%D0%BD%D0%B0
	sum := 0
	parity := len(o.Number) % 2

	for i, value := range o.Number {
		digit := int(value - '0')

		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	return sum%10 == 0
}

type Balance struct {
	UserID    int64  `json:"-"`
	Current   Money  `json:"current"`
	Withdrawn Money  `json:"withdrawn"`
	CreatedAt string `json:"-"`
}

type Money float32

func (m Money) Set() int64 {
	return int64(m * 100)
}

func (m Money) Get() float32 {
	return float32(m) / 100
}

type AccrualState string

const (
	AccrualStateRegistered AccrualState = "REGISTERED" // заказ загружен в систему, но не попал в обработку
	AccrualStateProcessing AccrualState = "PROCESSING" // вознаграждение за заказ рассчитывается
	AccrualStateInvalid    AccrualState = "INVALID"    // система расчёта вознаграждений отказала в расчёте
	AccrualStateProcessed  AccrualState = "PROCESSED"  // данные по заказу проверены и информация о расчёте успешно получена
)

type AccrualRequest struct {
	Number string
	UserID int64
}

type AccrualResponse struct {
	Number  string       `json:"order"`
	Status  AccrualState `json:"status"`
	Accrual Money        `json:"accrual"`
}

type Withdrawal struct {
	Order     string `json:"order"`
	UserID    int64  `json:"-"`
	Sum       Money  `json:"sum"`
	CreatedAt string `json:"processed_at"`
}
