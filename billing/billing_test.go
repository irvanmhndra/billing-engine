package billing

import (
	"strings"
	"testing"
	"time"
)

var startDate = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

// week returns the time exactly n complete weeks after the loan start.
func week(n int) time.Time {
	return startDate.Add(time.Duration(n) * 7 * 24 * time.Hour)
}

// --- Schedule ---

func TestNewLoan_ScheduleHas50Entries(t *testing.T) {
	l := NewLoan("L100", startDate)
	if got := len(l.GetSchedule()); got != 50 {
		t.Fatalf("expected 50 schedule entries, got %d", got)
	}
}

func TestNewLoan_WeeklyInstallmentIs110000(t *testing.T) {
	l := NewLoan("L100", startDate)
	for _, e := range l.GetSchedule() {
		if e.Amount != 110_000 {
			t.Errorf("week %d: expected installment 110000, got %d", e.Week, e.Amount)
		}
	}
}

func TestNewLoan_AllInstallmentsArePending(t *testing.T) {
	l := NewLoan("L100", startDate)
	for _, e := range l.GetSchedule() {
		if e.Status != StatusPending {
			t.Errorf("week %d: expected StatusPending, got %v", e.Week, e.Status)
		}
	}
}

// --- GetOutstanding ---

func TestGetOutstanding_InitialIs5500000(t *testing.T) {
	l := NewLoan("L100", startDate)
	if got := l.GetOutstanding(); got != 5_500_000 {
		t.Errorf("expected 5500000, got %d", got)
	}
}

func TestGetOutstanding_DecreasesAfterPayment(t *testing.T) {
	l := NewLoan("L100", startDate)
	_ = l.MakePayment(110_000, week(1))
	_ = l.MakePayment(110_000, week(2))

	want := int64(5_500_000 - 2*110_000)
	if got := l.GetOutstanding(); got != want {
		t.Errorf("expected %d, got %d", want, got)
	}
}

