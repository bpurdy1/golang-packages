package financing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/shopspring/decimal"

	"realestate-financial-tool/internal/gofinancial"
	"realestate-financial-tool/internal/gofinancial/enums/frequency"
	"realestate-financial-tool/internal/gofinancial/enums/interesttype"
	"realestate-financial-tool/internal/gofinancial/enums/paymentperiod"
	"realestate-financial-tool/internal/ptr"
)

type LoanTerm int

const (
	Term30Years LoanTerm = iota
	Term20Years
	Term15Years
	Term10Years
)

func (lt LoanTerm) String() string {
	return [...]string{"30 Years", "20 Years", "15 Years", "10 Years"}[lt]
}

func (lt LoanTerm) Years() int {
	switch lt {
	case Term30Years:
		return 30
	case Term20Years:
		return 20
	case Term15Years:
		return 15
	case Term10Years:
		return 10
	default:
		panic(fmt.Sprintf("Unknown loan term: %d", lt))
	}
}

type Loan struct {
	HomePrice   decimal.Decimal `json:"home_price" env:"LOAN_HOME_PRICE" envDefault:"300_000"`
	DownPayment decimal.Decimal `json:"down_payment" env:"LOAN_DOWN_PAYMENT" envDefault:"60_000"`
	// InterestRate           decimal.Decimal
	InterestRate           InterestRate    `json:"interest_rate" env:"INTEREST_RATE" envDefault:"500"`       // in basis points
	StartDate              time.Time       `json:"start_date" env:"LOAN_START_DATE" envDefault:"2024-01-01"` // inclusive
	TermYears              int             `json:"term_years" env:"LOAN_TERM_YEARS" envDefault:"30"`         // NEW: loan term in years
	EnableRounding         bool            `json:"enable_rounding" env:"LOAN_ENABLE_ROUNDING" envDefault:"false"`
	RoundingPlaces         int             `json:"rounding_places" env:"LOAN_ROUNDING_PLACES" envDefault:"2"`                   // 0 for nearest int
	RoundingErrorTolerance decimal.Decimal `json:"rounding_error_tolerance" env:"LOAN_ROUNDING_ERROR_TOLERANCE" envDefault:"0"` // 0 for no error
	EndDate                time.Time       `json:"end_date" env:"LOAN_END_DATE" envDefault:"2054-01-01"`                        // inclusive
}

func NewLoan(
	homePrice int64,
	downPayment int64,
	// point based interest rate in basis points
	interestRate float64,
	years LoanTerm,
	roundingErrorTolerance decimal.Decimal) *Loan {

	if years.Years() <= 0 {
		panic(fmt.Sprintf("Invalid loan term: %d years. Must be greater than 0.", years))
	}

	st := time.Now()
	// Start on the 1st of next month to ensure clean monthly boundaries
	startDate := time.Date(st.Year(), st.Month()+1, 1, 0, 0, 0, 0, st.Location())
	// End date is the last day before the anniversary (years later)
	endDate := startDate.AddDate(years.Years(), 0, -1)
	return &Loan{
		HomePrice:              decimal.NewFromInt(homePrice),
		DownPayment:            decimal.NewFromInt(downPayment),
		InterestRate:           NewInterestRate(float64(interestRate)), // Convert basis points to decimal
		StartDate:              startDate,
		EndDate:                endDate,
		TermYears:              years.Years(),
		EnableRounding:         true,
		RoundingPlaces:         0,
		RoundingErrorTolerance: roundingErrorTolerance,
	}
}

func (l *Loan) Report() string {
	var sb strings.Builder

	totalPayment, _ := l.GetTotalPayment()
	totalInterest, _ := l.GetTotalInterest()
	totalPrincipal, _ := l.GetTotalPrincipal()

	sb.WriteString("Loan Report:\n")
	sb.WriteString("  Home Price: $" + l.HomePrice.StringFixed(2) + "\n")
	sb.WriteString("  Down Payment: $" + l.DownPayment.StringFixed(2) + "\n")
	sb.WriteString("  Interest Rate: " + l.InterestRate.String() + "\n")
	sb.WriteString("  Start Date: " + l.StartDate.Format("2006-01-02") + "\n")
	sb.WriteString("  End Date: " + l.EndDate.Format("2006-01-02") + "\n")
	sb.WriteString(fmt.Sprintf("  Term: %d years\n", l.TermYears))
	sb.WriteString("  Monthly Payment: $" + l.MonthlyPayment().StringFixed(2) + "\n")
	sb.WriteString("  Total Payment: $" + totalPayment.StringFixed(2) + "\n")
	sb.WriteString("  Total Interest: $" + totalInterest.StringFixed(2) + "\n")
	sb.WriteString("  Total Principal: $" + totalPrincipal.StringFixed(2) + "\n")

	return sb.String()
}
func (l *Loan) MonthlyPayment() decimal.Decimal {
	// P = principal, r = monthly rate, n = total payments
	P := l.HomePrice.Sub(l.DownPayment)
	years := l.TermYears
	r := l.InterestRate.Decimal().Div(decimal.NewFromInt(12)) // monthly rate as decimal
	n := int64(years * 12)

	one := decimal.NewFromInt(1)
	if r.IsZero() || n == 0 {
		return decimal.Zero
	}
	num := r.Mul(P).Mul((one.Add(r)).Pow(decimal.NewFromInt(n)))
	den := (one.Add(r)).Pow(decimal.NewFromInt(n)).Sub(one)
	return num.Div(den)
}

