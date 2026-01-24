package property

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"realestate-financial-tool/realestate/financing"
)

// Scenario represents a what-if scenario for property analysis
type Scenario struct {
	Name           string
	DownPayment    float64            // Override down payment
	InterestRate   float64            // Override interest rate (basis points)
	LoanTerm       financing.LoanTerm // Override loan term
	RentMultiplier float64            // Multiply all rents by this factor (1.0 = no change)
}

// ScenarioResult contains the analysis results for a scenario
type ScenarioResult struct {
	Scenario        Scenario
	MonthlyCashFlow decimal.Decimal
	AnnualCashFlow  decimal.Decimal
	CashOnCash      decimal.Decimal
	CapRate         decimal.Decimal
	IsPositive      bool
}

// DefaultScenario creates a scenario with current property values
func DefaultScenario(p *Property) Scenario {
	return Scenario{
		Name:           "Current",
		DownPayment:    p.Financial.DownPayment,
		InterestRate:   p.Financial.InterestRate,
		LoanTerm:       p.Financial.LoanTermYears,
		RentMultiplier: 1.0,
	}
}

// CompareScenarios analyzes multiple scenarios for a property
func CompareScenarios(p *Property, scenarios []Scenario) []ScenarioResult {
	results := make([]ScenarioResult, 0, len(scenarios))

	for _, scenario := range scenarios {
		result := analyzeScenario(p, scenario)
		results = append(results, result)
	}

	return results
}

// analyzeScenario runs cash flow analysis for a specific scenario
func analyzeScenario(p *Property, scenario Scenario) ScenarioResult {
	// Create a copy of financial with scenario overrides
	modifiedFinancial := &Financial{
		AskingPrice:   p.Financial.AskingPrice,
		PurchasePrice: p.Financial.PurchasePrice,
		DownPayment:   scenario.DownPayment,
		LoanAmount:    p.Financial.PurchasePrice - scenario.DownPayment,
		InterestRate:  scenario.InterestRate,
		LoanTermYears: scenario.LoanTerm,
		Expenses:      p.Financial.Expenses,
	}

	// Create modified property
	modifiedProperty := &Property{
		PropertyName:  p.PropertyName,
		Address:       p.Address,
		City:          p.City,
		State:         p.State,
		ZipCode:       p.ZipCode,
		County:        p.County,
		YearBuilt:     p.YearBuilt,
		NumberOfUnits: p.NumberOfUnits,
		BuildingSF:    p.BuildingSF,
		LotSF:         p.LotSF,
		Financial:     modifiedFinancial,
	}

	// Apply rent multiplier to units
	for _, unit := range p.Units {
		modifiedUnit := &Unit{
			Uuid:      unit.Uuid,
			Name:      unit.Name,
			Bedrooms:  unit.Bedrooms,
			Bathrooms: unit.Bathrooms,
			Size:      unit.Size,
			Rent:      unit.Rent * scenario.RentMultiplier,
			Occupied:  unit.Occupied,
		}
		modifiedProperty.Units = append(modifiedProperty.Units, modifiedUnit)
	}

	// Run analysis
	analysis := AnalyzeCashFlow(modifiedProperty)

	return ScenarioResult{
		Scenario:        scenario,
		MonthlyCashFlow: analysis.MonthlyCashFlow,
		AnnualCashFlow:  analysis.AnnualCashFlow,
		CashOnCash:      analysis.CashOnCash,
		CapRate:         analysis.CapRate,
		IsPositive:      analysis.IsCashFlowPositive(),
	}
}

// FindBreakEvenRent calculates the minimum monthly rent per unit to break even
func FindBreakEvenRent(p *Property) decimal.Decimal {
	if len(p.Units) == 0 {
		return decimal.Zero
	}

	// Get total monthly costs
	expenses := decimal.NewFromFloat(p.Financial.Expenses.TotalMonthly())
	mortgage := p.Financial.Loan().MonthlyPayment().Abs()
	totalMonthlyCost := expenses.Add(mortgage)

	// Divide by number of units to get per-unit break-even rent
	numUnits := decimal.NewFromInt(int64(len(p.Units)))
	return totalMonthlyCost.Div(numUnits)
}

