package property

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// ProjectionConfig configures multi-year cash flow projections
type ProjectionConfig struct {
	Years             int     // Number of years to project
	RentGrowthRate    float64 // Annual rent increase (e.g., 0.03 for 3%)
	ExpenseGrowthRate float64 // Annual expense increase (e.g., 0.02 for 2%)
	AppreciationRate  float64 // Annual property appreciation (e.g., 0.03 for 3%)
	VacancyRate       float64 // Expected vacancy rate (e.g., 0.05 for 5%)
}

// DefaultProjectionConfig returns sensible defaults for projections
func DefaultProjectionConfig() ProjectionConfig {
	return ProjectionConfig{
		Years:             10,
		RentGrowthRate:    0.03, // 3% annual rent increase
		ExpenseGrowthRate: 0.02, // 2% annual expense increase
		AppreciationRate:  0.00, // 3% annual appreciation
		VacancyRate:       0.05, // 5% vacancy
	}
}

// YearlyProjection contains projected values for a single year
type YearlyProjection struct {
	Year            int
	GrossIncome     decimal.Decimal
	VacancyLoss     decimal.Decimal
	EffectiveIncome decimal.Decimal
	Expenses        decimal.Decimal
	NOI             decimal.Decimal
	MortgagePayment decimal.Decimal
	CashFlow        decimal.Decimal
	CumulativeCF    decimal.Decimal
	PropertyValue   decimal.Decimal
	LoanBalance     decimal.Decimal
	PrincipalPaid   decimal.Decimal // Total principal paid to date
	Equity          decimal.Decimal // Down payment + principal paid (actual equity from loan paydown)
	EquityAtSale    decimal.Decimal // PropertyValue - LoanBalance (includes unrealized appreciation)
	TotalReturn     decimal.Decimal // Cash flow + equity gain (based on actual equity)
	CashOnCash      decimal.Decimal
}