func (l *Loan) LoanAmount() decimal.Decimal {
	return l.HomePrice.Sub(l.DownPayment)
}

func (l *Loan) FinancialConfig() gofinancial.Config {
	return gofinancial.Config{
		StartDate:              l.StartDate,
		EndDate:                l.EndDate,
		Frequency:              frequency.MONTHLY,
		AmountBorrowed:         l.LoanAmount(),
		InterestType:           interesttype.REDUCING,
		Interest:               l.InterestRate.Points(),
		PaymentPeriod:          paymentperiod.ENDING,
		EnableRounding:         l.EnableRounding,
		RoundingErrorTolerance: l.RoundingErrorTolerance,
	}
}
func (l *Loan) AmortizationSchedule() ([]gofinancial.Row, error) {
	config := l.FinancialConfig()
	amortization, err := gofinancial.NewAmortization(&config)
	if err != nil {
		return nil, err
	}
	return amortization.GenerateTable()
}
func (l *Loan) PrintAmortizationSchedule() error {
	rows, err := l.AmortizationSchedule()
	if err != nil {
		return err
	}
	gofinancial.PrintRows(rows)
	return nil
}

//	func (l *Loan) PlotAmortizationSchedule(filename string) error {
//		rows, err := l.AmortizationSchedule()
//		if err != nil {
//			return err
//		}
//		return financial.PlotRows(rows, filename)
//	}
func (l *Loan) GetTotalPayment() (decimal.Decimal, error) {
	rows, err := l.AmortizationSchedule()
	if err != nil {
		return decimal.Zero, err
	}
	b, _ := json.MarshalIndent(rows, "", "  ") // For debugging purposes, you can remove this line if not needed
	os.WriteFile("amortization_schedule.json", b, 0644)

	var totalPayment decimal.Decimal
	for _, row := range rows {
		totalPayment = totalPayment.Add(row.Payment)
	}
	return totalPayment, nil
}
func (l *Loan) GetTotalInterest() (decimal.Decimal, error) {
	rows, err := l.AmortizationSchedule()
	if err != nil {
		return decimal.Zero, err
	}
	var totalInterest decimal.Decimal
	for _, row := range rows {
		totalInterest = totalInterest.Add(row.Interest)
	}
	return totalInterest, nil
}
func (l *Loan) GetTotalPrincipal() (decimal.Decimal, error) {
	rows, err := l.AmortizationSchedule()
	if err != nil {
		return decimal.Zero, err
	}

	var totalPrincipal decimal.Decimal
	for _, row := range rows {
		totalPrincipal = totalPrincipal.Add(row.Principal)
	}
	return totalPrincipal, nil
}

func (l *Loan) Plot() (string, error) {
	rows, err := l.AmortizationSchedule()
	if err != nil {
		return "", err
	}
	// create a new bar instance
	barChart := charts.NewBar()
	// set some global options like Title/Legend/ToolTip or anything else
	barChart.SetGlobalOptions(
		charts.WithTitleOpts(
			opts.Title{
				Title: "Loan repayment schedule",
				Subtitle: l.HomePrice.String() +
					" " + l.InterestRate.String() +
					" " + l.StartDate.Format("2006-01-02") +
					" - " + l.EndDate.Format("2006-01-02"),
			},
		),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1200px",
			Height: "600px",
		}),

		charts.WithToolboxOpts(
			opts.Toolbox{Show: ptr.BoolPtr(true)}), //1.26 new(true)
		charts.WithTooltipOpts(
			opts.Tooltip{
				Show:    ptr.BoolPtr(true),
				Trigger: "axis",
				AxisPointer: &opts.AxisPointer{
					Type: "shadow",
				},
			}),
		charts.WithLegendOpts(opts.Legend{Show: ptr.BoolPtr(true)}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "inside",
			Start: 0,
			End:   50,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 0,
			End:   50,
		}),
	)

	var xAxis []string
	var interestArr []opts.BarData
	var principalArr []opts.BarData
	var paymentArr []opts.BarData

	for _, row := range rows {
		xAxis = append(xAxis, row.EndDate.Format("2006-01-02"))
		interestArr = append(interestArr,
			opts.BarData{Value: row.Interest.Abs().InexactFloat64()})
		principalArr = append(principalArr,
			opts.BarData{Value: row.Principal.Abs().InexactFloat64()})
		paymentArr = append(paymentArr,
			opts.BarData{Value: row.Payment.Abs().InexactFloat64()})
	}
	// Put data into instance
	barChart.SetXAxis(xAxis).
		AddSeries("interest", interestArr).
		AddSeries("principal", principalArr).
		AddSeries("payment", paymentArr).
		SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{
				Show:     ptr.BoolPtr(false), // Show values on bars
				Position: "top",              // Position values on top of bars
			}),
		)

	// Where the magic happens
	var buf bytes.Buffer
	if err := barChart.Render(&buf); err != nil {
		return "", err
	}
	os.WriteFile("plot.html", buf.Bytes(), 0644)

	return buf.String(), nil
}