// FindBreakEvenDownPayment calculates the minimum down payment to break even
// Returns the down payment amount, or -1 if break-even is not possible
func FindBreakEvenDownPayment(p *Property) decimal.Decimal {
	// Calculate monthly income
	var monthlyIncome decimal.Decimal
	for _, unit := range p.Units {
		monthlyIncome = monthlyIncome.Add(decimal.NewFromFloat(unit.Rent))
	}

	// Calculate monthly expenses
	expenses := decimal.NewFromFloat(p.Financial.Expenses.TotalMonthly())

	// Available for mortgage = income - expenses
	availableForMortgage := monthlyIncome.Sub(expenses)

	if availableForMortgage.LessThanOrEqual(decimal.Zero) {
		// Can't break even even with 100% down
		return decimal.NewFromInt(-1)
	}

	// Binary search for break-even down payment
	purchasePrice := decimal.NewFromFloat(p.Financial.PurchasePrice)
	low := decimal.Zero
	high := purchasePrice
	tolerance := decimal.NewFromFloat(100) // $100 tolerance

	for high.Sub(low).GreaterThan(tolerance) {
		mid := low.Add(high).Div(decimal.NewFromInt(2))

		// Create scenario with this down payment
		scenario := Scenario{
			Name:           "test",
			DownPayment:    mid.InexactFloat64(),
			InterestRate:   p.Financial.InterestRate,
			LoanTerm:       p.Financial.LoanTermYears,
			RentMultiplier: 1.0,
		}

		result := analyzeScenario(p, scenario)

		if result.IsPositive {
			// Can afford less down payment
			high = mid
		} else {
			// Need more down payment
			low = mid
		}
	}

	// Return the higher value to ensure positive cash flow
	return high.Round(0)
}

// GenerateDownPaymentScenarios creates scenarios with different down payment percentages
func GenerateDownPaymentScenarios(p *Property, percentages []float64) []Scenario {
	scenarios := make([]Scenario, 0, len(percentages))
	purchasePrice := p.Financial.PurchasePrice

	for _, pct := range percentages {
		downPayment := purchasePrice * (pct / 100)
		scenarios = append(scenarios, Scenario{
			Name:           fmt.Sprintf("%.0f%% Down", pct),
			DownPayment:    downPayment,
			InterestRate:   p.Financial.InterestRate,
			LoanTerm:       p.Financial.LoanTermYears,
			RentMultiplier: 1.0,
		})
	}

	return scenarios
}

// GenerateInterestRateScenarios creates scenarios with different interest rates
func GenerateInterestRateScenarios(p *Property, rates []float64) []Scenario {
	scenarios := make([]Scenario, 0, len(rates))

	for _, rate := range rates {
		scenarios = append(scenarios, Scenario{
			Name:           fmt.Sprintf("%.2f%% Rate", rate),
			DownPayment:    p.Financial.DownPayment,
			InterestRate:   rate * 100, // Convert to basis points
			LoanTerm:       p.Financial.LoanTermYears,
			RentMultiplier: 1.0,
		})
	}

	return scenarios
}

// ScenarioComparisonReport generates a formatted comparison table
func ScenarioComparisonReport(results []ScenarioResult) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("╔════════════════════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                         SCENARIO COMPARISON                                    ║\n")
	sb.WriteString("╠════════════════════════════════════════════════════════════════════════════════╣\n")
	sb.WriteString("║  Scenario          │ Monthly CF  │ Annual CF   │ Cash/Cash │ Status           ║\n")
	sb.WriteString("╠════════════════════════════════════════════════════════════════════════════════╣\n")

	for _, r := range results {
		status := "✓ Positive"
		if !r.IsPositive {
			status = "✗ Negative"
		}

		cfSign := ""
		if r.MonthlyCashFlow.LessThan(decimal.Zero) {
			cfSign = "-"
		}

		sb.WriteString(fmt.Sprintf("║  %-17s │ %s$%-9s │ %s$%-9s │ %8s%% │ %-16s ║\n",
			truncateString(r.Scenario.Name, 17),
			cfSign, r.MonthlyCashFlow.Abs().Round(0).String(),
			cfSign, r.AnnualCashFlow.Abs().Round(0).String(),
			r.CashOnCash.Round(1).String(),
			status))
	}

	sb.WriteString("╚════════════════════════════════════════════════════════════════════════════════╝\n")

	return sb.String()
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-2] + ".."
}
