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
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Golang-Tanzania/tra-vfd/pkg/env"

	vhttp "github.com/Golang-Tanzania/tra-vfd/internal/http"
	"github.com/Golang-Tanzania/tra-vfd/internal/models"
)

type (
	// RawRequest contains information needed to send receipt/z report file
	// to the vfd server.
	RawRequest struct {
		Env      env.Env
		Action   Action
		FilePath string
	}
)

// SubmitRawRequest is useful for submitting requests that are in form of XML files
// content of the file is read and submitted to the server as is.
func SubmitRawRequest(ctx context.Context, headers *RequestHeaders,
	raw *RawRequest) (*Response, error) {
	var (
		client      = vhttp.Instance()
		certSerial  = headers.CertSerial
		bearerToken = headers.BearerToken
		reqURL      = RequestURL(raw.Env, raw.Action)
	)

	payload := bytes.NewBuffer(nil)

	// read the file if the file path is provided and return the content as bytes
	if raw.FilePath != "" {
		file, err := os.Open(raw.FilePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		if _, err := io.Copy(payload, file); err != nil {
			return nil, err
		}
	}

	newContext, cancel := context.WithCancel(ctx)
	defer cancel()

	req, err := http.NewRequestWithContext(newContext, http.MethodPost, reqURL, payload)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", ContentTypeXML)
	req.Header.Set("Cert-Serial", encodeBase64String(certSerial))
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", bearerToken))

	if raw.Action == SubmitReceiptAction {
		req.Header.Set("Routing-Key", SubmitReceiptRoutingKey)
	}

	if raw.Action == SubmitReportAction {
		req.Header.Set("Routing-Key", SubmitReportRoutingKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, checkNetworkError(newContext, "raw request submit", err)
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

	if raw.Action == SubmitReportAction {
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

	if raw.Action == SubmitReceiptAction {
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

	return nil, fmt.Errorf("couldnt figure out the action")
}
