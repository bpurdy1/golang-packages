package financing

import (
	"fmt"

	"github.com/shopspring/decimal"
)

const pointDevider = 100

// InterestRate interface defines the methods for interest rate handling
type InterestRate interface {
	Decimal() decimal.Decimal
	Points() decimal.Decimal
	String() string
	AnnualRate() float64
}

// interestRate struct implements InterestRate interface
type interestRate struct {
	rate decimal.Decimal // decimal representation, e.g., 0.05 for 5%
}

// NewInterestRate creates a new InterestRate from a float64 percent (e.g., 5.0 for 5%)
func NewInterestRate(percent float64) InterestRate {
	r := &interestRate{
		// int to float persent conversion
		rate: decimal.NewFromFloat(percent).Div(decimal.NewFromInt(pointDevider)),
	}
	return r
}

func (ir *interestRate) Decimal() decimal.Decimal {
	return ir.rate
}

func (ir *interestRate) Points() decimal.Decimal {
	// float to points 100* int
	return ir.rate.Mul(decimal.NewFromInt(pointDevider * 100))
}

func (ir *interestRate) String() string {
	return fmt.Sprintf("%g%%", ir.rate.Mul(decimal.NewFromInt(pointDevider)).InexactFloat64())
}

func (ir *interestRate) AnnualRate() float64 {
	return ir.rate.Mul(decimal.NewFromInt(pointDevider)).InexactFloat64()
}
