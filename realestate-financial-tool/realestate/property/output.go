package property

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/shopspring/decimal"
)

// OutputFormat specifies the output format type
type OutputFormat int

const (
	FormatCLI OutputFormat = iota
	FormatJSON
	FormatCSV
)

// Output handles formatting and outputting analysis results
type Output struct {
	analysis *FullAnalysis
}

// NewOutput creates an output handler for the given analysis
func NewOutput(analysis *FullAnalysis) *Output {
	return &Output{analysis: analysis}
}

// Print outputs the analysis to stdout
func (o *Output) Print() {
	fmt.Print(o.ToCLI())
}

// ToFile writes the analysis to a file in the specified format
func (o *Output) ToFile(filename string, format OutputFormat) error {
	var content string
	switch format {
	case FormatJSON:
		content = o.ToJSON()
	case FormatCSV:
		content = o.ToCSV()
	default:
		content = o.ToCLI()
	}
	return os.WriteFile(filename, []byte(content), 0644)
}

// ToJSON returns the analysis as JSON
func (o *Output) ToJSON() string {
	data, _ := json.MarshalIndent(o.analysis, "", "  ")
	return string(data)
}

// ToCSV returns the analysis as CSV (key metrics only)
func (o *Output) ToCSV() string {
	var sb strings.Builder
	a := o.analysis

	sb.WriteString("Metric,Monthly,Annual\n")
	sb.WriteString(fmt.Sprintf("Gross Income,%s,%s\n",
		a.CashFlow.MonthlyGrossIncome.Round(2).String(),
		a.CashFlow.AnnualGrossIncome.Round(2).String()))
	sb.WriteString(fmt.Sprintf("Operating Expenses,%s,%s\n",
		a.CashFlow.MonthlyExpenses.Round(2).String(),
		a.CashFlow.AnnualExpenses.Round(2).String()))
	sb.WriteString(fmt.Sprintf("Mortgage Payment,%s,%s\n",
		a.CashFlow.MonthlyMortgage.Round(2).String(),
		a.CashFlow.AnnualMortgage.Round(2).String()))
	sb.WriteString(fmt.Sprintf("Net Operating Income,%s,%s\n",
		a.CashFlow.MonthlyNOI.Round(2).String(),
		a.CashFlow.AnnualNOI.Round(2).String()))
	sb.WriteString(fmt.Sprintf("Net Cash Flow,%s,%s\n",
		a.CashFlow.MonthlyCashFlow.Round(2).String(),
		a.CashFlow.AnnualCashFlow.Round(2).String()))

	sb.WriteString("\nMetric,Value\n")
	sb.WriteString(fmt.Sprintf("Cap Rate,%s%%\n", a.CashFlow.CapRate.Round(2).String()))
	sb.WriteString(fmt.Sprintf("Cash-on-Cash Return,%s%%\n", a.CashFlow.CashOnCash.Round(2).String()))
	sb.WriteString(fmt.Sprintf("GRM,%s\n", a.CashFlow.GRM.Round(2).String()))
	sb.WriteString(fmt.Sprintf("DSCR,%s\n", a.CashFlow.DSCR.Round(2).String()))

	return sb.String()
}

// ToCLI returns the analysis as formatted CLI output
func (o *Output) ToCLI() string {
	var sb strings.Builder
	a := o.analysis

	// Header
	sb.WriteString("\n")
	sb.WriteString("=============================================================================\n")
	sb.WriteString("                      PROPERTY INVESTMENT ANALYSIS\n")
	sb.WriteString("=============================================================================\n")

	// Property Info
	sb.WriteString(fmt.Sprintf("\n  Property: %s\n", a.Property.Name))
	sb.WriteString(fmt.Sprintf("  Address:  %s, %s, %s %s\n", a.Property.Address, a.Property.City, a.Property.State, a.Property.ZipCode))
	sb.WriteString(fmt.Sprintf("  Units:    %d\n", a.Property.NumberOfUnits))

	// Loan Details
	sb.WriteString("\n  LOAN DETAILS:\n")
	loanTable := tablewriter.NewTable(&sb)
	loanTable.Append([]string{"Purchase Price", "$" + a.CashFlow.PurchasePrice.Round(0).String()})
	downPaymentPct := a.CashFlow.DownPayment.Div(a.CashFlow.PurchasePrice).Mul(decimal.NewFromInt(100))
	loanTable.Append([]string{"Down Payment", fmt.Sprintf("$%s (%.0f%%)", a.CashFlow.DownPayment.Round(0).String(), downPaymentPct.InexactFloat64())})
	loanTable.Append([]string{"Loan Amount", "$" + a.CashFlow.LoanAmount.Round(0).String()})
	loanTable.Append([]string{"Interest Rate", a.CashFlow.InterestRate.Round(2).String() + "%"})
	loanTable.Append([]string{"Loan Term", fmt.Sprintf("%d years", a.CashFlow.LoanTermYears)})
	loanTable.Append([]string{"Monthly Mortgage", "$" + a.CashFlow.MonthlyMortgage.Round(2).String()})
	loanTable.Render()

	// Units table
	if len(a.Units) > 0 {
		sb.WriteString("\n  RENTAL UNITS:\n")
		unitsTable := tablewriter.NewTable(&sb)
		unitsTable.Header("Unit", "Bed", "Bath", "Rent")

		var totalRent decimal.Decimal
		for _, unit := range a.Units {
			totalRent = totalRent.Add(unit.Rent)
			unitsTable.Append([]string{
				unit.Name,
				fmt.Sprintf("%d", unit.Bedrooms),
				fmt.Sprintf("%d", unit.Bathrooms),
				"$" + unit.Rent.Round(0).String(),
			})
		}
		unitsTable.Footer("", "", "Total", "$"+totalRent.Round(0).String())
		unitsTable.Render()
	}

	// Cash Flow Analysis
	sb.WriteString(o.formatCashFlowSection())

	// Break-even Analysis
	sb.WriteString(o.formatBreakEvenSection())

	// Scenario Comparison
	sb.WriteString(o.formatScenarioSection())

	// Multi-year Projection
	sb.WriteString(o.formatProjectionSection())

	// IRR
	sb.WriteString(o.formatIRRSection())

	return sb.String()
}