func TestGetOutstanding_ZeroWhenFullyRepaid(t *testing.T) {
	l := NewLoan("L100", startDate)
	for i := 1; i <= 50; i++ {
		if err := l.MakePayment(110_000, week(i)); err != nil {
			t.Fatalf("payment %d failed: %v", i, err)
		}
	}
	if got := l.GetOutstanding(); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

// --- MakePayment ---

func TestMakePayment_AppliesToOldestPendingWeek(t *testing.T) {
	l := NewLoan("L100", startDate)
	_ = l.MakePayment(110_000, week(1))

	sched := l.GetSchedule()
	if sched[0].Status != StatusPaid {
		t.Error("week 1 should be paid")
	}
	if sched[1].Status != StatusPending {
		t.Error("week 2 should still be pending")
	}
}

func TestMakePayment_RecordsPaidAt(t *testing.T) {
	l := NewLoan("L100", startDate)
	payTime := week(1)
	_ = l.MakePayment(110_000, payTime)

	sched := l.GetSchedule()
	if sched[0].PaidAt == nil || !sched[0].PaidAt.Equal(payTime) {
		t.Error("PaidAt should record the payment timestamp")
	}
}

func TestMakePayment_WrongAmountReturnsError(t *testing.T) {
	l := NewLoan("L100", startDate)
	if err := l.MakePayment(100_000, week(1)); err == nil {
		t.Error("expected error for wrong amount, got nil")
	}
}

func TestMakePayment_AfterFullRepaymentReturnsError(t *testing.T) {
	l := NewLoan("L100", startDate)
	for i := 1; i <= 50; i++ {
		_ = l.MakePayment(110_000, week(i))
	}
	if err := l.MakePayment(110_000, week(51)); err == nil {
		t.Error("expected error when no pending payments remain")
	}
}

func TestMakePayment_CatchUpMultipleMissedWeeks(t *testing.T) {
	l := NewLoan("L100", startDate)
	// Missed weeks 1 and 2; borrower catches up in week 3
	_ = l.MakePayment(110_000, week(3)) // pays week 1
	_ = l.MakePayment(110_000, week(3)) // pays week 2

	sched := l.GetSchedule()
	if sched[0].Status != StatusPaid {
		t.Error("week 1 should be paid after catch-up")
	}
	if sched[1].Status != StatusPaid {
		t.Error("week 2 should be paid after catch-up")
	}
}

// --- IsDelinquent ---

func TestIsDelinquent_FalseBeforeTwoWeeksElapse(t *testing.T) {
	l := NewLoan("L100", startDate)
	if l.IsDelinquent(week(1)) {
		t.Error("cannot be delinquent with fewer than 2 weeks elapsed")
	}
}

func TestIsDelinquent_FalseAtLoanStart(t *testing.T) {
	l := NewLoan("L100", startDate)
	if l.IsDelinquent(startDate) {
		t.Error("cannot be delinquent at loan start")
	}
}

func TestIsDelinquent_FalseWhenAllPaymentsMade(t *testing.T) {
	l := NewLoan("L100", startDate)
	_ = l.MakePayment(110_000, week(1))
	_ = l.MakePayment(110_000, week(2))

	if l.IsDelinquent(week(2)) {
		t.Error("borrower who paid all dues should not be delinquent")
	}
}

func TestIsDelinquent_TrueAfterTwoConsecutiveMissedPayments(t *testing.T) {
	l := NewLoan("L100", startDate)
	// No payments made; 2 weeks elapsed
	if !l.IsDelinquent(week(2)) {
		t.Error("borrower who missed 2 consecutive payments should be delinquent")
	}
}

func TestIsDelinquent_FalseAfterOnlyOneMissedPayment(t *testing.T) {
	l := NewLoan("L100", startDate)
	_ = l.MakePayment(110_000, week(1)) // paid week 1, missed week 2

	if l.IsDelinquent(week(2)) {
		t.Error("missing only 1 consecutive payment should not trigger delinquency")
	}
}

func TestIsDelinquent_FalseAfterCatchingUp(t *testing.T) {
	l := NewLoan("L100", startDate)
	// Was delinquent (missed weeks 1 and 2)
	if !l.IsDelinquent(week(2)) {
		t.Fatal("precondition: should be delinquent before catch-up")
	}

	_ = l.MakePayment(110_000, week(3)) // pay week 1
	_ = l.MakePayment(110_000, week(3)) // pay week 2

	if l.IsDelinquent(week(3)) {
		t.Error("borrower who caught up should no longer be delinquent")
	}
}

func TestIsDelinquent_TrueMidLoanAfterTwoMissed(t *testing.T) {
	l := NewLoan("L100", startDate)
	for i := 1; i <= 3; i++ {
		_ = l.MakePayment(110_000, week(i))
	}
	// Miss weeks 4 and 5

	if !l.IsDelinquent(week(5)) {
		t.Error("should be delinquent after missing 2 consecutive mid-loan payments")
	}
}

func TestIsDelinquent_FalseWhenMissedPaymentsAreNonConsecutive(t *testing.T) {
	l := NewLoan("L100", startDate)
	_ = l.MakePayment(110_000, week(1)) // paid week 1
	// missed week 2
	_ = l.MakePayment(110_000, week(3)) // paid week 2 (catch-up)
	// missed week 3

	// Only 1 consecutive missed at the end → not delinquent
	if l.IsDelinquent(week(3)) {
		t.Error("non-consecutive missed payments should not trigger delinquency")
	}
}

// --- FormatSchedule ---

func TestFormatSchedule_ContainsAllWeeks(t *testing.T) {
	l := NewLoan("L100", startDate)
	s := l.FormatSchedule()

	if !strings.Contains(s, "W1 : 110000") {
		t.Error("schedule should contain W1 : 110000")
	}
	if !strings.Contains(s, "W50: 110000") {
		t.Error("schedule should contain W50: 110000")
	}

	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) != 50 {
		t.Errorf("expected 50 lines, got %d", len(lines))
	}
}
