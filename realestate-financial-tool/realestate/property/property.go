package property

import (
	"fmt"

	"realestate-financial-tool/realestate/financing"
)

type Property struct {
	PropertyName  string
	Address       string
	City          string
	State         string
	ZipCode       string
	County        string
	YearBuilt     int
	NumberOfUnits int
	BuildingSF    float64
	LotSF         float64
	Units         Units
	Financial     *Financial
}

// New creates a new property builder
func New(name string) *Property {
	return &Property{
		PropertyName: name,
		Financial:    &Financial{},
	}
}

// At sets the property address
func (p *Property) At(address, city, state, zipCode string) *Property {
	p.Address = address
	p.City = city
	p.State = state
	p.ZipCode = zipCode
	return p
}

// Built sets the year built and building details
func (p *Property) Built(year int, buildingSF, lotSF float64) *Property {
	p.YearBuilt = year
	p.BuildingSF = buildingSF
	p.LotSF = lotSF
	return p
}

// InCounty sets the county
func (p *Property) InCounty(county string) *Property {
	p.County = county
	return p
}

// AddUnit adds a rental unit with beds, baths, sqft, and monthly rent
func (p *Property) AddUnit(beds, baths int, sqft float64, rent float64) *Property {
	p.Units.AddUnit(beds, baths, sqft)
	unitName := fmt.Sprintf("unit%d", len(p.Units))
	p.Units.GetUnit(unitName).SetRent(rent)
	p.NumberOfUnits = len(p.Units)
	return p
}

// Purchase sets purchase price and asking price
func (p *Property) Purchase(purchasePrice float64, askingPrice ...float64) *Property {
	p.Financial.PurchasePrice = purchasePrice
	if len(askingPrice) > 0 {
		p.Financial.AskingPrice = askingPrice[0]
	} else {
		p.Financial.AskingPrice = purchasePrice
	}
	return p
}

// Loan sets loan terms: interest rate (basis points, e.g., 600 = 6%) and term
func (p *Property) Loan(interestRateBasisPoints float64, term financing.LoanTerm) *Property {
	p.Financial.InterestRate = interestRateBasisPoints
	p.Financial.LoanTermYears = term
	return p
}

// DownPayment sets a specific down payment amount
func (p *Property) WithDownPayment(amount float64) *Property {
	p.Financial.DownPayment = amount
	return p
}

// DownPaymentPercent sets down payment as a percentage (e.g., 25 for 25%)
func (p *Property) WithDownPaymentPercent(percent float64) *Property {
	p.Financial.DownPayment = p.Financial.PurchasePrice * (percent / 100)
	return p
}

// Expenses sets monthly operating expenses
func (p *Property) Expenses(taxes, insurance, utilities, maintenance float64) *Property {
	p.Financial.Expenses.Taxes = taxes
	p.Financial.Expenses.Insurance = insurance
	p.Financial.Expenses.Utilities = utilities
	p.Financial.Expenses.RepairsMaintenance = maintenance
	return p
}

// Vacancy sets the vacancy rate (e.g., 0.05 for 5%)
func (p *Property) Vacancy(rate float64) *Property {
	p.Financial.Expenses.VacancyRate = rate
	return p
}

// AnalysisResult wraps the analysis data with convenience methods
type AnalysisResult struct {
	*FullAnalysis
	property *Property
}

// Print outputs the full analysis report to stdout
func (r *AnalysisResult) Print() *AnalysisResult {
	output := NewOutput(r.FullAnalysis)
	output.Print()
	return r
}

// PrintLoanSummary prints the loan summary to stdout
func (r *AnalysisResult) PrintLoanSummary() *AnalysisResult {
	if summary, err := r.property.LoanSummary(); err == nil {
		fmt.Println("\n" + summary)
	}
	return r
}

// GenerateCharts creates the amortization and loan summary HTML charts
func (r *AnalysisResult) GenerateCharts() *AnalysisResult {
	if _, err := r.property.PlotLoan(); err == nil {
		fmt.Println("Amortization chart saved to: plot.html")
	}
	if _, err := r.property.PlotLoanSummary(); err == nil {
		fmt.Println("Loan summary chart saved to: loan_summary.html")
	}
	return r
}

// ToJSON returns the analysis as JSON string
func (r *AnalysisResult) ToJSON() string {
	output := NewOutput(r.FullAnalysis)
	return output.ToJSON()
}

// ToCSV returns the analysis as CSV string
func (r *AnalysisResult) ToCSV() string {
	output := NewOutput(r.FullAnalysis)
	return output.ToCSV()
}

// ToFile writes the analysis to a file in the specified format
func (r *AnalysisResult) ToFile(filename string, format OutputFormat) error {
	output := NewOutput(r.FullAnalysis)
	return output.ToFile(filename, format)
}

// RunFullAnalysis performs complete analysis and returns a result object for further operations
func (p *Property) RunFullAnalysis() *AnalysisResult {
	// Normalize financial defaults
	p.Financial.Normalize()

	// Analyze
	analysis := Analyze(p)

	return &AnalysisResult{
		FullAnalysis: analysis,
		property:     p,
	}
}

func NewProperty(
	propertyName,
	address,
	city,
	state,
	zipCode,
	county string,
	yearBuilt,
	numberOfUnits int,
	buildingSF,
	lotSF float64) Property {
	return Property{
		PropertyName:  propertyName,
		Address:       address,
		City:          city,
		State:         state,
		ZipCode:       zipCode,
		County:        county,
		YearBuilt:     yearBuilt,
		NumberOfUnits: numberOfUnits,
		BuildingSF:    buildingSF,
		LotSF:         lotSF,
	}
}
func (p *Property) SetFinancials(
	askingPrice,
	purchasePrice,
	downPayment,
	loanAmount float64,
	interestRate,
	loanTermYears financing.LoanTerm) {
	p.Financial = &Financial{
		AskingPrice:   askingPrice,
		PurchasePrice: purchasePrice,
		DownPayment:   downPayment,
		LoanAmount:    loanAmount,
		InterestRate:  float64(interestRate), // Convert basis points to decimal
		LoanTermYears: loanTermYears,
	}
}

// Analyze performs complete analysis on the property
func (p *Property) Analyze() *FullAnalysis {
	return Analyze(p)
}

// AnalyzeWithConfig performs complete analysis with custom projection config
func (p *Property) AnalyzeWithConfig(config ProjectionConfig) *FullAnalysis {
	return AnalyzeWithConfig(p, config)
}

// Print analyzes the property and prints the full report to stdout
func (p *Property) Print() {
	analysis := p.Analyze()
	output := NewOutput(analysis)
	output.Print()
}

// PrintSummary prints a quick one-line cash flow summary
func (p *Property) PrintSummary() {
	analysis := p.Analyze()
	println(analysis.Summary())
}

// LoanSummary returns a formatted string with all loan totals
func (p *Property) LoanSummary() (string, error) {
	return p.Financial.LoanSummary()
}

// PlotLoan generates an amortization chart and saves to plot.html
func (p *Property) PlotLoan() (string, error) {
	return p.Financial.PlotLoan()
}

// PlotLoanSummary generates a summary chart and saves to loan_summary.html
func (p *Property) PlotLoanSummary() (string, error) {
	return p.Financial.PlotLoanSummary()
}
