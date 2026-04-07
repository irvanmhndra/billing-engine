# Billing Engine

A loan billing engine that generates repayment schedules, tracks outstanding balances, and detects borrower delinquency.

## Problem: Billing Engine (System Design and Abstraction)

Build a billing system for a 50-week loan of Rp 5,000,000 at a flat interest rate of 10% per annum.

- **Total repayable**: Rp 5,500,000
- **Weekly installment**: Rp 110,000

### Key Methods

| Method | Description |
|--------|-------------|
| `GetOutstanding()` | Returns the current outstanding balance. Returns 0 when fully repaid. |
| `IsDelinquent(at)` | Returns true if the borrower missed 2 consecutive due payments. |
| `MakePayment(amount, at)` | Applies a payment to the earliest unpaid week. Amount must match exactly. |
| `FormatSchedule()` | Returns the billing schedule (W1: 110000, W2: 110000, ...). |

### Assumptions

- **Delinquency threshold**: The problem says "miss 2 continuous repayments" in one place and "more than 2 weeks of non-payment" in another. I interpreted this as **>= 2 consecutive missed payments**, since the first description is more explicit.
- **Payment order**: Missed payments must be caught up in order — the oldest unpaid week is always paid first.
- **Money as integers**: All amounts use `int64` (Rupiah) to avoid floating-point precision issues.

## Run Demo

```bash
go run .
```

Walks through a realistic scenario: schedule output, payments, missed payments triggering delinquency, and catching up.

## Run Tests

```bash
go test ./billing/... -v
```

## Project Structure

```
main.go            # Runnable demo
billing/
  billing.go       # Core domain logic
  billing_test.go  # 20 tests covering all requirements
go.mod
README.md
```
