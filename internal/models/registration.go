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

import "encoding/xml"

type (
	REGRESPACK struct {
		XMLName        xml.Name    `xml:"EFDMS"`
		Text           string      `xml:",chardata"`
		EFDMSRESP      REGDATARESP `xml:"EFDMSRESP"`
		EFDMSSIGNATURE string      `xml:"EFDMSSIGNATURE"`
	}
)

// REGDATARESP is the response message received from the VFD
// after a successful registration.
type REGDATARESP struct {
	XMLName     xml.Name `xml:"EFDMSRESP"`
	Text        string   `xml:",chardata"`
	ACKCODE     string   `xml:"ACKCODE"`
	ACKMSG      string   `xml:"ACKMSG"`
	REGID       string   `xml:"REGID"`
	SERIAL      string   `xml:"SERIAL"`
	UIN         string   `xml:"UIN"`
	TIN         string   `xml:"TIN"`
	VRN         string   `xml:"VRN"`
	MOBILE      string   `xml:"MOBILE"`
	ADDRESS     string   `xml:"ADDRESS"`
	STREET      string   `xml:"STREET"`
	CITY        string   `xml:"CITY"`
	COUNTRY     string   `xml:"COUNTRY"`
	NAME        string   `xml:"NAME"`
	RECEIPTCODE string   `xml:"RECEIPTCODE"`
	REGION      string   `xml:"REGION"`
	ROUTINGKEY  string   `xml:"ROUTINGKEY"`
	GC          int64    `xml:"GC"`
	TAXOFFICE   string   `xml:"TAXOFFICE"`
	USERNAME    string   `xml:"USERNAME"`
	PASSWORD    string   `xml:"PASSWORD"`
	TOKENPATH   string   `xml:"TOKENPATH"`
	TAXCODES    TAXCODES `xml:"TAXCODES"`
}

type TAXCODES struct {
	XMLName xml.Name `xml:"TAXCODES"`
	Text    string   `xml:",chardata"`
	CODEA   string   `xml:"CODEA"`
	CODEB   string   `xml:"CODEB"`
	CODEC   string   `xml:"CODEC"`
	CODED   string   `xml:"CODED"`
}

type REGDATAEFDMS struct {
	XMLName        xml.Name `xml:"EFDMS"`
	Text           string   `xml:",chardata"`
	REGDATA        REGDATA  `xml:"REGDATA"`
	EFDMSSIGNATURE string   `xml:"EFDMSSIGNATURE"`
}

type REGDATA struct {
	XMLName xml.Name `xml:"REGDATA"`
	Text    string   `xml:",chardata"`
	TIN     string   `xml:"TIN"`
	CERTKEY string   `xml:"CERTKEY"`
}

type Error struct {
	XMLName xml.Name `xml:"Error"`
	Text    string   `xml:",chardata"`
	Message string   `xml:"Message"`
}
