package main

import (
	"fmt"
	"math"

	"github.com/google/uuid"

	"realestate-financial-tool/realestate/financing"
	"realestate-financial-tool/realestate/property"
)

func main() {

	p := property.Property{
		PropertyName:  "Test Property",
		Address:       "123 Main St",
		City:          "Testville",
		State:         "TS",
		ZipCode:       "12345",
		County:        "Test County",
		YearBuilt:     2000,
		NumberOfUnits: 2,
		BuildingSF:    2000,
		LotSF:         5000,
		Units: property.Units{
			&property.Unit{
				Uuid:      uuid.New(),
				Name:      "Unit 2",
				Rent:      1000,
				Size:      1000,
				Bedrooms:  1,
				Bathrooms: 1,
				Occupied:  true,
			},
			&property.Unit{
				Uuid:      uuid.New(),
				Name:      "Unit 2",
				Rent:      1000,
				Size:      1000,
				Bedrooms:  1,
				Bathrooms: 1,
				Occupied:  true,
			},
		},
		Financial: &property.Financial{
			AskingPrice:   300_000,
			PurchasePrice: 300_000,
			DownPayment:   60_000,
			LoanAmount:    240_000,
			InterestRate:  5,
			LoanTermYears: financing.Term30Years,
			Expenses: property.ExpensesMonthly{
				Taxes:              math.Round(float64(300_000) * 0.01 / 12),
				Insurance:          500,
				Utilities:          300,
				RepairsMaintenance: 200,
			},
		},
	}

	fmt.Println(property.CalculateMetrics(&p).String())
	fmt.Println(float64(300_000) * (0.01 / 12))
	fmt.Println(p.Financial.Expenses.TotalMonthly())
}
