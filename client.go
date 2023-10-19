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
	"net/http"
)

type (
	Client struct {
		http *http.Client
	}

	Option func(*Client)
)

func WithHttpClient(http *http.Client) Option {
	return func(c *Client) {
		c.http = http
	}
}

// SetHttpClient sets the http client
func (c *Client) SetHttpClient(http *http.Client) {
	if http != nil {
		c.http = http
	}
}

func NewClient(options ...Option) *Client {
	client := &Client{
		http: http.DefaultClient,
	}
	for _, option := range options {
		option(client)
	}
	return client
}

func (c *Client) Register(ctx context.Context,
	url string, privateKey *rsa.PrivateKey,
	request *RegistrationRequest,
) (*RegistrationResponse, error) {
	response, err := register(ctx, c.http, url, privateKey, request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) FetchToken(ctx context.Context, url string,
	request *TokenRequest,
) (*TokenResponse, error) {
	return fetchToken(ctx, c.http, url, request)
}

func (c *Client) FetchTokenWithMw(ctx context.Context, url string,
	request *TokenRequest, callback OnTokenResponse,
) (*TokenResponse, error) {
	response, err := fetchToken(ctx, c.http, url, request)
	if err != nil {
		return nil, err
	}

	err = callback(ctx, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) SubmitReceipt(
	ctx context.Context,
	url string,
	headers *RequestHeaders,
	privateKey *rsa.PrivateKey,
	receipt *ReceiptRequest,
) (*Response, error) {
	return submitReceipt(ctx, c.http, url, headers, privateKey, receipt)
}

func (c *Client) SubmitReport(
	ctx context.Context,
	url string,
	headers *RequestHeaders,
	privateKey *rsa.PrivateKey,
	report *ReportRequest,
) (*Response, error) {
	return submitReport(ctx, c.http, url, headers, privateKey, report)
}
