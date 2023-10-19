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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	xhttp "github.com/Golang-Tanzania/tra-vfd/internal/http"
)

// ErrFetchToken is the error returned when the token request fails.
// It is a wrapper for the underlying error.
var ErrFetchToken = errors.New("fetch token failed")

type (
	// TokenRequest contains the request parameters needed to get a token.
	// GrantType - The type of the grant_type.
	// Username - The username of the user.
	// Password - The password of the user.
	TokenRequest struct {
		Username  string
		Password  string
		GrantType string
	}

	// TokenResponse contains the response parameters returned by the token endpoint.
	TokenResponse struct {
		Code        string `json:"code,omitempty"`
		Message     string `json:"message,omitempty"`
		AccessToken string `json:"access_token,omitempty"`
		TokenType   string `json:"token_type,omitempty"`
		ExpiresIn   int64  `json:"expires_in,omitempty"`
		Error       string `json:"error,omitempty"`
	}

	// FetchTokenFunc is a function that fetches a token from the VFD server.
	FetchTokenFunc func(ctx context.Context, url string, request *TokenRequest) (*TokenResponse, error)

	// OnTokenResponse is a callback function that is called when a token is received.
	OnTokenResponse func(context.Context, *TokenResponse) error

	// TokenResponseMiddleware is a middleware function that is called when a token is received.
	TokenResponseMiddleware func(next OnTokenResponse) OnTokenResponse
)

// WrapTokenResponseMiddleware wraps a TokenResponseMiddleware with a OnTokenResponse.
func WrapTokenResponseMiddleware(next OnTokenResponse, middlewares ...TokenResponseMiddleware) OnTokenResponse {
	// loop backwards through the middlewares
	for i := len(middlewares) - 1; i >= 0; i-- {
		next = middlewares[i](next)
	}
	return next
}

// FetchTokenWithMw retrieves a token from the VFD server then passes it to the callback function
// This is beacuse the response might have a code and message that needs to be handled.
func FetchTokenWithMw(ctx context.Context, url string, request *TokenRequest, callback OnTokenResponse) (*TokenResponse, error) {
	httpClient := xhttp.Instance()

	response, err := fetchToken(ctx, httpClient, url, request)
	if err != nil {
		return nil, err
	}

	err = callback(ctx, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// FetchToken retrieves a token from the VFD server. If the status code is not 200, an error is
// returned. Error Message will contain TokenResponse.Code and TokenResponse.Message
// FetchToken wraps internally a *http.Client responsible for making http calls. It has a timeout
// of 70 seconds. It is advised to call this only when the previous token has expired. It will still
// work if called before the token expires.
func FetchToken(ctx context.Context, url string, request *TokenRequest) (*TokenResponse, error) {
	httpClient := xhttp.Instance()
	return fetchToken(ctx, httpClient, url, request)
}

// fetchToken retrieves a token from the VFD server. If the status code is not 200, an error is returned.
// It is a context-aware function with a timeout of 1 minute
func fetchToken(ctx2 context.Context, client *http.Client, path string, request *TokenRequest) (*TokenResponse, error) {
	var (
		username  = request.Username
		password  = request.Password
		grantType = request.GrantType
	)

	// this request should have a max of 1 Minute timeout
	ctx2, cancel := context.WithTimeout(ctx2, 1*time.Minute)
	defer cancel()
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)
	form.Set("grant_type", grantType)
	buffer := bytes.NewBufferString(form.Encode())
	req, err := http.NewRequestWithContext(ctx2, http.MethodPost, path, buffer)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", ErrFetchToken, err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return nil, checkNetworkError(ctx2, "fetch token", err)
	}
	defer resp.Body.Close()

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", ErrFetchToken, err)
	}

	response := new(TokenResponse)

	if err := json.NewDecoder(bytes.NewBuffer(out)).Decode(response); err != nil {
		return nil, fmt.Errorf("response decode error: %w", err)
	}

	response.Code = resp.Header.Get("ACKCODE")
	response.Message = resp.Header.Get("ACKMSG")

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: error code=[%s],message=[%s], error=[%s]",
			ErrFetchToken, response.Code, response.Message, response.Error)
	}

	return response, nil
}

func (tr *TokenResponse) String() string {
	return fmt.Sprintf(
		"FetchToken Response: [Code=%s,Message=%s,AccessToken=%s,TokenType=%s,ExpiresIn=%d seconds,Error=%s]",
		tr.Code, tr.Message, tr.AccessToken, tr.TokenType, tr.ExpiresIn, tr.Error)
}
