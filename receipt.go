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
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Golang-Tanzania/tra-vfd/pkg/env"

	xhttp "github.com/Golang-Tanzania/tra-vfd/internal/http"

	"github.com/Golang-Tanzania/tra-vfd/internal/models"
)

var ErrReceiptUploadFailed = errors.New("receipt upload failed")

type (
	// ReceiptParams contains parameters icluded while sending the receipts
	ReceiptParams struct {
		Date           string
		Time           string
		TIN            string
		RegistrationID string
		EFDSerial      string
		ReceiptNum     string
		DailyCounter   int64
		GlobalCounter  int64
		ZNum           string
		ReceiptVNum    string
	}

	// Customer contains customer information
	Customer struct {
		Type   CustomerID
		ID     string
		Name   string
		Mobile string
	}

	// Item represent a purchased item. TaxCode is an integer that can take the
	// value of 1 for taxable items and 3 for non-taxable items.
	// Discount is for the whole package not a unit discount
	Item struct {
		ID          string
		Description string
		TaxCode     int64
		Quantity    float64
		UnitPrice   float64
		Discount    float64
	}

	ReceiptRequest struct {
		Params   ReceiptParams
		Customer Customer
		Items    []Item
		Payments []Payment
	}
)

// SubmitReceipt uploads a receipt to the VFD server.
func SubmitReceipt(ctx context.Context, requestURL string, headers *RequestHeaders, privateKey *rsa.PrivateKey,
	receiptRequest *ReceiptRequest,
) (*Response, error) {
	client := xhttp.Instance()
	return submitReceipt(ctx, client, requestURL, headers, privateKey, receiptRequest)
}

func submitReceipt(ctx context.Context, client *http.Client, requestURL string, headers *RequestHeaders,
	privateKey *rsa.PrivateKey, rct *ReceiptRequest,
) (*Response, error) {
	var (
		certSerial  = headers.CertSerial
		bearerToken = headers.BearerToken
	)

	newContext, cancel := context.WithCancel(ctx)
	defer cancel()

	payload, err := ReceiptBytes(
		privateKey, rct.Params, rct.Customer, rct.Items, rct.Payments)
	if err != nil {
		return nil, fmt.Errorf("%v : %w", ErrReceiptUploadFailed, err)
	}

	req, err := http.NewRequestWithContext(newContext, http.MethodPost, requestURL,
		bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", ContentTypeXML)
	req.Header.Set("Routing-Key", SubmitReceiptRoutingKey)
	req.Header.Set("Cert-Serial", encodeBase64String(certSerial))
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", bearerToken))

	resp, err := client.Do(req)
	if err != nil {
		return nil, checkNetworkError(newContext, "receipt upload", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "registration: could not close response body %v", err)
		}
	}(resp.Body)

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%v : %w", ErrReceiptUploadFailed, err)
	}

	if resp.StatusCode == 500 {
		errBody := models.Error{}
		err = xml.NewDecoder(bytes.NewBuffer(out)).Decode(&errBody)
		if err != nil {
			return nil, fmt.Errorf("%v : %w", ErrReceiptUploadFailed, err)
		}

		return nil, fmt.Errorf("registration error: %s", errBody.Message)
	}

	response := models.RCTACKEFDMS{}
	err = xml.NewDecoder(bytes.NewBuffer(out)).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("%v : %w", ErrReceiptUploadFailed, err)
	}

	return &Response{
		Number:  response.RCTACK.RCTNUM,
		Date:    response.RCTACK.DATE,
		Time:    response.RCTACK.TIME,
		Code:    response.RCTACK.ACKCODE,
		Message: response.RCTACK.ACKMSG,
	}, nil
}

func generateReceipt(params ReceiptParams, customer Customer, items []Item, payments []Payment) *models.RCT {
	rctPayments := make([]*models.PAYMENT, len(payments))
	for i, payment := range payments {
		rctPayments[i] = &models.PAYMENT{
			PMTTYPE:   string(payment.Type),
			PMTAMOUNT: fmt.Sprintf("%.2f", payment.Amount),
		}
	}

	RESULTS := ProcessItems(items)
	ITEMS := models.ITEMS{ITEM: RESULTS.ITEMS}
	TOTALS := RESULTS.TOTALS
	VATTOTALS := models.VATTOTALS{VATTOTAL: RESULTS.VATTOTALS}
	PAYMENTS := models.PAYMENTS{PAYMENT: rctPayments}

	RECEIPT := &models.RCT{
		DATE:       params.Date,
		TIME:       params.Time,
		TIN:        params.TIN,
		REGID:      params.RegistrationID,
		EFDSERIAL:  params.EFDSerial,
		CUSTIDTYPE: int64(customer.Type),
		CUSTID:     customer.ID,
		CUSTNAME:   customer.Name,
		MOBILENUM:  customer.Mobile,
		RCTNUM:     params.ReceiptNum,
		DC:         params.DailyCounter,
		GC:         params.GlobalCounter,
		ZNUM:       params.ZNum,
		RCTVNUM:    params.ReceiptVNum,
		ITEMS:      ITEMS,
		TOTALS:     TOTALS,
		PAYMENTS:   PAYMENTS,
		VATTOTALS:  VATTOTALS,
	}

	// round off all values to 2 decimal places
	RECEIPT.RoundOff()

	return RECEIPT
}

