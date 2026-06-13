// Package money provides a safe int64-based money type for NZD amounts.
// All amounts are stored as cents (smallest currency unit) to avoid
// floating-point rounding errors.
package money

import (
	"fmt"
	"math"
)

// NZD is an amount in New Zealand cents (1 NZD = 100 NZD cents).
// Always use int64 for money — never float64.
type NZD int64

// FromDollars converts a dollar amount to NZD cents.
// Example: FromDollars(10, 50) == 1050 (NZD $10.50)
func FromDollars(dollars, cents int64) NZD {
	return NZD(dollars*100 + cents)
}

// Dollars returns the whole dollar portion of the amount.
func (m NZD) Dollars() int64 {
	return int64(m) / 100
}

// Cents returns the cents portion of the amount (0–99).
func (m NZD) Cents() int64 {
	c := int64(m) % 100
	if c < 0 {
		c = -c
	}
	return c
}

// String formats the amount as a NZD string, e.g. "NZD $10.50".
func (m NZD) String() string {
	if m < 0 {
		return fmt.Sprintf("-NZD $%d.%02d", (-m).Dollars(), (-m).Cents())
	}
	return fmt.Sprintf("NZD $%d.%02d", m.Dollars(), m.Cents())
}

// Add returns m + other.
func (m NZD) Add(other NZD) NZD { return m + other }

// Sub returns m - other.
func (m NZD) Sub(other NZD) NZD { return m - other }

// IsZero returns true if the amount is zero.
func (m NZD) IsZero() bool { return m == 0 }

// IsNegative returns true if the amount is negative.
func (m NZD) IsNegative() bool { return m < 0 }

// GST returns the GST component (15%) of the amount, rounded to nearest cent.
func (m NZD) GST() NZD {
	return NZD(math.Round(float64(m) * 3 / 23))
}

// ExGST returns the amount excluding GST (15%), rounded to nearest cent.
func (m NZD) ExGST() NZD {
	return m - m.GST()
}