func (o *Output) formatCashFlowSection() string {
	var sb strings.Builder
	a := &o.analysis.CashFlow

	// Status
	status := "CASH FLOW POSITIVE"
	if !a.IsCashFlowPositive() {
		status = "CASH FLOW NEGATIVE"
	}

	sb.WriteString("\n-----------------------------------------------------------------------------\n")
	sb.WriteString(fmt.Sprintf("  CASH FLOW ANALYSIS                                    Status: %s\n", status))
	sb.WriteString("-----------------------------------------------------------------------------\n")

	// Income/Expense table
	table := tablewriter.NewTable(&sb)
	table.Header("", "Monthly", "Annual")

	table.Append([]string{"Gross Rental Income", "$" + a.MonthlyGrossIncome.Round(2).String(), "$" + a.AnnualGrossIncome.Round(2).String()})
	table.Append([]string{"Operating Expenses", "$" + a.MonthlyExpenses.Round(2).String(), "$" + a.AnnualExpenses.Round(2).String()})
	table.Append([]string{"Mortgage Payment", "$" + a.MonthlyMortgage.Round(2).String(), "$" + a.AnnualMortgage.Round(2).String()})
	table.Append([]string{"", "", ""})
	table.Append([]string{"Net Operating Income", "$" + a.MonthlyNOI.Round(2).String(), "$" + a.AnnualNOI.Round(2).String()})
	table.Append([]string{"NET CASH FLOW", formatMoney(a.MonthlyCashFlow), formatMoney(a.AnnualCashFlow)})
	table.Render()

	// Key Metrics
	sb.WriteString("\n  KEY METRICS:\n")
	metricsTable := tablewriter.NewTable(&sb)
	metricsTable.Append([]string{"Cap Rate", a.CapRate.Round(2).String() + "%"})
	metricsTable.Append([]string{"Cash-on-Cash Return", a.CashOnCash.Round(2).String() + "%"})
	metricsTable.Append([]string{"Gross Rent Multiplier", a.GRM.Round(2).String()})
	metricsTable.Append([]string{"Debt Service Coverage", a.DSCR.Round(2).String()})
	metricsTable.Append([]string{"Break-Even Ratio", a.BreakEvenRatio.Round(2).String() + "%"})
	metricsTable.Render()

	// Investment Summary
	sb.WriteString("\n  INVESTMENT SUMMARY:\n")
	investTable := tablewriter.NewTable(&sb)
	investTable.Append([]string{"Purchase Price", "$" + a.PurchasePrice.Round(0).String()})
	investTable.Append([]string{"Down Payment", "$" + a.DownPayment.Round(0).String()})
	investTable.Append([]string{"Loan Amount", "$" + a.LoanAmount.Round(0).String()})
	investTable.Render()

	return sb.String()
}

func (o *Output) formatBreakEvenSection() string {
	var sb strings.Builder
	b := &o.analysis.BreakEven

	sb.WriteString("\n-----------------------------------------------------------------------------\n")
	sb.WriteString("  BREAK-EVEN ANALYSIS\n")
	sb.WriteString("-----------------------------------------------------------------------------\n")

	table := tablewriter.NewTable(&sb)
	table.Append([]string{"Break-even rent per unit", "$" + b.BreakEvenRentPerUnit.Round(2).String() + "/month"})

	if b.CanBreakEven {
		table.Append([]string{"Break-even down payment", "$" + b.BreakEvenDownPayment.Round(0).String()})
	} else {
		table.Append([]string{"Break-even down payment", "N/A (expenses exceed income)"})
	}
	table.Render()

	return sb.String()
}

