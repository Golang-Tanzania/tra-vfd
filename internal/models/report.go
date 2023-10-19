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

type (
	// ZACK is the Z report acknowledge received from vfd server after
	// successfully Z report upload request
	ZACK struct {
		XMLName xml.Name `xml:"ZACK"`
		Text    string   `xml:",chardata"`
		ZNUMBER int64    `xml:"ZNUMBER"`
		DATE    string   `xml:"DATE"`
		TIME    string   `xml:"TIME"`
		ACKCODE int64    `xml:"ACKCODE"`
		ACKMSG  string   `xml:"ACKMSG"`
	}

	ReportAckEFDMS struct {
		XMLName        xml.Name `xml:"EFDMS"`
		Text           string   `xml:",chardata"`
		ZACK           ZACK     `xml:"ZACK"`
		EFDMSSIGNATURE string   `xml:"EFDMSSIGNATURE"`
	}

	ZREPORT struct {
		XMLName xml.Name `xml:"ZREPORT"`
		Text    string   `xml:",chardata"`
		DATE    string   `xml:"DATE"`
		TIME    string   `xml:"TIME"`
		HEADER  struct {
			Text string   `xml:",chardata"`
			LINE []string `xml:"LINE"`
		} `xml:"HEADER"`
		VRN              string       `xml:"VRN"`
		TIN              string       `xml:"TIN"`
		TAXOFFICE        string       `xml:"TAXOFFICE"`
		REGID            string       `xml:"REGID"`
		ZNUMBER          string       `xml:"ZNUMBER"`
		EFDSERIAL        string       `xml:"EFDSERIAL"`
		REGISTRATIONDATE string       `xml:"REGISTRATIONDATE"`
		USER             string       `xml:"USER"`
		SIMIMSI          string       `xml:"SIMIMSI"`
		TOTALS           REPORTTOTALS `xml:"TOTALS"`
		VATTOTALS        VATTOTALS    `xml:"VATTOTALS"`
		PAYMENTS         PAYMENTS     `xml:"PAYMENTS"`
		CHANGES          struct {
			Text          string `xml:",chardata"`
			VATCHANGENUM  string `xml:"VATCHANGENUM"`
			HEADCHANGENUM string `xml:"HEADCHANGENUM"`
		} `xml:"CHANGES"`
		ERRORS     string `xml:"ERRORS"`
		FWVERSION  string `xml:"FWVERSION"`
		FWCHECKSUM string `xml:"FWCHECKSUM"`
	}

	REPORTTOTALS struct {
		XMLName          xml.Name `xml:"TOTALS"`
		Text             string   `xml:",chardata"`
		DAILYTOTALAMOUNT float64  `xml:"DAILYTOTALAMOUNT"`
		GROSS            float64  `xml:"GROSS"`
		CORRECTIONS      float64  `xml:"CORRECTIONS"`
		DISCOUNTS        float64  `xml:"DISCOUNTS"`
		SURCHARGES       float64  `xml:"SURCHARGES"`
		TICKETSVOID      int64    `xml:"TICKETSVOID"`
		TICKETSVOIDTOTAL float64  `xml:"TICKETSVOIDTOTAL"`
		TICKETSFISCAL    int64    `xml:"TICKETSFISCAL"`
		TICKETSNONFISCAL int64    `xml:"TICKETSNONFISCAL"`
	}
)

// RoundOff is a helper function to round off all the ZREPORT values with data type
// float64 to 2 decimal places
func (z *ZREPORT) RoundOff() {
	z.TOTALS.DAILYTOTALAMOUNT = math.Round(z.TOTALS.DAILYTOTALAMOUNT*100) / 100
	z.TOTALS.GROSS = math.Round(z.TOTALS.GROSS*100) / 100
	z.TOTALS.CORRECTIONS = math.Round(z.TOTALS.CORRECTIONS*100) / 100
	z.TOTALS.DISCOUNTS = math.Round(z.TOTALS.DISCOUNTS*100) / 100
	z.TOTALS.SURCHARGES = math.Round(z.TOTALS.SURCHARGES*100) / 100
	z.TOTALS.TICKETSVOIDTOTAL = math.Round(z.TOTALS.TICKETSVOIDTOTAL*100) / 100
}
