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
	"errors"
	"fmt"
	"net"
)

// ACKCODE	STATUS	DESCRIPTION	POSSIBLE REASON
// 0	    SUCCESS	Success
// 1	FAIL	Invalid Signature	Signature generated not in correct format. Signature generated with missing nodes, signature generated with empty lines in XML or
// 3	FAIL	Invalid TIN	TIN specified with dash or wrong TIN specified
// 4	FAIL	VFD Registration Approval required	Request posted without Client details, which is WEBAPI.
// 5	FAIL	Unhandled Exception	Contact TRA for further troubleshooting
// 6		Invalid Serial or Serial not Registered to Web API/TIN	CERTKEY is not registered to TIN sending registration request. Use only TIN and CERTKEY provided by TRA
// 7	FAIL	Invalid client header	Wrong client value specified
// 8	FAIL	Wrong Certificate used to Register Web API	Wrong certificate used

const (
	SuccessCode          int64 = 0
	InvalidSignatureCode int64 = 1
	InvalidTaxID         int64 = 3
	ApprovalRequired     int64 = 4
	UnhandledException   int64 = 5
	InvalidSerial        int64 = 6
	InvalidClientHeader  int64 = 7
	InvalidCertificate   int64 = 8
)

type (
	Error struct {
		Code    int64  `json:"code,omitempty"`
		Message string `json:"message,omitempty"`
	}

	// NetworkError is returned when there is an error in the network.
	NetworkError struct {
		Err     error
		Message string
	}
)

func (e *NetworkError) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
}

// Unwrap returns the underlying error.
func (e *NetworkError) Unwrap() error {
	return e.Err
}

// ParseErrorCode parses the error code and returns the corresponding error message.
func ParseErrorCode(code int64) string {
	switch code {
	case 0:
		return "SUCCESS"
	case 1:
		return "FAIL"
	case 3:
		return "Invalid TIN"
	case 4:
		return "VFD Registration Approval required"
	case 5:
		return "Unhandled Exception"
	case 6:
		return "Invalid Serial or Serial not Registered to Web API/TIN"
	case 7:
		return "Invalid client header"
	case 8:
		return "Wrong Certificate used to Register Web API"
	default:
		return "Unknown error"
	}
}

// IsNetworkError returns true if the error is a NetworkError.
func IsNetworkError(err error) bool {
	netErr := &NetworkError{}
	return errors.As(err, &netErr)
}

// checkNetworkError checks if the errors is about network and returns a NetworkError.
// An errors is considered a network error if it is a net.Error, context.Canceled
// or context.DeadlineExceeded. or just a context error.
func checkNetworkError(ctx context.Context, prefix string, err error) error {
	if err == nil {
		return nil
	}

	if ctx.Err() != nil || err == context.Canceled || err == context.DeadlineExceeded {
		return &NetworkError{
			Err:     ctx.Err(),
			Message: fmt.Sprintf("%s: context error (canceled or deadline exceeded)", prefix),
		}
	}

	if netErr, ok := err.(net.Error); ok {
		return &NetworkError{
			Err:     netErr,
			Message: fmt.Sprintf("%s: network error", prefix),
		}
	}

	return fmt.Errorf("%s: %w", prefix, err)
}
