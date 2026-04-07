// Package billing implements a loan billing engine that tracks repayment
// schedules, outstanding balances, and borrower delinquency.
package billing

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	DefaultPrincipal  int64   = 5_000_000 // Rp 5,000,000
	DefaultWeeks              = 50
	DefaultAnnualRate float64 = 0.10 // 10% flat per annum
)

// PaymentStatus indicates whether a scheduled installment has been paid.
type PaymentStatus int8

const (
	StatusPending PaymentStatus = iota
	StatusPaid
)

// ScheduleEntry represents a single weekly installment.
type ScheduleEntry struct {
	Week   int
	Amount int64 // amount in Rupiah
	Status PaymentStatus
	PaidAt *time.Time
}

// Loan holds the repayment schedule and tracks payment state.
type Loan struct {
	ID        string
	StartDate time.Time
	Principal int64
	Weeks     int
	Rate      float64
	schedule  []*ScheduleEntry
}

// NewLoan creates a standard 50-week, 10%-per-annum flat-rate loan of Rp 5,000,000.
//
// Total repayable = 5,000,000 × 1.10 = 5,500,000.
// Weekly installment = 5,500,000 / 50 = 110,000.
func NewLoan(id string, startDate time.Time) *Loan {
	return newLoan(id, startDate, DefaultPrincipal, DefaultWeeks, DefaultAnnualRate)
}

func newLoan(id string, startDate time.Time, principal int64, weeks int, annualRate float64) *Loan {
	totalRepayable := int64(float64(principal) * (1 + annualRate))
	weeklyAmount := totalRepayable / int64(weeks)

	schedule := make([]*ScheduleEntry, weeks)
	for i := range schedule {
		schedule[i] = &ScheduleEntry{Week: i + 1, Amount: weeklyAmount, Status: StatusPending}
	}

	return &Loan{
		ID:        id,
		StartDate: startDate,
		Principal: principal,
		Weeks:     weeks,
		Rate:      annualRate,
		schedule:  schedule,
	}
}

// GetOutstanding returns the total amount (in Rupiah) the borrower still needs
// to repay. Returns 0 when the loan is fully repaid.
func (l *Loan) GetOutstanding() int64 {
	var total int64
	for _, e := range l.schedule {
		if e.Status == StatusPending {
			total += e.Amount
		}
	}
	return total
}

// IsDelinquent returns true if the borrower has missed 2 or more consecutive
// due payments as of the given reference time.
//
// Assumption: the problem statement says "miss 2 continuous repayments" in one
// place and "more than 2 weeks of non-payment" in another. We interpret this
// as ≥ 2 consecutive missed payments, since the first description is more
// explicit and aligns with the delinquency example given.
func (l *Loan) IsDelinquent(at time.Time) bool {
	due := l.weeksElapsed(at)
	if due < 2 {
		return false
	}

	missed := 0
	for i := due - 1; i >= 0; i-- {
		if l.schedule[i].Status == StatusPending {
			missed++
			if missed == 2 {
				return true
			}
		} else {
			break // consecutive streak broken
		}
	}
	return false
}

// MakePayment applies a payment to the earliest unpaid installment.
//
// The amount must exactly match the weekly installment — borrowers can only
// pay the full amount or nothing at all. To catch up on missed weeks, call
// MakePayment once per missed week.
func (l *Loan) MakePayment(amount int64, at time.Time) error {
	for _, e := range l.schedule {
		if e.Status == StatusPending {
			if amount != e.Amount {
				return fmt.Errorf("payment amount %d does not match required installment %d", amount, e.Amount)
			}
			t := at
			e.Status = StatusPaid
			e.PaidAt = &t
			return nil
		}
	}
	return errors.New("no pending payments: loan is fully repaid")
}

// GetSchedule returns a snapshot of the full repayment schedule.
func (l *Loan) GetSchedule() []ScheduleEntry {
	out := make([]ScheduleEntry, len(l.schedule))
	for i, e := range l.schedule {
		out[i] = *e
	}
	return out
}

// FormatSchedule returns the billing schedule in the format:
//
//	W1 : 110000
//	W2 : 110000
//	...
//	W50: 110000
func (l *Loan) FormatSchedule() string {
	var b strings.Builder
	for _, e := range l.schedule {
		fmt.Fprintf(&b, "W%-2d: %d\n", e.Week, e.Amount)
	}
	return b.String()
}

// weeksElapsed returns how many complete weeks have passed since the loan start,
// capped at the total loan duration.
func (l *Loan) weeksElapsed(at time.Time) int {
	hours := at.Sub(l.StartDate).Hours()
	if hours <= 0 {
		return 0
	}
	n := int(hours / (7 * 24))
	if n > l.Weeks {
		return l.Weeks
	}
	return n
}