func (o *Output) formatScenarioSection() string {
	var sb strings.Builder
	results := o.analysis.Scenarios

	sb.WriteString("\n-----------------------------------------------------------------------------\n")
	sb.WriteString("  SCENARIO COMPARISON\n")
	sb.WriteString("-----------------------------------------------------------------------------\n")

	table := tablewriter.NewTable(&sb)
	table.Header("Scenario", "Monthly CF", "Annual CF", "Cash/Cash", "Status")

	for _, r := range results {
		status := "Positive"
		if !r.IsPositive {
			status = "Negative"
		}

		table.Append([]string{
			r.Scenario.Name,
			formatMoney(r.MonthlyCashFlow),
			formatMoney(r.AnnualCashFlow),
			r.CashOnCash.Round(1).String() + "%",
			status,
		})
	}
	table.Render()

	return sb.String()
}

func (o *Output) formatProjectionSection() string {
	var sb strings.Builder
	projections := o.analysis.Projections

	sb.WriteString("\n-----------------------------------------------------------------------------\n")
	sb.WriteString("  MULTI-YEAR CASH FLOW PROJECTION\n")
	sb.WriteString("-----------------------------------------------------------------------------\n")

	table := tablewriter.NewTable(&sb)
	table.Header("Year", "Gross Inc", "Expenses", "NOI", "Mortgage", "Cash Flow", "Principal Paid", "Loan Balance")

	for _, p := range projections {
		table.Append([]string{
			fmt.Sprintf("%d", p.Year),
			"$" + p.GrossIncome.Round(0).String(),
			"$" + p.Expenses.Round(0).String(),
			"$" + p.NOI.Round(0).String(),
			"$" + p.MortgagePayment.Round(0).String(),
			formatMoney(p.CashFlow),
			"$" + p.PrincipalPaid.Round(0).String(),
			"$" + p.LoanBalance.Round(0).String(),
		})
	}
	table.Render()

	if len(projections) > 0 {
		lastYear := projections[len(projections)-1]
		downPayment := o.analysis.CashFlow.DownPayment
		sb.WriteString(fmt.Sprintf("\n  After %d years:\n", lastYear.Year))
		sb.WriteString(fmt.Sprintf("    Cumulative Cash Flow:    $%s\n", lastYear.CumulativeCF.Round(0).String()))
		sb.WriteString(fmt.Sprintf("    Principal Paid:          $%s\n", lastYear.PrincipalPaid.Round(0).String()))
		sb.WriteString(fmt.Sprintf("    Remaining Loan Balance:  $%s\n", lastYear.LoanBalance.Round(0).String()))
		sb.WriteString(fmt.Sprintf("    Equity (loan paydown):   $%s (down payment + principal paid)\n", lastYear.Equity.Round(0).String()))
		sb.WriteString(fmt.Sprintf("    Equity if sold:          $%s (property value - loan balance)\n", lastYear.EquityAtSale.Round(0).String()))
		sb.WriteString(fmt.Sprintf("    Property Value:          $%s\n", lastYear.PropertyValue.Round(0).String()))
		sb.WriteString(fmt.Sprintf("    Total Return:            $%s (cash flow + principal paid)\n", lastYear.TotalReturn.Round(0).String()))
		appreciationGain := lastYear.EquityAtSale.Sub(downPayment).Sub(lastYear.PrincipalPaid)
		sb.WriteString(fmt.Sprintf("    Appreciation (if sold):  $%s\n", appreciationGain.Round(0).String()))
	}

	return sb.String()
}

func (o *Output) formatIRRSection() string {
	var sb strings.Builder
	irr := &o.analysis.IRR

	if irr.HoldingYears >= 5 {
		sb.WriteString("\n-----------------------------------------------------------------------------\n")
		sb.WriteString("  INTERNAL RATE OF RETURN (IRR)\n")
		sb.WriteString("-----------------------------------------------------------------------------\n")

		table := tablewriter.NewTable(&sb)
		table.Append([]string{"5-Year Hold", irr.FiveYearIRR.Round(2).String() + "%"})
		table.Append([]string{"10-Year Hold", irr.TenYearIRR.Round(2).String() + "%"})
		table.Render()
	}

	return sb.String()
}

// formatMoney formats a decimal as money, handling negatives
func formatMoney(d decimal.Decimal) string {
	if d.LessThan(decimal.Zero) {
		return "-$" + d.Abs().Round(0).String()
	}
	return "$" + d.Round(0).String()
}
