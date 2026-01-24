package main

import (
	"realestate-financial-tool/realestate/financing"
	"realestate-financial-tool/realestate/property"
)

func main() {
	property.New("Maple Street Fourplex").
		At("456 Maple Street", "Austin", "TX", "78701").
		InCounty("Travis").
		Built(1995, 3200, 8000).
		AddUnit(2, 1, 800, 1200). // 2 bed, 1 bath, 800 sqft, $1200/mo
		AddUnit(2, 1, 800, 1200).
		AddUnit(1, 1, 600, 950).
		AddUnit(1, 1, 600, 950).
		Purchase(640_000, 650_000).       // purchase price, asking price
		Loan(600, financing.Term30Years). // 6% interest, 30 year term
		Expenses(333, 125, 200, 200).     // taxes, insurance, utilities, maintenance (monthly)
		Vacancy(0.05).                    // 5% vacancy rate
		RunFullAnalysis().
		Print().
		PrintLoanSummary().
		GenerateCharts()
}
