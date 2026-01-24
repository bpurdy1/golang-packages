package financing

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestInterestRate(t *testing.T) {
	tests := []struct {
		name    string
		input   float64
		wantDec decimal.Decimal
		wantPts float64
		wantStr string
		wantAnn float64
	}{
		{
			name:    "5 percent",
			input:   5.0,
			wantDec: decimal.NewFromFloat(0.05),
			wantPts: 500,
			wantStr: "5%",
			wantAnn: 5.0,
		},
		{
			name:    "3.25 percent",
			input:   3.25,
			wantDec: decimal.NewFromFloat(0.0325),
			wantPts: 325,
			wantStr: "3.25%",
			wantAnn: 3.25,
		},
		{
			name:    "0 percent",
			input:   0.0,
			wantDec: decimal.NewFromFloat(0.0),
			wantPts: 0,
			wantStr: "0%",
			wantAnn: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ir := NewInterestRate(tt.input).(*interestRate)

			if !ir.Decimal().Equal(tt.wantDec) {
				t.Errorf("Decimal() = %v, want %v", ir.Decimal(), tt.wantDec)
			}
			if ir.Points() != decimal.NewFromFloat(tt.wantPts) {
				t.Errorf("Points() = %v, want %v", ir.Points(), decimal.NewFromFloat(tt.wantPts))
			}
			if ir.String() != tt.wantStr {
				t.Errorf("String() = %v, want %v", ir.String(), tt.wantStr)
			}
			if ir.AnnualRate() != tt.wantAnn {
				t.Errorf("AnnualRate() = %v, want %v", ir.AnnualRate(), tt.wantAnn)
			}
		})
	}
}