func ReceiptBytes(privateKey *rsa.PrivateKey, params ReceiptParams, customer Customer,
	items []Item, payments []Payment,
) ([]byte, error) {
	receipt := generateReceipt(params, customer, items, payments)
	receiptBytes, err := xml.Marshal(receipt)
	if err != nil {
		return nil, fmt.Errorf("could not marshal receipt: %w", err)
	}
	replacer := strings.NewReplacer(
		"<PAYMENT>", "",
		"</PAYMENT>", "",
		"<VATTOTAL>", "",
		"</VATTOTAL>", "")

	receiptBytes = []byte(replacer.Replace(string(receiptBytes)))
	signedReceipt, err := Sign(privateKey, receiptBytes)
	if err != nil {
		return nil, fmt.Errorf("could not sign receipt: %w", err)
	}
	base64SignedReceipt := encodeBase64Bytes(signedReceipt)
	receiptString := string(receiptBytes)

	report := fmt.Sprintf("<EFDMS>%s<EFDMSSIGNATURE>%s</EFDMSSIGNATURE></EFDMS>", receiptString, base64SignedReceipt)
	report = fmt.Sprintf("%s%s", xml.Header, report)

	return []byte(report), nil
}

// ReceiptLink creates a link to the receipt it accepts RECEIPTCODE, GC and the RECEIPTTIME
// and env.Env to know if the receipt was created during testing or production.
func ReceiptLink(e env.Env, receiptCode string, gc int64, receiptTime string) string {
	var baseURL string

	if e == env.PROD {
		baseURL = VerifyReceiptProductionURL
	} else {
		baseURL = VerifyReceiptTestingURL
	}
	return receiptLink(baseURL, receiptCode, gc, receiptTime)
}

func receiptLink(baseURL string, receiptCode string, gc int64, receiptTime string) string {
	return fmt.Sprintf(
		"%s%s%d_%s",
		baseURL,
		receiptCode,
		gc,
		strings.ReplaceAll(receiptTime, ":", ""))
}

type (
	ItemProcessResponse struct {
		ITEMS     []*models.ITEM
		VATTOTALS []*models.VATTOTAL
		TOTALS    models.TOTALS
	}

	vatTotal struct {
		VATRATE    string
		NETTAMOUNT float64
		TAXAMOUNT  float64
	}
)

// ProcessItems processes the []Items in the submitted receipt request
// and create []*models.ITEM which is used to create the xml request also
// calculates the total discount, total tax exclusive and total tax inclusive
func ProcessItems(items []Item) *ItemProcessResponse {
	var (
		DISCOUNT          = 0.0
		TOTALTAXEXCLUSIVE = 0.0
		TOTALTAXINCLUSIVE = 0.0
	)

	// TotalPrice = UnitPrice * Quantity
	// Amount = TotalPrice - Discount
	// TaxableAmount + TaxableAmount * TaxRate = Amount
	vatTotals := make(map[string]*vatTotal)
	var ITEMS []*models.ITEM
	for _, item := range items {
		item := item
		itemAmount := item.Quantity * item.UnitPrice
		itemXML := &models.ITEM{
			ID:      item.ID,
			DESC:    item.Description,
			QTY:     item.Quantity,
			TAXCODE: item.TaxCode,
			AMT:     itemAmount,
		}
		itemAmountWithoutDiscount := itemAmount - item.Discount
		DISCOUNT += item.Discount
		ITEMS = append(ITEMS, itemXML)
		NETAMOUNT := NetAmount(item.TaxCode, itemAmountWithoutDiscount)
		TOTALTAXEXCLUSIVE += NETAMOUNT
		TOTALTAXINCLUSIVE += itemAmountWithoutDiscount
		TAXAMOUNT := ValueAddedTaxAmount(item.TaxCode, itemAmountWithoutDiscount)
		vatID := ParseTaxCode(item.TaxCode).ID
		// check if the tax code is already in the map if not add it
		if _, ok := vatTotals[vatID]; !ok {
			vatTotals[vatID] = &vatTotal{
				VATRATE:    vatID,
				NETTAMOUNT: NETAMOUNT,
				TAXAMOUNT:  TAXAMOUNT,
			}
		} else {
			vatTotals[vatID].NETTAMOUNT += NETAMOUNT
			vatTotals[vatID].TAXAMOUNT += TAXAMOUNT
		}
	}

	VATTOTALS := make([]*models.VATTOTAL, 0)
	for _, v := range vatTotals {
		V := &models.VATTOTAL{
			VATRATE:    v.VATRATE,
			NETTAMOUNT: fmt.Sprintf("%.2f", v.NETTAMOUNT),
			TAXAMOUNT:  fmt.Sprintf("%.2f", v.TAXAMOUNT),
		}
		VATTOTALS = append(VATTOTALS, V)
	}
	TOTALS := models.TOTALS{
		TOTALTAXEXCL: TOTALTAXEXCLUSIVE,
		TOTALTAXINCL: TOTALTAXINCLUSIVE,
		DISCOUNT:     DISCOUNT,
	}
	return &ItemProcessResponse{
		ITEMS:     ITEMS,
		VATTOTALS: VATTOTALS,
		TOTALS:    TOTALS,
	}
}
