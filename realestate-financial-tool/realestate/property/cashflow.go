package property

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// CashFlowAnalysis contains comprehensive cash flow analysis for a property
type CashFlowAnalysis struct {
	// Income
	MonthlyGrossIncome decimal.Decimal
	AnnualGrossIncome  decimal.Decimal

	// Expenses (excluding mortgage)
	MonthlyExpenses decimal.Decimal
	AnnualExpenses  decimal.Decimal

	// Net Operating Income (before mortgage)
	MonthlyNOI decimal.Decimal
	AnnualNOI  decimal.Decimal

	// Mortgage
	MonthlyMortgage decimal.Decimal
	AnnualMortgage  decimal.Decimal

	// Cash Flow (after mortgage)
	MonthlyCashFlow decimal.Decimal
	AnnualCashFlow  decimal.Decimal

	// Key Metrics
	CapRate        decimal.Decimal // NOI / Purchase Price
	CashOnCash     decimal.Decimal // Annual Cash Flow / Down Payment
	GRM            decimal.Decimal // Gross Rent Multiplier: Price / Annual Rent
	DSCR           decimal.Decimal // Debt Service Coverage Ratio: NOI / Mortgage
	BreakEvenRatio decimal.Decimal // (Expenses + Mortgage) / Gross Income

	// Investment Summary
	TotalInvestment decimal.Decimal // Down payment + closing costs
	PurchasePrice   decimal.Decimal
	DownPayment     decimal.Decimal
	LoanAmount      decimal.Decimal
	InterestRate    decimal.Decimal // Interest rate as percentage (e.g., 6.0 for 6%)
	LoanTermYears   int             // Loan term in years
}

// AnalyzeCashFlow performs comprehensive cash flow analysis on a property
func AnalyzeCashFlow(p *Property) *CashFlowAnalysis {
	analysis := &CashFlowAnalysis{}

	// Calculate gross income from units
	var monthlyIncome decimal.Decimal
	for _, unit := range p.Units {
		monthlyIncome = monthlyIncome.Add(decimal.NewFromFloat(unit.Rent))
	}
	analysis.MonthlyGrossIncome = monthlyIncome
	analysis.AnnualGrossIncome = monthlyIncome.Mul(decimal.NewFromInt(12))

	// Get expenses
	expenses := p.Financial.Expenses
	analysis.MonthlyExpenses = decimal.NewFromFloat(expenses.TotalMonthly())
	analysis.AnnualExpenses = decimal.NewFromFloat(expenses.TotalYearly())

	// Calculate NOI (Net Operating Income)
	analysis.MonthlyNOI = analysis.MonthlyGrossIncome.Sub(analysis.MonthlyExpenses)
	analysis.AnnualNOI = analysis.AnnualGrossIncome.Sub(analysis.AnnualExpenses)

	// Get mortgage payment
	loan := p.Financial.Loan()
	analysis.MonthlyMortgage = loan.MonthlyPayment().Abs()
	analysis.AnnualMortgage = analysis.MonthlyMortgage.Mul(decimal.NewFromInt(12))

	// Calculate cash flow (after mortgage)
	analysis.MonthlyCashFlow = analysis.MonthlyNOI.Sub(analysis.MonthlyMortgage)
	analysis.AnnualCashFlow = analysis.AnnualNOI.Sub(analysis.AnnualMortgage)

	// Investment details
	analysis.PurchasePrice = decimal.NewFromFloat(p.Financial.PurchasePrice)
	analysis.DownPayment = decimal.NewFromFloat(p.Financial.DownPayment)
	analysis.LoanAmount = decimal.NewFromFloat(p.Financial.LoanAmount)
	analysis.InterestRate = decimal.NewFromFloat(p.Financial.InterestRatePercent())
	analysis.LoanTermYears = p.Financial.LoanTermYears.Years()
	analysis.TotalInvestment = analysis.DownPayment // Can add closing costs later

	// Calculate key metrics
	if analysis.PurchasePrice.GreaterThan(decimal.Zero) {
		analysis.CapRate = analysis.AnnualNOI.Div(analysis.PurchasePrice).Mul(decimal.NewFromInt(100))
	}

	if analysis.DownPayment.GreaterThan(decimal.Zero) {
		analysis.CashOnCash = analysis.AnnualCashFlow.Div(analysis.DownPayment).Mul(decimal.NewFromInt(100))
	}

	if analysis.AnnualGrossIncome.GreaterThan(decimal.Zero) {
		analysis.GRM = analysis.PurchasePrice.Div(analysis.AnnualGrossIncome)

		totalDebtService := analysis.AnnualExpenses.Add(analysis.AnnualMortgage)
		analysis.BreakEvenRatio = totalDebtService.Div(analysis.AnnualGrossIncome).Mul(decimal.NewFromInt(100))
	}

	if analysis.AnnualMortgage.GreaterThan(decimal.Zero) {
		analysis.DSCR = analysis.AnnualNOI.Div(analysis.AnnualMortgage)
	}

	return analysis
}

// IsCashFlowPositive returns true if the property generates positive cash flow
func (a *CashFlowAnalysis) IsCashFlowPositive() bool {
	return a.MonthlyCashFlow.GreaterThan(decimal.Zero)
}

