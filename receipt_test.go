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
	"crypto/rand"
	"crypto/rsa"
	"encoding/xml"
	"fmt"
	"reflect"
	"testing"

	"github.com/Golang-Tanzania/tra-vfd/internal/models"
)

func TestProcessItems(t *testing.T) {
	type args struct {
		items []Item
	}
	type result struct {
		ItemsAmount []float64
		VAT         []*models.VATTOTAL
		Totals      models.TOTALS
	}
	tests := []struct {
		name string
		args args
		want *result
	}{
		{
			name: "TestProcessItems",
			args: args{
				items: []Item{
					{
						ID:          "1",
						Description: "Item 1",
						TaxCode:     TaxableItemCode,
						Quantity:    5,
						UnitPrice:   2000,
						Discount:    5000,
					},
				},
			},
			want: &result{
				ItemsAmount: []float64{10000},
				VAT: []*models.VATTOTAL{
					{
						XMLName:    xml.Name{},
						Text:       "",
						VATRATE:    "A",
						NETTAMOUNT: "4237.29",
						TAXAMOUNT:  "762.71",
					},
				},
				Totals: models.TOTALS{
					TOTALTAXEXCL: 4237.29,
					TOTALTAXINCL: 5000.00,
					DISCOUNT:     5000.00,
				},
			},
		},
		{
			name: "TestProcessItems",
			args: args{
				items: []Item{
					{
						ID:          "1",
						Description: "Item 1",
						TaxCode:     TaxableItemCode,
						Quantity:    5,
						UnitPrice:   1000,
						Discount:    0,
					},
				},
			},
			want: &result{
				ItemsAmount: []float64{5000},
				VAT: []*models.VATTOTAL{
					{
						XMLName:    xml.Name{},
						Text:       "",
						VATRATE:    "A",
						NETTAMOUNT: "4237.29",
						TAXAMOUNT:  "762.71",
					},
				},
				Totals: models.TOTALS{
					TOTALTAXEXCL: 4237.29,
					TOTALTAXINCL: 5000.00,
					DISCOUNT:     0.00,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ProcessItems(tt.args.items); !reflect.DeepEqual(got, tt.want) {
				items := got.ITEMS
				for i, item := range items {
					wantItemAmount := tt.want.ItemsAmount[i]
					message := fmt.Sprintf("[GOT]: Item %d: id: %s, quantity: %.2f, amount: %.2f [EXPECTED]: %.2f\n", i, item.ID, item.QTY, item.AMT, wantItemAmount)
					if item.AMT != wantItemAmount {
						t.Errorf("[ERROR] ProcessItems Error(): %s", message)
					}
					t.Logf("[INFO] ProcessItems(): %s", message)
				}

				// Comparing TOTALS
				func(got, want models.TOTALS) {
					if got.TOTALTAXEXCL != want.TOTALTAXEXCL {
						t.Errorf("[ERROR] TOTALTAXEXCL: got %.2f, want %.2f", got.TOTALTAXEXCL, want.TOTALTAXEXCL)
					}
					if got.TOTALTAXINCL != want.TOTALTAXINCL {
						t.Errorf("[ERROR] TOTALTAXINCL: got %.2f, want %.2f", got.TOTALTAXINCL, want.TOTALTAXINCL)
					}
					if got.DISCOUNT != want.DISCOUNT {
						t.Errorf("[ERROR] DISCOUNT: got %.2f, want %.2f", got.DISCOUNT, want.DISCOUNT)
					}
					t.Logf("[INFO] TOTALS: [GOT]: TAXEXCL: %.2f, TAXINCL: %.2f, DISCOUNT: %.2f [EXPECTED]: TAXEXCL: %.2f, TAXINCL: %.2f, DISCOUNT: %.2f ",
						got.TOTALTAXEXCL, got.TOTALTAXINCL, got.DISCOUNT, want.TOTALTAXEXCL, want.TOTALTAXINCL, want.DISCOUNT)
				}(got.TOTALS, tt.want.Totals)

				// Comparing VATS
				func(got, want []*models.VATTOTAL) {
					for i, v := range got {
						if v.TAXAMOUNT != want[i].TAXAMOUNT {
							t.Errorf("[ERROR] VATRATE[%s]: TAXAMOUNT: got %s, want %s", v.VATRATE, v.TAXAMOUNT, want[i].TAXAMOUNT)
						}
						if v.NETTAMOUNT != want[i].NETTAMOUNT {
							t.Errorf("[ERROR] VATRATE[%s].NETTAMOUNT: got %s, want %s", v.VATRATE, v.NETTAMOUNT, want[i].NETTAMOUNT)
						}
						t.Logf("[INFO] VATRATE[%s]: [GOT]: TAXAMOUNT: %s, NETTAMOUNT: %s [EXPECTED]: TAXAMOUNT: %s, NETTAMOUNT: %s ",
							v.VATRATE, v.TAXAMOUNT, v.NETTAMOUNT, want[i].TAXAMOUNT, want[i].NETTAMOUNT)
					}
				}(got.VATTOTALS, tt.want.VAT)

			}
		})
	}
}

func TestReceiptBytes(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Errorf("Error generating private key: %v", err)
	}
	customer := Customer{
		Type:   NIDACustomerID,
		ID:     "516221638383907151",
		Name:   "John doe",
		Mobile: "255765992153",
	}

	items := []Item{
		{
			ID:          "1",
			Description: "Item 1",
			TaxCode:     TaxableItemCode,
			Quantity:    5,
			UnitPrice:   2000,
			Discount:    5000,
		},
	}

	payments := []Payment{
		{
			Type:   CashPaymentType,
			Amount: 5000,
		},
	}

	params := ReceiptParams{
		Date:           "2022-11-17",
		Time:           "14:00:00",
		TIN:            "TQR6W5FWC",
		RegistrationID: "262T3FSSS",
		EFDSerial:      "SGSYSTHSSJ",
		ReceiptNum:     "",
		DailyCounter:   1,
		GlobalCounter:  100,
		ZNum:           "",
		ReceiptVNum:    "",
	}

	got, err := ReceiptBytes(privateKey, params, customer, items, payments)
	if err != nil {
		t.Errorf("Error generating receipt bytes: %v", err)
	}

	t.Logf("Receipt bytes: \n\n%s\n\n", string(got))
}
