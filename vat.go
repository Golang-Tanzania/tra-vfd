/*
 * Copyright (c) 2023 Golang Tanzania
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software, and to permit persons to whom the Software is furnished to do
 * so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED,
 * INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
 * PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
 * HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF
 * CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
 * OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package vfd

import (
	"fmt"
	"math"
)

const (
	StandardVATID        = "A"
	StandardVATRATE      = 18.00
	StandardVATCODE      = 1
	SpecialVATID         = "B"
	SpecialVATRATE       = 0.00
	SpecialVATCODE       = 2
	ZeroVATID            = "C"
	ZeroVATRATE          = 0.00
	ZeroVATCODE          = 3
	SpecialReliefVATID   = "D"
	SpecialReliefVATRATE = 0.00
	SpecialReliefVATCODE = 4
	ExemptedVATID        = "E"
	ExemptedVATRATE      = 0.00
	ExemptedVATCODE      = 5
	TaxableItemCode      = 1
	TaxableItemId        = "A"
	NonTaxableItemCode   = 3
	NonTaxableItemId     = "C"
)

type (
	ValueAddedTax struct {
		ID         string // ID is a character that identifies the ValueAddedTax it can be A,B,C,D or E
		Code       int64  // Code is a number that identifies the ValueAddedTax it can be 0,1,2,3 or 4
		Name       string
		Percentage float64
	}
)

var (
	standardVAT = ValueAddedTax{
		ID:         StandardVATID,
		Code:       StandardVATCODE,
		Name:       "Standard ValueAddedTax",
		Percentage: StandardVATRATE,
	}
	specialVAT = ValueAddedTax{
		ID:         SpecialVATID,
		Code:       SpecialVATCODE,
		Name:       "Special ValueAddedTax",
		Percentage: SpecialVATRATE,
	}
	zeroVAT = ValueAddedTax{
		ID:         ZeroVATID,
		Code:       ZeroVATCODE,
		Name:       "Zero ValueAddedTax",
		Percentage: ZeroVATRATE,
	}
	specialReliefVAT = ValueAddedTax{
		ID:         SpecialReliefVATID,
		Code:       SpecialReliefVATCODE,
		Name:       "Special Relief ValueAddedTax",
		Percentage: SpecialReliefVATRATE,
	}
	exemptedVAT = ValueAddedTax{
		ID:         ExemptedVATID,
		Code:       ExemptedVATCODE,
		Name:       "Exempted ValueAddedTax",
		Percentage: ExemptedVATRATE,
	}
)

func (v *ValueAddedTax) NetAmount(totalAmount float64) float64 {
	rate := 1.00 + (v.Percentage / 100)
	netAmount := totalAmount / rate
	return math.Round(netAmount*100) / 100
}

// Amount calculates the amount of ValueAddedTax that is charged to the buyer.
// The answer is rounded to 2 decimal places.
func (v *ValueAddedTax) Amount(totalAmount float64) float64 {
	netAmount := v.NetAmount(totalAmount)
	amount := totalAmount - netAmount
	return math.Round(amount*100) / 100
}

func ParseTaxCode(code int64) ValueAddedTax {
	switch code {
	case 1:
		return standardVAT
	case 2:
		return specialVAT
	case 3:
		return zeroVAT
	case 4:
		return specialReliefVAT
	case 5:
		return exemptedVAT
	default:
		return standardVAT
	}
}

// ValueAddedTaxRate returns the ValueAddedTax rate of a certain ValueAddedTax category
func ValueAddedTaxRate(taxCode int64) float64 {
	vat := ParseTaxCode(taxCode)
	return vat.Percentage
}

// ValueAddedTaxID returns the ValueAddedTax id of a certain ValueAddedTax category
// It returns "A" for standard ValueAddedTax, "B" for special ValueAddedTax,
// "C" for zero ValueAddedTax,"D" for special relief and "E" for
// exempted ValueAddedTax.
func ValueAddedTaxID(taxCode int64) string {
	vat := ParseTaxCode(taxCode)
	return vat.ID
}

// NetAmount calculates the net price of a product of a certain ValueAddedTax category.
// This is the NetAmount which is collected by the seller without the ValueAddedTax.
// The buyer is charged this NetAmount plus the ValueAddedTax NetAmount. After calculating the
// answer is rounded to 2 decimal places.
// price = netPrice + netPrice * (vatRate / 100)
func NetAmount(taxCode int64, price float64) float64 {
	vat := ParseTaxCode(taxCode)
	return vat.NetAmount(price)
}

// ValueAddedTaxAmount calculates the amount of ValueAddedTax that is charged to the buyer.
// The answer is rounded to 2 decimal places.
func ValueAddedTaxAmount(taxCode int64, price float64) float64 {
	vat := ParseTaxCode(taxCode)
	netAmount := vat.NetAmount(price)
	amount := price - netAmount
	return math.Round(amount*100) / 100
}

// ReportTaxRateID creates a string that contains the ValueAddedTax rate and the ValueAddedTax id
// of a certain ValueAddedTax category. It returns "A-18.00" for standard ValueAddedTax,
// "B-10.00" for special ValueAddedTax, "C-0.00" for zero ValueAddedTax and so on. The ID is then
// used in Z Report to indicate the ValueAddedTax rate and the ValueAddedTax id.
func ReportTaxRateID(taxCode int64) string {
	vat := ParseTaxCode(taxCode)
	return fmt.Sprintf("%s-%.2f", vat.ID, vat.Percentage)
}
