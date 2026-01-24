package main

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"realestate-financial-tool/internal/gofinancial"
	"realestate-financial-tool/internal/gofinancial/enums/frequency"
	"realestate-financial-tool/internal/gofinancial/enums/interesttype"
	"realestate-financial-tool/internal/gofinancial/enums/paymentperiod"
	"realestate-financial-tool/realestate/financing"
)

func main() {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		panic("location loading error")
	}
	currentDate := time.Date(2009, 11, 11, 4, 30, 0, 0, loc)

	config := gofinancial.Config{
		// start date is inclusive
		StartDate: currentDate,
		// end date is inclusive.
		EndDate:   currentDate.AddDate(30, 0, 0).AddDate(0, 0, -1),
		Frequency: frequency.ANNUALLY,
		// AmountBorrowed is in paisa
		AmountBorrowed: decimal.NewFromInt(200000000),
		// InterestType can be flat or reducing
		InterestType: interesttype.REDUCING,
		// interest is in basis points
		Interest: decimal.NewFromInt(1200),
		// amount is paid at the end of the period
		PaymentPeriod: paymentperiod.ENDING,
		// all values will be rounded
		EnableRounding: true,
		// it will be rounded to nearest int
		RoundingPlaces: 0,
		// no error is tolerated
		RoundingErrorTolerance: decimal.Zero,
	}
	_, err = gofinancial.NewAmortization(&config)
	if err != nil {
		panic(err)
	}

	loan := financing.NewLoan(
		300000,
		60000,
		5,
		financing.Term30Years,
		decimal.NewFromFloat(0),
	)
	fmt.Printf("Loan: %d years, %s to %s\n", loan.TermYears, loan.StartDate.Format("2006-01-02"), loan.EndDate.Format("2006-01-02"))

	output, err := loan.Plot()
	if err != nil {
		panic(err)
	}
	fmt.Println("Plot generated, output length:", len(output))
}
