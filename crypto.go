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
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1" //nolint:gosec
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"os"

	"software.sslmate.com/src/go-pkcs12"
)

type (
	// CertLoader loads a certificate from a file and returns the private key and the certificate
	CertLoader func(certPath string, certPassword string) (*rsa.PrivateKey, *x509.Certificate, error)

	// SignatureVerifier verifies the signature of a payload using the public key
	// of the signing certificate
	SignatureVerifier func(publicKey *rsa.PublicKey, payload []byte, signature string) error

	// PayloadSigner signs a payload using the private key of the signing certificate
	// all requests to the VFD API must be signed.
	PayloadSigner func(privateKey *rsa.PrivateKey, payload []byte) ([]byte, error)
)

func LoadCertChain(certPath string, certPassword string) (*rsa.PrivateKey, *x509.Certificate, []*x509.Certificate, error) {
	pfxData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not read the certificate file: %w", err)
	}
	pfx, cert, caCerts, err := pkcs12.DecodeChain(pfxData, certPassword)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not decode the certificate file: %w", err)
	}

	// type check to make sure we have a private key
	privateKey, ok := pfx.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, nil, fmt.Errorf("private key is not of type *rsa.PrivateKey: %w", err)
	}

	return privateKey, cert, caCerts, nil
}

func LoadCert(path, password string) (*rsa.PrivateKey, *x509.Certificate, error) {
	pfxData, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	pfx, cert, err := pkcs12.Decode(pfxData, password)
	if err != nil {
		if err.Error() == "pkcs12: expected exactly two safe bags in the PFX PDU" {
			privateKey, cert, _, err := LoadCertChain(path, password)
			if err != nil {
				return nil, nil, err
			}
			return privateKey, cert, nil
		}
		return nil, nil, err
	}

	// type check to make sure we have a private key
	privateKey, ok := pfx.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("private key is not of type rsa.PrivateKey")
	}

	return privateKey, cert, nil
}

func Sign(privateKey *rsa.PrivateKey, payload []byte) ([]byte, error) {
	signature, err := signPayload(privateKey, payload)
	if err != nil {
		return nil, fmt.Errorf("unable to sign the payload: %w", err)
	}

	hash := sha1.Sum(payload) //nolint:gosec
	err = verifySignature(&privateKey.PublicKey, hash[:], signature)
	if err != nil {
		return nil, fmt.Errorf("could not verify signature %w", err)
	}

	return signature, nil
}

func VerifySignature(publicKey *rsa.PublicKey, payload []byte, signature string) error {
	sg, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("could not verify signature %w", err)
	}

	hash := sha1.Sum(payload) //nolint:gosec
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA1, hash[:], sg)
	if err != nil {
		return fmt.Errorf("could not verify signature %w", err)
	}

	return nil
}

func SignPayload(privateKey *rsa.PrivateKey, payload []byte) ([]byte, error) {
	out, err := signPayload(privateKey, payload)
	if err != nil {
		return nil, fmt.Errorf("unable to sign the payload: %w", err)
	}

	err = VerifySignature(&privateKey.PublicKey, payload, base64.StdEncoding.EncodeToString(out))
	if err != nil {
		return nil, fmt.Errorf("invalid signature %w", err)
	}

	return out, nil
}

func signPayload(pub *rsa.PrivateKey, payload []byte) ([]byte, error) {
	hasher := crypto.SHA1.New()
	hasher.Write(payload)

	out, err := rsa.SignPKCS1v15(rand.Reader, pub, crypto.SHA1, hasher.Sum(nil))
	if err != nil {
		return nil, err
	}

	return out, nil
}

func verifySignature(pub *rsa.PublicKey, hash []byte, sig []byte) error {
	return rsa.VerifyPKCS1v15(pub, crypto.SHA1, hash, sig)
}
