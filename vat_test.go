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

package vfd_test

import (
	"testing"

	vfd "github.com/Golang-Tanzania/tra-vfd"
)

func TestNetAmountAndTaxAmount(t *testing.T) {
	t.Parallel()
	type testCase struct {
		name    string
		vatCode int64
		amount  float64
	}

	testCases := []testCase{
		{"taxable items", vfd.TaxableItemCode, 10000},
		{"non-taxable items", vfd.NonTaxableItemCode, 10000},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			netAmount := vfd.NetAmount(tc.vatCode, tc.amount)
			taxAmount := vfd.ValueAddedTaxAmount(tc.vatCode, tc.amount)
			totalAmount := netAmount + taxAmount
			if totalAmount != tc.amount {
				t.Errorf("totalAmount %.2f != amount %.2f", totalAmount, tc.amount)
			}
			t.Logf("netAmount: %.2f, taxAmount: %.2f, totalAmount: %.2f",
				netAmount, taxAmount, totalAmount)
		})
	}
}

func TestReportTaxRateID(t *testing.T) {
	type args struct {
		taxCode int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "taxable items",
			args: args{taxCode: vfd.TaxableItemCode},
			want: "A-18.00",
		},
		{
			name: "non-taxable items",
			args: args{taxCode: vfd.NonTaxableItemCode},
			want: "C-0.00",
		},
		{
			name: "standard vat",
			args: args{taxCode: vfd.StandardVATCODE},
			want: "A-18.00",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := vfd.ReportTaxRateID(tt.args.taxCode); got != tt.want {
				t.Errorf("ReportTaxRateID() = %v, want %v", got, tt.want)
			}
		})
	}
}
