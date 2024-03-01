package models

import "strconv"

type User struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Password  string `json:"password"`
	CreatedAt string `json:"created_at"`
}

type Order int64

func (o Order) IsValid() bool {
	// алгоритм Луна - https://ru.wikipedia.org/wiki/%D0%90%D0%BB%D0%B3%D0%BE%D1%80%D0%B8%D1%82%D0%BC_%D0%9B%D1%83%D0%BD%D0%B0
	ccn := strconv.Itoa(int(o))
	sum := 0
	parity := len(ccn) % 2

	for i, value := range ccn {
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