// ProjectCashFlow generates multi-year cash flow projections
func ProjectCashFlow(p *Property, config ProjectionConfig) []YearlyProjection {
	projections := make([]YearlyProjection, 0, config.Years)

	// Initial values
	var baseMonthlyRent decimal.Decimal
	for _, unit := range p.Units {
		baseMonthlyRent = baseMonthlyRent.Add(decimal.NewFromFloat(unit.Rent))
	}
	baseAnnualRent := baseMonthlyRent.Mul(decimal.NewFromInt(12))

	baseAnnualExpenses := decimal.NewFromFloat(p.Financial.Expenses.TotalYearly())
	propertyValue := decimal.NewFromFloat(p.Financial.PurchasePrice)
	downPayment := decimal.NewFromFloat(p.Financial.DownPayment)

	loan := p.Financial.Loan()
	annualMortgage := loan.MonthlyPayment().Abs().Mul(decimal.NewFromInt(12))

	// Get amortization schedule for loan balance
	schedule, err := loan.AmortizationSchedule()

	// Growth rate multipliers
	rentGrowth := decimal.NewFromFloat(1 + config.RentGrowthRate)
	expenseGrowth := decimal.NewFromFloat(1 + config.ExpenseGrowthRate)
	appreciation := decimal.NewFromFloat(1 + config.AppreciationRate)
	vacancyRate := decimal.NewFromFloat(config.VacancyRate)

	var cumulativeCashFlow decimal.Decimal
	currentRent := baseAnnualRent
	currentExpenses := baseAnnualExpenses
	currentPropertyValue := propertyValue

	for year := 1; year <= config.Years; year++ {
		proj := YearlyProjection{Year: year}

		// Apply growth for years after the first
		if year > 1 {
			currentRent = currentRent.Mul(rentGrowth)
			currentExpenses = currentExpenses.Mul(expenseGrowth)
			currentPropertyValue = currentPropertyValue.Mul(appreciation)
		}

		// Income
		proj.GrossIncome = currentRent
		proj.VacancyLoss = currentRent.Mul(vacancyRate)
		proj.EffectiveIncome = proj.GrossIncome.Sub(proj.VacancyLoss)

		// Expenses and NOI
		proj.Expenses = currentExpenses
		proj.NOI = proj.EffectiveIncome.Sub(proj.Expenses)

		// Mortgage and cash flow
		proj.MortgagePayment = annualMortgage
		proj.CashFlow = proj.NOI.Sub(proj.MortgagePayment)
		cumulativeCashFlow = cumulativeCashFlow.Add(proj.CashFlow)
		proj.CumulativeCF = cumulativeCashFlow

		// Property value and equity
		proj.PropertyValue = currentPropertyValue

		// Calculate remaining loan balance and principal paid
		// Each year has 12 months of payments
		monthIndex := (year * 12) - 1
		initialLoanAmount := decimal.NewFromFloat(p.Financial.LoanAmount)

		if err == nil && monthIndex < len(schedule) {
			// Get total principal paid from schedule
			totalPrincipalPaid := decimal.Zero
			for i := 0; i <= monthIndex; i++ {
				totalPrincipalPaid = totalPrincipalPaid.Add(schedule[i].Principal.Abs())
			}
			proj.PrincipalPaid = totalPrincipalPaid
			proj.LoanBalance = initialLoanAmount.Sub(totalPrincipalPaid)
		} else {
			// Estimate if schedule not available
			yearsRemaining := p.Financial.LoanTermYears.Years() - year
			if yearsRemaining > 0 {
				proj.LoanBalance = initialLoanAmount.Mul(decimal.NewFromFloat(float64(yearsRemaining) / float64(p.Financial.LoanTermYears.Years())))
			} else {
				proj.LoanBalance = decimal.Zero
			}
			proj.PrincipalPaid = initialLoanAmount.Sub(proj.LoanBalance)
		}

		// Equity = Down payment + Principal paid (actual equity from loan paydown only)
		proj.Equity = downPayment.Add(proj.PrincipalPaid)

		// EquityAtSale = Property Value - Loan Balance (includes unrealized appreciation)
		proj.EquityAtSale = proj.PropertyValue.Sub(proj.LoanBalance)

		// Total return = cumulative cash flow + equity gain from principal paydown
		equityGain := proj.PrincipalPaid // Equity gain is just principal paid (down payment is initial investment)
		proj.TotalReturn = proj.CumulativeCF.Add(equityGain)

		// Cash on cash for this year
		if downPayment.GreaterThan(decimal.Zero) {
			proj.CashOnCash = proj.CashFlow.Div(downPayment).Mul(decimal.NewFromInt(100))
		}

		projections = append(projections, proj)
	}

	return projections
}

// ProjectionReport generates a formatted multi-year projection table
func ProjectionReport(projections []YearlyProjection) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("╔══════════════════════════════════════════════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                                    MULTI-YEAR CASH FLOW PROJECTION                                       ║\n")
	sb.WriteString("╠══════════════════════════════════════════════════════════════════════════════════════════════════════════╣\n")
	sb.WriteString("║ Year │  Gross Inc  │  Expenses  │     NOI     │  Mortgage  │  Cash Flow │   Equity    │ Total Return    ║\n")
	sb.WriteString("╠══════════════════════════════════════════════════════════════════════════════════════════════════════════╣\n")

	for _, p := range projections {
		cfSign := ""
		if p.CashFlow.LessThan(decimal.Zero) {
			cfSign = "-"
		}

		sb.WriteString(fmt.Sprintf("║  %2d  │ $%10s │ $%9s │ $%10s │ $%9s │ %s$%9s │ $%10s │ $%14s ║\n",
			p.Year,
			p.GrossIncome.Round(0).String(),
			p.Expenses.Round(0).String(),
			p.NOI.Round(0).String(),
			p.MortgagePayment.Round(0).String(),
			cfSign, p.CashFlow.Abs().Round(0).String(),
			p.Equity.Round(0).String(),
			p.TotalReturn.Round(0).String()))
	}

	sb.WriteString("╚══════════════════════════════════════════════════════════════════════════════════════════════════════════╝\n")

	// Summary
	if len(projections) > 0 {
		lastYear := projections[len(projections)-1]
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("  After %d years:\n", lastYear.Year))
		sb.WriteString(fmt.Sprintf("    • Cumulative Cash Flow:  $%s\n", lastYear.CumulativeCF.Round(0).String()))
		sb.WriteString(fmt.Sprintf("    • Total Equity:          $%s\n", lastYear.Equity.Round(0).String()))
		sb.WriteString(fmt.Sprintf("    • Total Return:          $%s\n", lastYear.TotalReturn.Round(0).String()))
		sb.WriteString(fmt.Sprintf("    • Property Value:        $%s\n", lastYear.PropertyValue.Round(0).String()))
	}

	return sb.String()
}

