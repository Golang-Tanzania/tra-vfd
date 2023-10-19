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
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	xhttp "github.com/Golang-Tanzania/tra-vfd/internal/http"

	"github.com/Golang-Tanzania/tra-vfd/internal/models"
)

var ErrReportSubmitFailed = fmt.Errorf("report submit failed")

type (
	// ReportTotals contains different number of totals
	ReportTotals struct {
		DailyTotalAmount float64
		Gross            float64
		Corrections      float64
		Discounts        float64
		Surcharges       float64
		TicketsVoid      int64
		TicketsVoidTotal float64
		TicketsFiscal    int64
		TicketsNonFiscal int64
	}

	Address struct {
		Name    string
		Street  string
		Mobile  string
		City    string
		Country string
	}

	ReportParams struct {
		Date             string
		Time             string
		VRN              string
		TIN              string
		UIN              string
		TaxOffice        string
		RegistrationID   string
		ZNumber          string
		EFDSerial        string
		RegistrationDate string
	}

	ReportRequest struct {
		Params  *ReportParams
		Address *Address
		Totals  *ReportTotals
		VATS    []VATTOTAL
		Payment []Payment
	}
)

// submitReport submits a report to the VFD server.
func submitReport(ctx context.Context, client *http.Client, requestURL string, headers *RequestHeaders,
	privateKey *rsa.PrivateKey,
	report *ReportRequest,
) (*Response, error) {
	var (
		certSerial  = headers.CertSerial
		bearerToken = headers.BearerToken
	)

	newContext, cancel := context.WithCancel(ctx)
	defer cancel()

	payload, err := ReportBytes(
		privateKey, report.Params, *report.Address, report.VATS,
		report.Payment, *report.Totals)
	if err != nil {
		return nil, fmt.Errorf("failed to generate the report payload: %w", err)
	}

	req, err := http.NewRequestWithContext(newContext, http.MethodPost, requestURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", ContentTypeXML)
	req.Header.Set("Routing-Key", SubmitReportRoutingKey)
	req.Header.Set("Cert-Serial", encodeBase64String(certSerial))
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", bearerToken))

	resp, err := client.Do(req)
	if err != nil {
		return nil, checkNetworkError(newContext, "submit report", err)
	}
	defer resp.Body.Close()

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%v : %w", ErrReportSubmitFailed, err)
	}

	if resp.StatusCode == http.StatusInternalServerError {
		errBody := models.Error{}
		err = xml.NewDecoder(bytes.NewBuffer(out)).Decode(&errBody)
		if err != nil {
			return nil, fmt.Errorf("%v : %w", ErrReportSubmitFailed, err)
		}

		return nil, fmt.Errorf("registration error: %s", errBody.Message)
	}

	response := models.ReportAckEFDMS{}
	err = xml.NewDecoder(bytes.NewBuffer(out)).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("%v : %w", ErrReportSubmitFailed, err)
	}

	return &Response{
		Number:  response.ZACK.ZNUMBER,
		Date:    response.ZACK.DATE,
		Time:    response.ZACK.TIME,
		Code:    response.ZACK.ACKCODE,
		Message: response.ZACK.ACKMSG,
	}, nil
}

func SubmitReport(ctx context.Context, url string, headers *RequestHeaders, privateKey *rsa.PrivateKey,
	report *ReportRequest,
) (*Response, error) {
	client := xhttp.Instance()
	return submitReport(ctx, client, url, headers, privateKey, report)
}

func (lines *Address) AsList() []string {
	return []string{
		strings.ToUpper(lines.Name),
		strings.ToUpper(lines.Street),
		fmt.Sprintf("MOBILE: %s", lines.Mobile),
		strings.ToUpper(fmt.Sprintf("%s,%s", lines.City, lines.Country)),
	}
}

