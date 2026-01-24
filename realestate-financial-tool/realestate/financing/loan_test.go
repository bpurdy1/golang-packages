package financing

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestFunction(t *testing.T) {
	loan := NewLoan(300000, 60000, 500, 30, decimal.Zero)

	if loan.HomePrice.String() != "300000" {
		t.Errorf("HomePrice = %v, want %v", loan.HomePrice, "300000")
	}
	if loan.DownPayment.String() != "60000" {
		t.Errorf("DownPayment = %v, want %v", loan.DownPayment, "60000")
	}
	if loan.InterestRate.String() != "5%" {
		t.Errorf("InterestRate = %v, want %v", loan.InterestRate, "5%")
	}
	if loan.TermYears != 30 {
		t.Errorf("TermYears = %v, want %v", loan.TermYears, 30)
	}
}

func TestMonthlyPayment(t *testing.T) {
	loan := NewLoan(300_000, 60_000, 5, Term30Years, decimal.Zero)
	monthlyPayment := loan.MonthlyPayment()

	expected := decimal.NewFromFloat(1280.88) // Example expected value
	if monthlyPayment.Round(2).String() != expected.String() {
		t.Errorf("MonthlyPayment = %v, want %v", monthlyPayment.Round(2), expected)
	}
}
func TestTotalPayment(t *testing.T) {
	loan := NewLoan(300000, 60000, 5, Term30Years, decimal.Zero)
	totalPayment, _ := loan.GetTotalPayment()

	expected := decimal.NewFromFloat(461916.80) // Example expected value
	if totalPayment.Round(2).String() != expected.String() {
		t.Errorf("TotalPayment = %v, want %v", totalPayment.Round(2), expected)
	}
}
