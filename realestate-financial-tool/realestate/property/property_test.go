package property

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var tests = []struct {
	name          string
	propertyName  string
	address       string
	city          string
	state         string
	zipCode       string
	county        string
	yearBuilt     int
	numberOfUnits int
	buildingSF    float64
	lotSF         float64
	taxes         float64
	insurance     float64
	utilities     float64
	repairs       float64
}{
	{
		name:          "basic",
		propertyName:  "Test Property",
		address:       "123 Main St",
		city:          "Testville",
		state:         "TS",
		zipCode:       "12345",
		county:        "Test County",
		yearBuilt:     2000,
		numberOfUnits: 4,
		buildingSF:    2000,
		lotSF:         5000,
		taxes:         1000,
		insurance:     500,
		utilities:     300,
		repairs:       200,
	},
}

func TestNewProperty(t *testing.T) {

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop := NewProperty(
				tt.propertyName, tt.address, tt.city, tt.state, tt.zipCode, tt.county,
				tt.yearBuilt, tt.numberOfUnits, tt.buildingSF, tt.lotSF,
			)
			assert.Equal(t, tt.propertyName, prop.PropertyName, "PropertyName")
			assert.Equal(t, tt.address, prop.Address, "Address")
			assert.Equal(t, tt.city, prop.City, "City")
			assert.Equal(t, tt.state, prop.State, "State")
			assert.Equal(t, tt.zipCode, prop.ZipCode, "ZipCode")
			assert.Equal(t, tt.county, prop.County, "County")
			assert.Equal(t, tt.yearBuilt, prop.YearBuilt, "YearBuilt")
			assert.Equal(t, tt.numberOfUnits, prop.NumberOfUnits, "NumberOfUnits")
			assert.Equal(t, tt.buildingSF, prop.BuildingSF, "BuildingSF")
			assert.Equal(t, tt.lotSF, prop.LotSF, "LotSF")
			assert.Equal(t, tt.taxes, prop.Financial.Expenses.Taxes, "Expenses.Taxes")
			assert.Equal(t, tt.insurance, prop.Financial.Expenses.Insurance, "Expenses.Insurance")
			assert.Equal(t, tt.utilities, prop.Financial.Expenses.Utilities, "Expenses.Utilities")
			assert.Equal(t, tt.repairs, prop.Financial.Expenses.RepairsMaintenance, "Expenses.RepairsMaintenance")
		})
	}
}

func TestProperty_Expenses(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop := NewProperty(
				tt.propertyName, tt.address, tt.city, tt.state, tt.zipCode, tt.county,
				tt.yearBuilt, tt.numberOfUnits, tt.buildingSF, tt.lotSF,
			)
			prop.Financial.SetExpenses(tt.taxes, tt.insurance, tt.utilities, tt.repairs)
			assert.Equal(t, tt.taxes+tt.insurance+tt.utilities+tt.repairs, prop.Financial.Expenses.TotalYearly(), "Total Expenses")
		})
	}
}
