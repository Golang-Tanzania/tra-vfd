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
	"testing"

	"github.com/Golang-Tanzania/tra-vfd/pkg/env"
)

func TestParsePayment(t *testing.T) {
	t.Parallel()
	type test struct {
		name  string
		value any
		want  PaymentType
	}

	tests := []test{
		{
			name:  "TestParsePayment",
			value: 1,
			want:  CashPaymentType,
		},
		{
			name:  "TestParsePayment",
			value: "cash",
			want:  CashPaymentType,
		},
		{
			name:  "TestParsePayment",
			value: "CHEQUE",
			want:  ChequePaymentType,
		},
		{
			name:  "TestParsePayment",
			value: 3,
			want:  CreditCardPaymentType,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ParsePayment(tt.value); got != tt.want {
				t.Errorf("ParsePayment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequestURL(t *testing.T) {
	t.Parallel()
	type args struct {
		e      env.Env
		action Action
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "TestReceiptSubmitURLSTAGING",
			args: args{
				e:      env.TEST,
				action: SubmitReceiptAction,
			},
			want: SubmitReceiptTestingURL,
		},
		{
			name: "TestReceiptSubmitURLPRODUCTION",
			args: args{
				e:      env.PROD,
				action: SubmitReceiptAction,
			},
			want: SubmitReceiptProductionURL,
		},
		{
			name: "TestFetchTokenURLSTAGING",
			args: args{
				e:      env.TEST,
				action: FetchTokenAction,
			},
			want: FetchTokenTestingURL,
		},
		{
			name: "TestFetchTokenURLPRODUCTION",
			args: args{
				e:      env.PROD,
				action: FetchTokenAction,
			},
			want: FetchTokenProductionURL,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := RequestURL(tt.args.e, tt.args.action); got != tt.want {
				t.Errorf("RequestURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