// CashFlowSummary returns a one-line summary of cash flow status
func (a *CashFlowAnalysis) CashFlowSummary() string {
	if a.IsCashFlowPositive() {
		return fmt.Sprintf("POSITIVE: $%s/month ($%s/year)",
			a.MonthlyCashFlow.Round(2).String(),
			a.AnnualCashFlow.Round(2).String())
	}
	return fmt.Sprintf("NEGATIVE: -$%s/month (-$%s/year)",
		a.MonthlyCashFlow.Abs().Round(2).String(),
		a.AnnualCashFlow.Abs().Round(2).String())
}

// Report generates a formatted CLI report of the cash flow analysis
func (a *CashFlowAnalysis) Report() string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("╔══════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║               CASH FLOW ANALYSIS                             ║\n")
	sb.WriteString("╠══════════════════════════════════════════════════════════════╣\n")

	// Status
	status := "✓ CASH FLOW POSITIVE"
	if !a.IsCashFlowPositive() {
		status = "✗ CASH FLOW NEGATIVE"
	}
	sb.WriteString(fmt.Sprintf("║  Status: %-51s  ║\n", status))
	sb.WriteString("╠══════════════════════════════════════════════════════════════╣\n")

	// Income
	sb.WriteString("║  INCOME                          Monthly        Annual       ║\n")
	sb.WriteString("║  ─────────────────────────────────────────────────────────   ║\n")
	sb.WriteString(fmt.Sprintf("║  Gross Rental Income      %12s  %14s   ║\n",
		"$"+a.MonthlyGrossIncome.Round(2).String(),
		"$"+a.AnnualGrossIncome.Round(2).String()))

	sb.WriteString("╠══════════════════════════════════════════════════════════════╣\n")

	// Expenses
	sb.WriteString("║  EXPENSES                        Monthly        Annual       ║\n")
	sb.WriteString("║  ─────────────────────────────────────────────────────────   ║\n")
	sb.WriteString(fmt.Sprintf("║  Operating Expenses       %12s  %14s   ║\n",
		"$"+a.MonthlyExpenses.Round(2).String(),
		"$"+a.AnnualExpenses.Round(2).String()))
	sb.WriteString(fmt.Sprintf("║  Mortgage Payment         %12s  %14s   ║\n",
		"$"+a.MonthlyMortgage.Round(2).String(),
		"$"+a.AnnualMortgage.Round(2).String()))

	sb.WriteString("╠══════════════════════════════════════════════════════════════╣\n")

	// NOI and Cash Flow
	sb.WriteString("║  CASH FLOW                       Monthly        Annual       ║\n")
	sb.WriteString("║  ─────────────────────────────────────────────────────────   ║\n")
	sb.WriteString(fmt.Sprintf("║  Net Operating Income     %12s  %14s   ║\n",
		"$"+a.MonthlyNOI.Round(2).String(),
		"$"+a.AnnualNOI.Round(2).String()))

	cashFlowSign := ""
	if a.MonthlyCashFlow.LessThan(decimal.Zero) {
		cashFlowSign = "-"
	}
	sb.WriteString(fmt.Sprintf("║  Net Cash Flow            %12s  %14s   ║\n",
		cashFlowSign+"$"+a.MonthlyCashFlow.Abs().Round(2).String(),
		cashFlowSign+"$"+a.AnnualCashFlow.Abs().Round(2).String()))

	sb.WriteString("╠══════════════════════════════════════════════════════════════╣\n")

	// Key Metrics
	sb.WriteString("║  KEY METRICS                                                 ║\n")
	sb.WriteString("║  ─────────────────────────────────────────────────────────   ║\n")
	sb.WriteString(fmt.Sprintf("║  Cap Rate:                %10s%%                        ║\n",
		a.CapRate.Round(2).String()))
	sb.WriteString(fmt.Sprintf("║  Cash-on-Cash Return:     %10s%%                        ║\n",
		a.CashOnCash.Round(2).String()))
	sb.WriteString(fmt.Sprintf("║  Gross Rent Multiplier:   %10s                         ║\n",
		a.GRM.Round(2).String()))
	sb.WriteString(fmt.Sprintf("║  Debt Service Coverage:   %10s                         ║\n",
		a.DSCR.Round(2).String()))
	sb.WriteString(fmt.Sprintf("║  Break-Even Ratio:        %10s%%                        ║\n",
		a.BreakEvenRatio.Round(2).String()))

	sb.WriteString("╠══════════════════════════════════════════════════════════════╣\n")

	// Investment Summary
	sb.WriteString("║  INVESTMENT SUMMARY                                          ║\n")
	sb.WriteString("║  ─────────────────────────────────────────────────────────   ║\n")
	sb.WriteString(fmt.Sprintf("║  Purchase Price:          %12s                       ║\n",
		"$"+a.PurchasePrice.Round(2).String()))
	sb.WriteString(fmt.Sprintf("║  Down Payment:            %12s                       ║\n",
		"$"+a.DownPayment.Round(2).String()))
	sb.WriteString(fmt.Sprintf("║  Loan Amount:             %12s                       ║\n",
		"$"+a.LoanAmount.Round(2).String()))

	sb.WriteString("╚══════════════════════════════════════════════════════════════╝\n")

	return sb.String()
}
