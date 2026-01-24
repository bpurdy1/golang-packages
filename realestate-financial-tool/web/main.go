package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/shopspring/decimal"

	"realestate-financial-tool/realestate/financing"
	"realestate-financial-tool/realestate/property"
)

//go:embed templates/*
var templateFS embed.FS

var templates *template.Template

func init() {
	funcMap := template.FuncMap{
		"formatMoney": func(d decimal.Decimal) string {
			if d.LessThan(decimal.Zero) {
				return "-$" + d.Abs().Round(0).String()
			}
			return "$" + d.Round(0).String()
		},
		"formatPercent": func(d decimal.Decimal) string {
			return d.Round(2).String() + "%"
		},
		"formatDecimal": func(d decimal.Decimal) string {
			return d.Round(2).String()
		},
	}

	templates = template.Must(template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/*.html"))
}

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/analyze", handleAnalyze)
	http.HandleFunc("/chart/amortization", handleAmortizationChart)
	http.HandleFunc("/chart/summary", handleSummaryChart)

	fmt.Println("Server starting at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleAmortizationChart(w http.ResponseWriter, r *http.Request) {
	purchasePrice, _ := strconv.ParseFloat(r.URL.Query().Get("purchasePrice"), 64)
	downPaymentPct, _ := strconv.ParseFloat(r.URL.Query().Get("downPaymentPct"), 64)
	interestRate, _ := strconv.ParseFloat(r.URL.Query().Get("interestRate"), 64)
	loanTermStr := r.URL.Query().Get("loanTerm")

	var loanTerm financing.LoanTerm
	switch loanTermStr {
	case "15":
		loanTerm = financing.Term15Years
	case "20":
		loanTerm = financing.Term20Years
	default:
		loanTerm = financing.Term30Years
	}

	downPayment := purchasePrice * (downPaymentPct / 100)
	loan := financing.NewLoan(
		int64(purchasePrice),
		int64(downPayment),
		interestRate, // Already a percentage (e.g., 6 for 6%)
		loanTerm,
		decimal.Zero,
	)

	html, err := loan.Plot()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleSummaryChart(w http.ResponseWriter, r *http.Request) {
	purchasePrice, _ := strconv.ParseFloat(r.URL.Query().Get("purchasePrice"), 64)
	downPaymentPct, _ := strconv.ParseFloat(r.URL.Query().Get("downPaymentPct"), 64)
	interestRate, _ := strconv.ParseFloat(r.URL.Query().Get("interestRate"), 64)
	loanTermStr := r.URL.Query().Get("loanTerm")

	var loanTerm financing.LoanTerm
	switch loanTermStr {
	case "15":
		loanTerm = financing.Term15Years
	case "20":
		loanTerm = financing.Term20Years
	default:
		loanTerm = financing.Term30Years
	}

	downPayment := purchasePrice * (downPaymentPct / 100)
	loan := financing.NewLoan(
		int64(purchasePrice),
		int64(downPayment),
		interestRate, // Already a percentage (e.g., 6 for 6%)
		loanTerm,
		decimal.Zero,
	)

	html, err := loan.PlotSummary()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Build property from form
	p := property.New(r.FormValue("name"))
	p.At(
		r.FormValue("address"),
		r.FormValue("city"),
		r.FormValue("state"),
		r.FormValue("zipCode"),
	)

	// Parse numeric fields
	yearBuilt, _ := strconv.Atoi(r.FormValue("yearBuilt"))
	buildingSF, _ := strconv.ParseFloat(r.FormValue("buildingSF"), 64)
	lotSF, _ := strconv.ParseFloat(r.FormValue("lotSF"), 64)
	p.Built(yearBuilt, buildingSF, lotSF)

	// Parse units
	unitBeds := r.Form["unitBeds[]"]
	unitBaths := r.Form["unitBaths[]"]
	unitSqft := r.Form["unitSqft[]"]
	unitRent := r.Form["unitRent[]"]

	for i := 0; i < len(unitBeds); i++ {
		beds, _ := strconv.Atoi(unitBeds[i])
		baths, _ := strconv.Atoi(unitBaths[i])
		sqft, _ := strconv.ParseFloat(unitSqft[i], 64)
		rent, _ := strconv.ParseFloat(unitRent[i], 64)
		if rent > 0 {
			p.AddUnit(beds, baths, sqft, rent)
		}
	}

	// Parse financials
	purchasePrice, _ := strconv.ParseFloat(r.FormValue("purchasePrice"), 64)
	askingPrice, _ := strconv.ParseFloat(r.FormValue("askingPrice"), 64)
	if askingPrice == 0 {
		askingPrice = purchasePrice
	}
	p.Purchase(purchasePrice, askingPrice)

	interestRate, _ := strconv.ParseFloat(r.FormValue("interestRate"), 64)
	loanTermStr := r.FormValue("loanTerm")
	var loanTerm financing.LoanTerm
	switch loanTermStr {
	case "15":
		loanTerm = financing.Term15Years
	case "20":
		loanTerm = financing.Term20Years
	default:
		loanTerm = financing.Term30Years
	}
	p.Loan(interestRate*100, loanTerm) // Convert percent to basis points

	downPaymentPct, _ := strconv.ParseFloat(r.FormValue("downPaymentPct"), 64)
	if downPaymentPct > 0 {
		p.WithDownPaymentPercent(downPaymentPct)
	}

	// Parse expenses
	taxes, _ := strconv.ParseFloat(r.FormValue("taxes"), 64)
	insurance, _ := strconv.ParseFloat(r.FormValue("insurance"), 64)
	utilities, _ := strconv.ParseFloat(r.FormValue("utilities"), 64)
	maintenance, _ := strconv.ParseFloat(r.FormValue("maintenance"), 64)
	p.Expenses(taxes, insurance, utilities, maintenance)

	vacancyRate, _ := strconv.ParseFloat(r.FormValue("vacancyRate"), 64)
	p.Vacancy(vacancyRate / 100) // Convert percent to decimal

	// Run analysis
	result := p.RunFullAnalysis()

	// Render results to buffer first to avoid superfluous WriteHeader
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, "results.html", result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(buf.Bytes())
}
