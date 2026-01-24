package property

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

type Unit struct {
	Uuid      uuid.UUID
	Name      string
	Bedrooms  int
	Bathrooms int
	Size      float64
	Rent      float64
	Occupied  bool
}

func NewUnit(beds, baths int, squareFootage float64) *Unit {
	return &Unit{
		Uuid:      uuid.New(),
		Bedrooms:  beds,
		Bathrooms: baths,
		Size:      squareFootage,
	}
}
func (unit *Unit) SetRent(i float64) {
	unit.Rent = i
}

func (unit *Unit) IncomePerBedRoom() (float64, error) {
	if unit.Rent == 0 {
		return 0, fmt.Errorf("Unit Income Not Set")
	}
	return float64(unit.Bedrooms) / unit.Rent, nil
}

type Units []*Unit

func (u *Units) AddUnit(beds, baths int, squareFootage float64) {
	nu := NewUnit(beds, baths, squareFootage)
	nu.Name = "unit" + strconv.Itoa(len(*u)+1)
	*u = append(*u, nu)
}

func (u Units) GetUnit(id string) *Unit {
	for _, unit := range u {
		if unit.Name == id || unit.Uuid.String() == id {
			return unit
		}
	}
	return nil
}