// sumVatTotals
func sumVatTotals(vats []VATTOTAL) models.VATTOTALS {
	vatTotalMap := map[string]struct {
		NetAmount float64
		TaxAmount float64
	}{
		"A-18.00": {0, 0},
		"B-0.00":  {0, 0},
		"C-0.00":  {0, 0},
		"D-0.00":  {0, 0},
		"E-0.00":  {0, 0},
	}

	for _, vat := range vats {
		rate := fmt.Sprintf("%s-%.2f", vat.ID, vat.Rate)
		vatTotalMap[rate] = struct {
			NetAmount float64
			TaxAmount float64
		}{
			vatTotalMap[rate].NetAmount + vat.NetAmount,
			vatTotalMap[rate].TaxAmount + vat.TaxAmount,
		}
	}

	return models.VATTOTALS{
		VATTOTAL: []*models.VATTOTAL{
			{
				VATRATE:    "A-18.00",
				NETTAMOUNT: fmt.Sprintf("%.2f", vatTotalMap["A-18.00"].NetAmount),
				TAXAMOUNT:  fmt.Sprintf("%.2f", vatTotalMap["A-18.00"].TaxAmount),
			},
			{
				VATRATE:    "B-0.00",
				NETTAMOUNT: fmt.Sprintf("%.2f", vatTotalMap["B-0.00"].NetAmount),
				TAXAMOUNT:  fmt.Sprintf("%.2f", vatTotalMap["B-0.00"].TaxAmount),
			},
			{
				VATRATE:    "C-0.00",
				NETTAMOUNT: fmt.Sprintf("%.2f", vatTotalMap["C-0.00"].NetAmount),
				TAXAMOUNT:  fmt.Sprintf("%.2f", vatTotalMap["C-0.00"].TaxAmount),
			},
			{
				VATRATE:    "D-0.00",
				NETTAMOUNT: fmt.Sprintf("%.2f", vatTotalMap["D-0.00"].NetAmount),
				TAXAMOUNT:  fmt.Sprintf("%.2f", vatTotalMap["D-0.00"].TaxAmount),
			},
			{
				VATRATE:    "E-0.00",
				NETTAMOUNT: fmt.Sprintf("%.2f", vatTotalMap["E-0.00"].NetAmount),
				TAXAMOUNT:  fmt.Sprintf("%.2f", vatTotalMap["E-0.00"].TaxAmount),
			},
		},
	}
}

// sumPayments sums all payments
func sumPayments(payments []Payment) models.PAYMENTS {
	paymentMap := map[string]float64{
		"CASH":    0.0,
		"CHEQUE":  0.0,
		"CCARD":   0.0,
		"EMONEY":  0.0,
		"INVOICE": 0.0,
	}

	for _, p := range payments {
		pType := string(p.Type)
		paymentMap[pType] += p.Amount
	}

	// paymentList contains a list of payments and the order
	// should be the same as in the paymentMap
	paymentsList := make([]*models.PAYMENT, 5)
	paymentsList[0] = &models.PAYMENT{
		PMTTYPE:   "CASH",
		PMTAMOUNT: fmt.Sprintf("%.2f", paymentMap["CASH"]),
	}
	paymentsList[1] = &models.PAYMENT{
		PMTTYPE:   "CHEQUE",
		PMTAMOUNT: fmt.Sprintf("%.2f", paymentMap["CHEQUE"]),
	}
	paymentsList[2] = &models.PAYMENT{
		PMTTYPE:   "CCARD",
		PMTAMOUNT: fmt.Sprintf("%.2f", paymentMap["CCARD"]),
	}

	paymentsList[3] = &models.PAYMENT{
		PMTTYPE:   "EMONEY",
		PMTAMOUNT: fmt.Sprintf("%.2f", paymentMap["EMONEY"]),
	}

	paymentsList[4] = &models.PAYMENT{
		PMTTYPE:   "INVOICE",
		PMTAMOUNT: fmt.Sprintf("%.2f", paymentMap["INVOICE"]),
	}

	return models.PAYMENTS{
		PAYMENT: paymentsList,
	}
}

