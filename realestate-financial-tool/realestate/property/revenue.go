package property

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

// Calculate the monthly loan payment using the standard amortization formula.

type Metrics struct {

	// Monthly stats
	MonthlyRentalIncome float64 `json:"monthly_rental_income"`
	MonthlyExpenses     float64 `json:"monthly_expenses"`
	MonthlyNetCashFlow  float64 `json:"monthly_net_cash_flow"`

	// New fields for explicit negative values
	MonthlyLoanPayment   float64 `json:"monthly_loan_payment"`   // always negative
	TotalMonthlyExpenses float64 `json:"total_monthly_expenses"` // sum of loan payment and expenses, always negative

	// Annual stats
	AnnualRentalIncome  float64 `json:"annual_rental_income"`
	AnnualExpenses      float64 `json:"annual_expenses"`
	AnnualNetIncome     float64 `json:"annual_net_income"`
	AnnualNetCashFlow   float64 `json:"annual_net_cash_flow"`    // after loan payments
	AnnualCashOnCashYoY string  `json:"annual_cash_on_cash_yoy"` // now a string percentage
	AnnualPaymentAmount float64 `json:"annual_payment_amount"`
	AnnualTotalCost     float64 `json:"annual_total_cost"`

	// Final annual
	AnnualCashFlow float64 `json:"annual_cash_flow"` // before loan payments
	CapRate        float64 `json:"cap_rate"`
}

func (m *Metrics) String() string {
	b, _ := json.MarshalIndent(m, "", "  ")
	return string(b)
}

// CalculateMetrics computes all requested metrics for a property.
func CalculateMetrics(p *Property) *Metrics {
	var totalMonthlyRent decimal.Decimal
	for _, unit := range p.Units {
		totalMonthlyRent = totalMonthlyRent.Add(decimal.NewFromFloat(unit.Rent))
	}
	monthlyRentalIncome := totalMonthlyRent
	annualRentalIncome := totalMonthlyRent.Mul(decimal.NewFromInt(12))

	financial := p.Financial
	expenses := financial.Expenses
	monthlyExpenses := decimal.NewFromFloat(expenses.TotalMonthly())
	annualExpenses := decimal.NewFromFloat(expenses.TotalYearly())

	monthlyPayment := financial.Loan().MonthlyPayment()
	annualPayment := monthlyPayment.Mul(decimal.NewFromInt(12))

	// Ensure expenses and loan payment are negative
	monthlyLoanPayment := monthlyPayment.Neg().Abs().Neg()             // always negative
	monthlyExpensesNeg := monthlyExpenses.Neg().Abs().Neg()            // always negative
	totalMonthlyExpenses := monthlyLoanPayment.Add(monthlyExpensesNeg) // sum, always negative

	annualNetIncome := annualRentalIncome.Sub(annualExpenses)
	capRate := 0.0
	if p.Financial.PurchasePrice > 0 {
		capRate = annualNetIncome.Div(decimal.NewFromFloat(p.Financial.PurchasePrice)).InexactFloat64()
	}

	annualNetCashFlow := annualNetIncome.Sub(annualPayment) // after loan payments
	monthlyNetCashFlow := monthlyRentalIncome.Sub(monthlyExpenses).Sub(monthlyPayment)
	annualTotalCost := annualExpenses.Add(annualPayment)
	annualCashOnCash := "0%"
	if p.Financial.DownPayment > 0 {
		coc := annualNetCashFlow.Div(decimal.NewFromFloat(p.Financial.DownPayment)).Mul(decimal.NewFromInt(100))
		annualCashOnCash = coc.Round(2).String() + "%"
	}

	return &Metrics{
		CapRate:              capRate,
		MonthlyRentalIncome:  monthlyRentalIncome.Round(2).InexactFloat64(),
		MonthlyExpenses:      monthlyExpensesNeg.Round(2).InexactFloat64(), // always negative
		MonthlyNetCashFlow:   monthlyNetCashFlow.Round(2).InexactFloat64(),
		MonthlyLoanPayment:   monthlyLoanPayment.Round(2).InexactFloat64(),   // always negative
		TotalMonthlyExpenses: totalMonthlyExpenses.Round(2).InexactFloat64(), // always negative
		AnnualRentalIncome:   annualRentalIncome.Round(2).InexactFloat64(),
		AnnualNetIncome:      annualNetIncome.Round(2).InexactFloat64(),
		AnnualExpenses:       annualExpenses.Round(2).InexactFloat64(),
		AnnualNetCashFlow:    annualNetCashFlow.Round(2).InexactFloat64(), // after loan payments
		AnnualCashOnCashYoY:  annualCashOnCash,                            // now a string percentage
		AnnualPaymentAmount:  annualPayment.Round(2).InexactFloat64(),
		AnnualTotalCost:      annualTotalCost.Round(2).InexactFloat64(),
		AnnualCashFlow:       annualNetCashFlow.Round(2).InexactFloat64(), // now matches net cash flow after loan payments
	}
}

func MonthlyRentalIncome() {
}