// CalculateIRR calculates the Internal Rate of Return for the investment
// Uses bisection method for reliable convergence
func CalculateIRR(p *Property, projections []YearlyProjection, holdingYears int) decimal.Decimal {
	if len(projections) == 0 || holdingYears > len(projections) {
		return decimal.Zero
	}

	// Cash flows: initial investment (negative) + annual cash flows + sale proceeds
	downPayment := decimal.NewFromFloat(p.Financial.DownPayment)
	if downPayment.IsZero() {
		return decimal.Zero
	}
	initialInvestment := downPayment.Neg()

	// Get the projection at the sale year
	saleYear := projections[holdingYears-1]
	saleProceeds := saleYear.EquityAtSale // What you'd get if you sold (property value - loan balance)

	// Build cash flow array
	cashFlows := make([]decimal.Decimal, holdingYears+1)
	cashFlows[0] = initialInvestment
	for i := 0; i < holdingYears; i++ {
		cashFlows[i+1] = projections[i].CashFlow
	}
	// Add sale proceeds to final year
	cashFlows[holdingYears] = cashFlows[holdingYears].Add(saleProceeds)

	// Use bisection method - more reliable than Newton-Raphson
	low := decimal.NewFromFloat(-0.99)  // -99% (can't go below -100%)
	high := decimal.NewFromFloat(2.0)   // 200% max
	tolerance := decimal.NewFromFloat(0.0001)

	npvLow := calculateNPV(cashFlows, low)
	npvHigh := calculateNPV(cashFlows, high)

	// Check if IRR exists in range (NPV should change sign)
	if npvLow.Sign() == npvHigh.Sign() {
		// No sign change - IRR may not exist or is outside range
		// Return a rough estimate based on total return
		totalCashFlow := decimal.Zero
		for _, cf := range cashFlows {
			totalCashFlow = totalCashFlow.Add(cf)
		}
		if downPayment.IsZero() {
			return decimal.Zero
		}
		annualReturn := totalCashFlow.Div(downPayment).Div(decimal.NewFromInt(int64(holdingYears)))
		return annualReturn.Mul(decimal.NewFromInt(100))
	}

	// Bisection method
	for i := 0; i < 100; i++ {
		mid := low.Add(high).Div(decimal.NewFromInt(2))
		npvMid := calculateNPV(cashFlows, mid)

		if npvMid.Abs().LessThan(tolerance) || high.Sub(low).Abs().LessThan(tolerance) {
			return mid.Mul(decimal.NewFromInt(100)) // Return as percentage
		}

		if npvMid.Sign() == npvLow.Sign() {
			low = mid
			npvLow = npvMid
		} else {
			high = mid
		}
	}

	// Return midpoint as best estimate
	return low.Add(high).Div(decimal.NewFromInt(2)).Mul(decimal.NewFromInt(100))
}

func calculateNPV(cashFlows []decimal.Decimal, rate decimal.Decimal) decimal.Decimal {
	npv := decimal.Zero
	one := decimal.NewFromInt(1)

	for i, cf := range cashFlows {
		discountFactor := one.Add(rate).Pow(decimal.NewFromInt(int64(i)))
		npv = npv.Add(cf.Div(discountFactor))
	}

	return npv
}

func calculateNPVDerivative(cashFlows []decimal.Decimal, rate decimal.Decimal) decimal.Decimal {
	derivative := decimal.Zero
	one := decimal.NewFromInt(1)

	for i, cf := range cashFlows {
		if i == 0 {
			continue
		}
		t := decimal.NewFromInt(int64(i))
		discountFactor := one.Add(rate).Pow(t.Add(one))
		derivative = derivative.Sub(t.Mul(cf).Div(discountFactor))
	}

	return derivative
}
