package iam

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/multierror"
)

var readAll func(r io.Reader) ([]byte, error) = ioutil.ReadAll
var stat func(name string) (os.FileInfo, error) = os.Stat

type CertificateValidator struct{}

func NewCertificateValidator() CertificateValidator {
	return CertificateValidator{}
}

func (c CertificateValidator) Validate(command, certPath, keyPath, chainPath string) error {
	var err error
	var certificateData []byte
	var chainData []byte
	var certificate *x509.Certificate
	var privateKey *rsa.PrivateKey

	validateErrors := multierror.NewMultiError(command)

	if certificateData, err = c.validateFileAndFormat("certificate", "--cert", certPath); err != nil {
		validateErrors.Add(err)
	}

	if _, err = c.validateFileAndFormat("key", "--key", keyPath); err != nil {
		validateErrors.Add(err)
	}

	if chainPath != "" {
		if chainData, err = c.validateFileAndFormat("chain", "--chain", chainPath); err != nil {
			validateErrors.Add(err)
		}
	}

	if validateErrors.Length() > 0 {
		return validateErrors
	}

	tlsCertificateStruct, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		validateErrors.Add(err)
	} else {
		privateKey = tlsCertificateStruct.PrivateKey.(*rsa.PrivateKey)
	}
	if certificate == nil {
		loadKeyPairError := err
		certificate, err = c.parseCertificate(certificateData, loadKeyPairError)
		if err != nil {
			validateErrors.Add(err)
		}
	}

	var certPool *x509.CertPool
	if chainPath != "" {
		certPool, err = c.parseChain(chainData)
		if err != nil {
			validateErrors.Add(err)
		}
	}

	if privateKey != nil && certificate != nil {
		if err := c.validateCertAndKey(certificate, privateKey); err != nil {
			validateErrors.Add(err)
		}
	}

	if certPool != nil && certificate != nil {
		if err := c.validateCertAndChain(certificate, certPool); err != nil {
			validateErrors.Add(err)
		}
	}

	if validateErrors.Length() > 0 {
		return validateErrors
	}

	return nil
}

func (CertificateValidator) validateFileAndFormat(propertyName string, flagName string, filePath string) ([]byte, error) {
	if filePath == "" {
		return []byte{}, fmt.Errorf("%s is required", flagName)
	}

	file, err := os.Open(filePath)
	if os.IsNotExist(err) {
		return []byte{}, fmt.Errorf(`%s file not found: %q`, propertyName, filePath)
	} else if err != nil {
		return []byte{}, err
	}

	fileInfo, err := stat(file.Name())
	if err != nil {
		return []byte{}, fmt.Errorf("%s: %s", err, filePath)
	}

	if !fileInfo.Mode().IsRegular() {
		return []byte{}, fmt.Errorf(`%s is not a regular file: %q`, propertyName, filePath)
	}

	fileData, err := readAll(file)
	if err != nil {
		return []byte{}, fmt.Errorf("%s: %s", err, filePath)
	}

	p, _ := pem.Decode(fileData)
	if p == nil {
		return []byte{}, fmt.Errorf("%s is not PEM encoded: %q", propertyName, filePath)
	}

	return fileData, nil
}

func (c CertificateValidator) validateCertAndKey(certificate *x509.Certificate, privateKey *rsa.PrivateKey) error {
	publicKey := certificate.PublicKey.(*rsa.PublicKey)
	if privateKey.PublicKey.N.Cmp(publicKey.N) != 0 || privateKey.PublicKey.E != publicKey.E {
		return errors.New("certificate and key mismatch")
	}

	return nil
}

func (CertificateValidator) validateCertAndChain(certificate *x509.Certificate, certPool *x509.CertPool) error {
	opts := x509.VerifyOptions{
		Roots: certPool,
	}

	if _, err := certificate.Verify(opts); err != nil {
		return fmt.Errorf("certificate and chain mismatch: %s", err.Error())
	}

	return nil
}

func (CertificateValidator) parseCertificate(certificateData []byte, loadKeyPairError error) (*x509.Certificate, error) {
	pemCertData, _ := pem.Decode(certificateData)
	cert, err := x509.ParseCertificate(pemCertData.Bytes)
	if err != nil && err != loadKeyPairError {
		return nil, fmt.Errorf("failed to parse certificate: %s", err)
	}

	return cert, nil
}

func (CertificateValidator) parseChain(chainData []byte) (*x509.CertPool, error) {
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(chainData)
	if !ok {
		return nil, fmt.Errorf("failed to parse chain")
	}

	return roots, nil
}
