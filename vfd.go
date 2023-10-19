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
	"context"
	"crypto/rsa"
	"strings"

	"github.com/Golang-Tanzania/tra-vfd/pkg/env"
)

const (
	RegisterProductionURL                  = "https://vfd.tra.go.tz/api/vfdRegReq"
	FetchTokenProductionURL                = "https://vfd.tra.go.tz/vfdtoken" //nolint:gosec
	SubmitReceiptProductionURL             = "https://vfd.tra.go.tz/api/efdmsRctInfo"
	SubmitReportProductionURL              = "https://vfd.tra.go.tz/api/efdmszreport"
	VerifyReceiptProductionURL             = "https://verify.tra.go.tz/"
	RegisterTestingURL                     = "https://virtual.tra.go.tz/efdmsRctApi/api/vfdRegReq"
	FetchTokenTestingURL                   = "https://virtual.tra.go.tz/efdmsRctApi/vfdtoken" //nolint:gosec
	SubmitReceiptTestingURL                = "https://virtual.tra.go.tz/efdmsRctApi/api/efdmsRctInfo"
	SubmitReportTestingURL                 = "https://virtual.tra.go.tz/efdmsRctApi/api/efdmszreport"
	VerifyReceiptTestingURL                = "https://virtual.tra.go.tz/efdmsRctVerify/"
	RegisterClientAction       Action      = "register"
	FetchTokenAction           Action      = "token"
	SubmitReceiptAction        Action      = "receipt"
	SubmitReportAction         Action      = "report"
	ReceiptVerificationAction  Action      = "verification"
	CashPaymentType            PaymentType = "CASH"
	CreditCardPaymentType      PaymentType = "CCARD"
	ChequePaymentType          PaymentType = "CHEQUE"
	InvoicePaymentType         PaymentType = "INVOICE"
	ElectronicPaymentType      PaymentType = "EMONEY"
	TINCustomerID              CustomerID  = 1
	LicenceCustomerID          CustomerID  = 2
	VoterIDCustomerID          CustomerID  = 3
	PassportCustomerID         CustomerID  = 4
	NIDACustomerID             CustomerID  = 5
	NonCustomerID              CustomerID  = 6
	MeterNumberCustomerID      CustomerID  = 7
	SubmitReceiptRoutingKey    string      = "vfdrct"
	SubmitReportRoutingKey     string      = "vfdzreport"
	ContentTypeXML             string      = "application/xml"
	RegistrationRequestClient  string      = "webapi"
)

