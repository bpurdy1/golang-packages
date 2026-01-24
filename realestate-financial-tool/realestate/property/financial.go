package property

import (
	"github.com/shopspring/decimal"

	"realestate-financial-tool/realestate/financing"
)

const (
	DefaultDownPaymentPercent = 20.0 // 20% default down payment
)

type Financial struct {
	AskingPrice   float64
	PurchasePrice float64
	DownPayment   float64
	LoanAmount    float64
	InterestRate  float64 // In basis points (e.g., 700 = 7%)
	LoanTermYears financing.LoanTerm
	Expenses      ExpensesMonthly
}

// Normalize ensures all derived fields are properly set
// - If DownPayment is 0, sets to 20% of PurchasePrice
// - If LoanAmount is 0, sets to PurchasePrice - DownPayment
func (f *Financial) Normalize() {
	if f.DownPayment == 0 && f.PurchasePrice > 0 {
		f.DownPayment = f.PurchasePrice * (DefaultDownPaymentPercent / 100)
	}
	if f.LoanAmount == 0 && f.PurchasePrice > 0 {
		f.LoanAmount = f.PurchasePrice - f.DownPayment
	}
}

func NewFinancial(
	askingPrice,
	purchasePrice,
	downPayment,
	interestRate float64, // basis points (700 = 7%)
	loanTermYears financing.LoanTerm) *Financial {

	// Default down payment to 20% if not provided
	if downPayment == 0 {
		downPayment = purchasePrice * (DefaultDownPaymentPercent / 100)
	}

	return &Financial{
		AskingPrice:   askingPrice,
		PurchasePrice: purchasePrice,
		DownPayment:   downPayment,
		LoanAmount:    purchasePrice - downPayment,
		InterestRate:  interestRate,
		LoanTermYears: loanTermYears,
	}
}

// percentage is in basis points (e.g., 5 for 5%)
func (f *Financial) SetLoanAmountPercentage(percentage float64) {
	if percentage < 0 || percentage > 100 {
		panic("percentage must be between 0 and 100")
	}
	f.LoanAmount = f.PurchasePrice * (percentage / 100)
}

func (f *Financial) SetExpenses(taxes, insurance, utilities, repairsMaintenance float64) {
	f.Expenses = ExpensesMonthly{
		Taxes:              taxes,
		Insurance:          insurance,
		Utilities:          utilities,
		RepairsMaintenance: repairsMaintenance,
	}
}

// InterestRatePercent returns interest rate as a percentage (e.g., 7.0 for 7%)
func (f *Financial) InterestRatePercent() float64 {
	return f.InterestRate / 100
}

func (f *Financial) Loan() *financing.Loan {
	f.Normalize() // Ensure derived fields are set
	return financing.NewLoan(
		int64(f.PurchasePrice),
		int64(f.DownPayment),
		f.InterestRatePercent(), // Convert basis points to percent for NewLoan
		f.LoanTermYears,
		decimal.Zero,
	)
}

// PlotLoan generates an amortization chart and returns the HTML filename
func (f *Financial) PlotLoan() (string, error) {
	loan := f.Loan()
	return loan.Plot()
}

// PlotLoanSummary generates a summary pie chart showing principal vs interest
func (f *Financial) PlotLoanSummary() (string, error) {
	loan := f.Loan()
	return loan.PlotSummary()
}

// LoanSummary returns a formatted string with all loan totals
func (f *Financial) LoanSummary() (string, error) {
	loan := f.Loan()
	return loan.LoanSummary()
}

type ExpensesMonthly struct {
	Taxes              float64
	Insurance          float64
	PMI                float64
	Utilities          float64
	RepairsMaintenance float64
	ManagementFee      float64
	OtherExpenses      float64
	CapitalReserves    float64
	VacancyRate        float64
}

func (e *ExpensesMonthly) TotalYearly() float64 {
	expenses := e.TotalMonthly() * 12
	return expenses
}
func (e *ExpensesMonthly) TotalMonthly() float64 {
	return e.Taxes +
		e.Insurance +
		e.Utilities +
		e.RepairsMaintenance +
		e.ManagementFee +
		e.OtherExpenses +
		e.CapitalReserves +
		e.PMI
}

func (e *ExpensesMonthly) VacancyCost(yearlyIncome float64) float64 {
	return yearlyIncome * e.VacancyRate
}
