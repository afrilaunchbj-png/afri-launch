package domain

import "testing"

func TestCreditAccountAvailable(t *testing.T) {
	cases := []struct {
		name     string
		balance  int64
		reserved int64
		want     int64
	}{
		{"simple", 100, 30, 70},
		{"zero reserved", 100, 0, 100},
		{"fully reserved", 100, 100, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a := CreditAccount{Balance: c.balance, Reserved: c.reserved}
			if got := a.Available(); got != c.want {
				t.Fatalf("Available() = %d, want %d", got, c.want)
			}
		})
	}
}
