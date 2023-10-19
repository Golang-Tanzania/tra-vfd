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

package models

import (
	"encoding/xml"
	"math"
)

const hundred = 100

type (
	// RCTACK is the receipt acknowledge received from
	// vfd server after successfully receipt upload request
	RCTACK struct {
		XMLName xml.Name `xml:"RCTACK"`
		Text    string   `xml:",chardata"`
		RCTNUM  int64    `xml:"RCTNUM"`
		DATE    string   `xml:"DATE"`
		TIME    string   `xml:"TIME"`
		ACKCODE int64    `xml:"ACKCODE"`
		ACKMSG  string   `xml:"ACKMSG"`
	}

	RCTACKEFDMS struct {
		XMLName        xml.Name `xml:"EFDMS"`
		Text           string   `xml:",chardata"`
		RCTACK         RCTACK   `xml:"RCTACK"`
		EFDMSSIGNATURE string   `xml:"EFDMSSIGNATURE"`
	}

	RCT struct {
		XMLName    xml.Name  `xml:"RCT"`
		Text       string    `xml:",chardata"`
		DATE       string    `xml:"DATE"`
		TIME       string    `xml:"TIME"`
		TIN        string    `xml:"TIN"`
		REGID      string    `xml:"REGID"`
		EFDSERIAL  string    `xml:"EFDSERIAL"`
		CUSTIDTYPE int64     `xml:"CUSTIDTYPE"`
		CUSTID     string    `xml:"CUSTID"`
		CUSTNAME   string    `xml:"CUSTNAME"`
		MOBILENUM  string    `xml:"MOBILENUM"`
		RCTNUM     string    `xml:"RCTNUM"`
		DC         int64     `xml:"DC"`
		GC         int64     `xml:"GC"`
		ZNUM       string    `xml:"ZNUM"`
		RCTVNUM    string    `xml:"RCTVNUM"`
		ITEMS      ITEMS     `xml:"ITEMS"`
		TOTALS     TOTALS    `xml:"TOTALS"`
		PAYMENTS   PAYMENTS  `xml:"PAYMENTS"`
		VATTOTALS  VATTOTALS `xml:"VATTOTALS"`
	}

	ITEMS struct {
		XMLName xml.Name `xml:"ITEMS"`
		Text    string   `xml:",chardata"`
		ITEM    []*ITEM  `xml:"ITEM"`
	}

	ITEM struct {
		XMLName xml.Name `xml:"ITEM"`
		Text    string   `xml:",chardata"`
		ID      string   `xml:"ID"`
		DESC    string   `xml:"DESC"`
		QTY     float64  `xml:"QTY"`
		TAXCODE int64    `xml:"TAXCODE"`
		AMT     float64  `xml:"AMT"`
	}

	TOTALS struct {
		XMLName      xml.Name `xml:"TOTALS"`
		Text         string   `xml:",chardata"`
		TOTALTAXEXCL float64  `xml:"TOTALTAXEXCL"`
		TOTALTAXINCL float64  `xml:"TOTALTAXINCL"`
		DISCOUNT     float64  `xml:"DISCOUNT"`
	}

	VATTOTAL struct {
		XMLName    xml.Name `xml:"VATTOTAL"`
		Text       string   `xml:",chardata"`
		VATRATE    string   `xml:"VATRATE"`
		NETTAMOUNT string   `xml:"NETTAMOUNT"`
		TAXAMOUNT  string   `xml:"TAXAMOUNT"`
	}

	VATTOTALS struct {
		XMLName  xml.Name    `xml:"VATTOTALS"`
		Text     string      `xml:",chardata"`
		VATTOTAL []*VATTOTAL `xml:"VATTOTAL"`
	}

	PAYMENT struct {
		XMLName   xml.Name `xml:"PAYMENT"`
		Text      string   `xml:",chardata"`
		PMTTYPE   string   `xml:"PMTTYPE"`
		PMTAMOUNT string   `xml:"PMTAMOUNT"`
	}

	PAYMENTS struct {
		XMLName xml.Name   `xml:"PAYMENTS"`
		Text    string     `xml:",chardata"`
		PAYMENT []*PAYMENT `xml:"PAYMENT"`
	}
)

// RoundOff is a helper function to round off the all the values of RCT with float64 as a Data Type
// to 2 decimal places.
func (r *RCT) RoundOff() {
	// RoundOff all the RCT.TOTALS
	r.TOTALS.TOTALTAXEXCL = math.Round(r.TOTALS.TOTALTAXEXCL*hundred) / hundred
	r.TOTALS.TOTALTAXINCL = math.Round(r.TOTALS.TOTALTAXINCL*hundred) / hundred
	r.TOTALS.DISCOUNT = math.Round(r.TOTALS.DISCOUNT*hundred) / hundred
}
