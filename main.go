package main

import (
	"fmt"
	"time"

	"amartha-test/billing"
)

func main() {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	loan := billing.NewLoan("L100", start)

	fmt.Println("=== Billing Schedule ===")
	fmt.Print(loan.FormatSchedule())

	fmt.Printf("\nOutstanding: %d\n", loan.GetOutstanding())

	// Borrower pays weeks 1–3 on time
	week := func(n int) time.Time {
		return start.Add(time.Duration(n) * 7 * 24 * time.Hour)
	}
	for i := 1; i <= 3; i++ {
		_ = loan.MakePayment(110_000, week(i))
	}
	fmt.Printf("\nAfter 3 payments:\n")
	fmt.Printf("  Outstanding: %d\n", loan.GetOutstanding())
	fmt.Printf("  Delinquent:  %v\n", loan.IsDelinquent(week(3)))

	// Borrower misses weeks 4 and 5
	fmt.Printf("\nAfter missing weeks 4 and 5:\n")
	fmt.Printf("  Outstanding: %d\n", loan.GetOutstanding())
	fmt.Printf("  Delinquent:  %v\n", loan.IsDelinquent(week(5)))

	// Borrower catches up
	_ = loan.MakePayment(110_000, week(6)) // pays week 4
	_ = loan.MakePayment(110_000, week(6)) // pays week 5
	fmt.Printf("\nAfter catching up in week 6:\n")
	fmt.Printf("  Outstanding: %d\n", loan.GetOutstanding())
	fmt.Printf("  Delinquent:  %v\n", loan.IsDelinquent(week(6)))
}