// PlotSummary generates an amortization chart showing cumulative principal paid,
// interest paid, and remaining loan balance over time (monthly intervals)
func (l *Loan) PlotSummary() (string, error) {
	rows, err := l.AmortizationSchedule()
	if err != nil {
		return "", err
	}

	totalInterest, _ := l.GetTotalInterest()
	totalPayment, _ := l.GetTotalPayment()
	loanAmount := l.LoanAmount()

	// Create a bar chart for amortization over time
	barChart := charts.NewBar()
	barChart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Amortization for Mortgage Loan",
			Subtitle: fmt.Sprintf("Loan: $%s | Total Interest: $%s | Total Cost: $%s | Payoff: %s",
				loanAmount.Round(0).String(),
				totalInterest.Abs().Round(0).String(),
				totalPayment.Abs().Round(0).String(),
				l.EndDate.Format("Jan 2006")),
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1400px",
			Height: "600px",
		}),
		charts.WithToolboxOpts(opts.Toolbox{Show: ptr.BoolPtr(true)}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:    ptr.BoolPtr(true),
			Trigger: "axis",
			AxisPointer: &opts.AxisPointer{
				Type: "shadow",
			},
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: ptr.BoolPtr(true),
			Top:  "bottom",
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "inside",
			Start: 0,
			End:   100,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 0,
			End:   100,
		}),
	)

	// Calculate monthly cumulative totals
	var xAxis []string
	var principalPaidArr []opts.BarData
	var interestPaidArr []opts.BarData
	var loanBalanceArr []opts.BarData

	cumulativePrincipal := decimal.Zero
	cumulativeInterest := decimal.Zero
	initialLoan := loanAmount

	// Add data point for every month
	for _, row := range rows {
		cumulativePrincipal = cumulativePrincipal.Add(row.Principal.Abs())
		cumulativeInterest = cumulativeInterest.Add(row.Interest.Abs())
		remainingBalance := initialLoan.Sub(cumulativePrincipal)

		// Use month/year format for x-axis
		xAxis = append(xAxis, row.EndDate.Format("Jan 2006"))
		principalPaidArr = append(principalPaidArr,
			opts.BarData{Value: cumulativePrincipal.Round(0).InexactFloat64()})
		interestPaidArr = append(interestPaidArr,
			opts.BarData{Value: cumulativeInterest.Round(0).InexactFloat64()})
		loanBalanceArr = append(loanBalanceArr,
			opts.BarData{Value: remainingBalance.Round(0).InexactFloat64()})
	}

	// Put data into chart
	barChart.SetXAxis(xAxis).
		AddSeries("Principal Paid", principalPaidArr).
		AddSeries("Interest Paid", interestPaidArr).
		AddSeries("Loan Balance", loanBalanceArr).
		SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{
				Show: ptr.BoolPtr(false),
			}),
		)

	var buf bytes.Buffer
	if err := barChart.Render(&buf); err != nil {
		return "", err
	}
	os.WriteFile("loan_summary.html", buf.Bytes(), 0644)

	return buf.String(), nil
}

// LoanSummary returns a formatted string with all loan totals
func (l *Loan) LoanSummary() (string, error) {
	totalPayment, err := l.GetTotalPayment()
	if err != nil {
		return "", err
	}
	totalInterest, err := l.GetTotalInterest()
	if err != nil {
		return "", err
	}

	loanAmount := l.LoanAmount()
	monthlyPayment := l.MonthlyPayment()

	var sb strings.Builder
	sb.WriteString("LOAN SUMMARY\n")
	sb.WriteString("============\n")
	sb.WriteString(fmt.Sprintf("Loan Amount:        $%s\n", loanAmount.Round(0).String()))
	sb.WriteString(fmt.Sprintf("Interest Rate:      %s\n", l.InterestRate.String()))
	sb.WriteString(fmt.Sprintf("Loan Term:          %d years\n", l.TermYears))
	sb.WriteString(fmt.Sprintf("Monthly Payment:    $%s\n", monthlyPayment.Abs().Round(2).String()))
	sb.WriteString(fmt.Sprintf("Total Interest:     $%s\n", totalInterest.Abs().Round(0).String()))
	sb.WriteString(fmt.Sprintf("Total Cost of Loan: $%s\n", totalPayment.Abs().Round(0).String()))
	sb.WriteString(fmt.Sprintf("Payoff Date:        %s\n", l.EndDate.Format("Jan 2006")))

	return sb.String(), nil
}
