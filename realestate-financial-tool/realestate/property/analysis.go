package property

import (
	"github.com/shopspring/decimal"
)

// FullAnalysis contains all analysis data for a property investment
type FullAnalysis struct {
	// Property Info
	Property PropertyInfo

	// Units
	Units []UnitInfo

	// Cash Flow Analysis
	CashFlow CashFlowAnalysis

	// Break-even Analysis
	BreakEven BreakEvenAnalysis

	// Scenario Comparisons
	Scenarios []ScenarioResult

	// Multi-year Projections
	Projections []YearlyProjection

	// IRR Calculations
	IRR IRRAnalysis
}

// PropertyInfo contains basic property information
type PropertyInfo struct {
	Name          string
	Address       string
	City          string
	State         string
	ZipCode       string
	County        string
	YearBuilt     int
	NumberOfUnits int
	BuildingSF    float64
	LotSF         float64
}

// UnitInfo contains information about a rental unit
type UnitInfo struct {
	Name      string
	Bedrooms  int
	Bathrooms int
	Size      float64
	Rent      decimal.Decimal
}

// BreakEvenAnalysis contains break-even calculations
type BreakEvenAnalysis struct {
	BreakEvenRentPerUnit decimal.Decimal
	BreakEvenDownPayment decimal.Decimal
	CanBreakEven         bool
}

// IRRAnalysis contains Internal Rate of Return calculations
type IRRAnalysis struct {
	FiveYearIRR  decimal.Decimal
	TenYearIRR   decimal.Decimal
	HoldingYears int
}

// Analyze performs complete analysis on a property and returns all data
func Analyze(p *Property) *FullAnalysis {
	return AnalyzeWithConfig(p, DefaultProjectionConfig())
}

// AnalyzeWithConfig performs complete analysis with custom projection config
func AnalyzeWithConfig(p *Property, projConfig ProjectionConfig) *FullAnalysis {
	analysis := &FullAnalysis{}

	// Property Info
	analysis.Property = PropertyInfo{
		Name:          p.PropertyName,
		Address:       p.Address,
		City:          p.City,
		State:         p.State,
		ZipCode:       p.ZipCode,
		County:        p.County,
		YearBuilt:     p.YearBuilt,
		NumberOfUnits: p.NumberOfUnits,
		BuildingSF:    p.BuildingSF,
		LotSF:         p.LotSF,
	}

	// Units
	for _, unit := range p.Units {
		analysis.Units = append(analysis.Units, UnitInfo{
			Name:      unit.Name,
			Bedrooms:  unit.Bedrooms,
			Bathrooms: unit.Bathrooms,
			Size:      unit.Size,
			Rent:      decimal.NewFromFloat(unit.Rent),
		})
	}

	// Cash Flow Analysis
	cashFlow := AnalyzeCashFlow(p)
	analysis.CashFlow = *cashFlow

	// Break-even Analysis
	breakEvenRent := FindBreakEvenRent(p)
	breakEvenDown := FindBreakEvenDownPayment(p)
	analysis.BreakEven = BreakEvenAnalysis{
		BreakEvenRentPerUnit: breakEvenRent,
		BreakEvenDownPayment: breakEvenDown,
		CanBreakEven:         breakEvenDown.GreaterThan(decimal.Zero),
	}

	// Scenario Comparisons
	scenarios := GenerateDownPaymentScenarios(p, []float64{10, 15, 20, 25, 30})
	analysis.Scenarios = CompareScenarios(p, scenarios)

	// Multi-year Projections
	analysis.Projections = ProjectCashFlow(p, projConfig)

	// IRR Calculations
	if len(analysis.Projections) >= 5 {
		analysis.IRR = IRRAnalysis{
			FiveYearIRR:  CalculateIRR(p, analysis.Projections, 5),
			TenYearIRR:   CalculateIRR(p, analysis.Projections, min(10, len(analysis.Projections))),
			HoldingYears: len(analysis.Projections),
		}
	}

	return analysis
}

// TotalMonthlyRent calculates total monthly rent from all units
func (a *FullAnalysis) TotalMonthlyRent() decimal.Decimal {
	total := decimal.Zero
	for _, unit := range a.Units {
		total = total.Add(unit.Rent)
	}
	return total
}

// IsCashFlowPositive returns true if property has positive cash flow
func (a *FullAnalysis) IsCashFlowPositive() bool {
	return a.CashFlow.IsCashFlowPositive()
}

// Summary returns a quick one-line summary
func (a *FullAnalysis) Summary() string {
	return a.CashFlow.CashFlowSummary()
}