type (
	// Action signifies the action to be performed among the four defined actions
	// which are INSTANCE registration, token fetching, submission of receipt and
	// submission of report.
	Action string

	// URL is a struct that holds the URLs for the four actions
	URL struct {
		Registration  string
		FetchToken    string
		SubmitReceipt string
		SubmitReport  string
		VerifyReceipt string
	}

	// PaymentType represent the type of payment that is recognized by the VFD server
	// There are five types of payments: CASH, CHEQUE, CCARD, EMONEY and INVOICE.
	PaymentType string

	// CustomerID is the type of ID the customer used during purchase
	// The Type of ID is to be included in the receipt.
	// Allowed values for CustomerID are 1 through 7. The number to type
	// mapping are as follows:
	// 1: Tax Identification Number (TIN), 2: Driving License, 3: Voters Number,
	// 4: Travel Passport, 5: National ID, 6: NIL (No Identity Used), 7: Meter Number
	CustomerID int

	// RequestHeaders represent collection of request headers during receipt or Z report
	// sending via VFD Service.
	RequestHeaders struct {
		CertSerial  string
		BearerToken string
	}

	Payment struct {
		Type   PaymentType
		Amount float64
	}

	// VATTOTAL represent the VAT details.
	VATTOTAL struct {
		ID        string
		Rate      float64
		TaxAmount float64
		NetAmount float64
	}

	// Response contains details returned when submitting a receipt to the VFD Service
	// or a Z report.
	// Number (int) is the receipt number in case of a receipt submission and the
	// Z report number in case of a Z report submission.
	// Date (string) is the date of the receipt or Z report submission. The format
	// is YYYY-MM-DD.
	// Time (string) is the time of the receipt or Z report submission. The format
	// is HH24:MI:SS
	// Code (int) is the response code. 0 means success.
	// Message (string) is the response message.
	Response struct {
		Number  int64  `json:"number,omitempty"`
		Date    string `json:"date,omitempty"`
		Time    string `json:"time,omitempty"`
		Code    int64  `json:"code,omitempty"`
		Message string `json:"message,omitempty"`
	}

	Service interface {
		// Register is used to register a virtual fiscal device (VFD) with the VFD Service.
		// If successful, the VFD Service returns a registration response containing the
		// VFD details and the credentials to use when submitting receipts and Z reports.
		// Registering a VFD is a one-time operation. The subsequent calls to Register will
		// yield the same response.VFD should store the registration response to
		// avoid calling Register again.
		Register(ctx context.Context, url string, privateKey *rsa.PrivateKey, request *RegistrationRequest,
		) (*RegistrationResponse, error)

		// FetchToken is used to fetch a token from the VFD Service. The token is used
		// to authenticate the VFD when submitting receipts and Z reports.
		// credentials used here are the ones returned by the Register method.
		FetchToken(ctx context.Context, url string, request *TokenRequest) (*TokenResponse, error)

		// SubmitReceipt is used to submit a receipt to the VFD Service. The receipt
		// is signed using the private key. The private key is obtained from the certificate
		// issued by the Revenue Authority during integration.
		SubmitReceipt(
			ctx context.Context, url string, headers *RequestHeaders,
			privateKey *rsa.PrivateKey, receipt *ReceiptRequest) (*Response, error)

		// SubmitReport is used to submit a Z report to the VFD Service. The Z report
		// is signed using the private key. The private key is obtained from the certificate
		// issued by the Revenue Authority during integration.
		SubmitReport(
			ctx context.Context, url string, headers *RequestHeaders,
			privateKey *rsa.PrivateKey, report *ReportRequest) (*Response, error)
	}
)

// IsSuccess checks the response ack code and return true if the code
// means success and false if otherwise
func IsSuccess(code int64) bool {
	return code == SuccessCode
}

// ParsePayment ...
func ParsePayment(value any) PaymentType {
	// heck if int or string
	switch i := value.(type) {
	case int, int64:
		valueInt := i.(int)
		switch valueInt {
		case 1:
			return CashPaymentType
		case 2:
			return ChequePaymentType
		case 3:
			return CreditCardPaymentType
		case 4:
			return ElectronicPaymentType
		case 5:
			return InvoicePaymentType
		default:
			return CashPaymentType
		}

	case string:
		valueString := strings.ToUpper(i)
		switch valueString {
		case "CASH":
			return CashPaymentType
		case "CHEQUE":
			return ChequePaymentType
		case "CCARD":
			return CreditCardPaymentType
		case "EMONEY":
			return ElectronicPaymentType
		case "INVOICE":
			return InvoicePaymentType
		default:
			return CashPaymentType
		}

	default:
		return CashPaymentType
	}
}

// RequestURL returns the URL for the specified Action and specified env.Env
// returns empty string if either action or env is not recognized
func RequestURL(e env.Env, action Action) string {
	var u *URL
	if e == env.PROD {
		u = &URL{
			Registration:  RegisterProductionURL,
			FetchToken:    FetchTokenProductionURL,
			SubmitReceipt: SubmitReceiptProductionURL,
			SubmitReport:  SubmitReportProductionURL,
			VerifyReceipt: VerifyReceiptProductionURL,
		}
	} else {
		u = &URL{
			Registration:  RegisterTestingURL,
			FetchToken:    FetchTokenTestingURL,
			SubmitReceipt: SubmitReceiptTestingURL,
			SubmitReport:  SubmitReportTestingURL,
			VerifyReceipt: VerifyReceiptTestingURL,
		}
	}

	switch action {
	case RegisterClientAction:
		return u.Registration
	case FetchTokenAction:
		return u.FetchToken
	case SubmitReceiptAction:
		return u.SubmitReceipt
	case SubmitReportAction:
		return u.SubmitReport
	case ReceiptVerificationAction:
		return u.VerifyReceipt
	default:
		return ""
	}
}