func generateZReport(params *ReportParams, address Address, vats []VATTOTAL, payments []Payment, totals ReportTotals) *models.ZREPORT {
	const (
		SIMIMSI       = "WEBAPI"
		FWVERSION     = "3.0"
		FWCHECKSUM    = "WEBAPI"
		VATCHANGENUM  = "0"
		HEADCHANGENUM = "0"
		ERRORS        = ""
	)

	PAYMENTS := sumPayments(payments)
	VATTOTALS := sumVatTotals(vats)

	TT := models.REPORTTOTALS{
		DAILYTOTALAMOUNT: totals.DailyTotalAmount,
		GROSS:            totals.Gross,
		CORRECTIONS:      totals.Corrections,
		DISCOUNTS:        totals.Discounts,
		SURCHARGES:       totals.Surcharges,
		TICKETSVOID:      totals.TicketsVoid,
		TICKETSVOIDTOTAL: totals.TicketsVoidTotal,
		TICKETSFISCAL:    totals.TicketsFiscal,
		TICKETSNONFISCAL: totals.TicketsNonFiscal,
	}
	report := &models.ZREPORT{
		XMLName: xml.Name{},
		Text:    "",
		DATE:    params.Date,
		TIME:    params.Time,
		HEADER: struct {
			Text string   `xml:",chardata"`
			LINE []string `xml:"LINE"`
		}{
			LINE: address.AsList(),
		},
		VRN:              params.VRN,
		TIN:              params.TIN,
		TAXOFFICE:        params.TaxOffice,
		REGID:            params.RegistrationID,
		ZNUMBER:          params.ZNumber,
		EFDSERIAL:        params.EFDSerial,
		REGISTRATIONDATE: params.RegistrationDate,
		USER:             params.UIN,
		SIMIMSI:          SIMIMSI,
		TOTALS:           TT,
		VATTOTALS:        VATTOTALS,
		PAYMENTS:         PAYMENTS,
		CHANGES: struct {
			Text          string `xml:",chardata"`
			VATCHANGENUM  string `xml:"VATCHANGENUM"`
			HEADCHANGENUM string `xml:"HEADCHANGENUM"`
		}{
			VATCHANGENUM:  VATCHANGENUM,
			HEADCHANGENUM: HEADCHANGENUM,
		},
		ERRORS:     ERRORS,
		FWVERSION:  FWVERSION,
		FWCHECKSUM: FWCHECKSUM,
	}

	// report.RoundOff()

	return report
}

// ReportBytes returns the bytes of the report payload. It calls xml.Marshal on the report.
// then replace all the occurrences of <PAYMENT>, </PAYMENT>, <VATTOTAL>, </VATTOTAL> with empty string ""
// and then add the xml.Header to the beginning of the payload.
func ReportBytes(privateKey *rsa.PrivateKey, params *ReportParams, address Address,
	vats []VATTOTAL, payments []Payment,
	totals ReportTotals,
) ([]byte, error) {
	zReport := generateZReport(params, address, vats, payments, totals)
	payload, err := xml.Marshal(zReport)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal the report: %w", err)
	}
	payloadString := formatReportXmlPayload(payload, totals, vats, payments)
	signedPayload, err := SignPayload(privateKey, []byte(payloadString))
	if err != nil {
		return nil, fmt.Errorf("failed to sign the payload: %w", err)
	}
	base64PayloadSignature := encodeBase64Bytes(signedPayload)
	report := fmt.Sprintf("<EFDMS>%s<EFDMSSIGNATURE>%s</EFDMSSIGNATURE></EFDMS>", payloadString, base64PayloadSignature)
	report = fmt.Sprintf("%s%s", xml.Header, report)

	return []byte(report), nil
}

func formatReportXmlPayload(payload []byte, totals ReportTotals, vats []VATTOTAL, payments []Payment) string {
	replaceList := []string{"<PAYMENT>", "", "</PAYMENT>", "", "<VATTOTAL>", "", "</VATTOTAL>", ""}
	replacer := strings.NewReplacer(replaceList...)
	payloadString := replacer.Replace(string(payload))

	var (
		regexDailyAmount = regexp.MustCompile(`<DAILYTOTALAMOUNT>.*</DAILYTOTALAMOUNT>`)
		regexGrossAmount = regexp.MustCompile(`<GROSS>.*</GROSS>`)
		dailyAmountTag   = fmt.Sprintf("<DAILYTOTALAMOUNT>%.2f</DAILYTOTALAMOUNT>", totals.DailyTotalAmount)
		grossAmountTag   = fmt.Sprintf("<GROSS>%.2f</GROSS>", totals.Gross)
	)

	payloadString = regexDailyAmount.ReplaceAllString(payloadString, dailyAmountTag)
	payloadString = regexGrossAmount.ReplaceAllString(payloadString, grossAmountTag)

	return payloadString
}
