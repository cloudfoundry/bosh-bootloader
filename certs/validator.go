package certs

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/crypto/pkcs12"

	"github.com/cloudfoundry/multierror"
)

type CertData struct {
	Cert []byte
	Key  []byte
}

type Validator struct{}

func NewValidator() Validator {
	return Validator{}
}

func (v Validator) ReadAndValidate(certPath, keyPath string) (CertData, error) {
	certData, readErrors := v.Read(certPath, keyPath)
	if readErrors != nil {
		return CertData{}, readErrors
	}

	validateErrors := v.Validate(certData.Cert, certData.Key)
	if validateErrors != nil {
		return CertData{}, validateErrors
	}

	return certData, nil
}

func (v Validator) Read(certPath, keyPath string) (CertData, error) {
	validateErrors := multierror.NewMultiError("")

	certBytes, err := readFile("certificate", "--lb-cert", certPath)
	if err != nil {
		validateErrors.Add(err)
	}

	keyBytes, err := readFile("key", "--lb-key", keyPath)
	if err != nil {
		validateErrors.Add(err)
	}

	if validateErrors.Length() > 0 {
		return CertData{}, validateErrors
	}

	return CertData{
		Cert: certBytes,
		Key:  keyBytes,
	}, nil
}

func (v Validator) Validate(cert, key []byte) error {
	validateErrors := multierror.NewMultiError("")

	err := validatePEM(cert)
	if err != nil {
		validateErrors.Add(fmt.Errorf("certificate %s: \"%s\"", err, cert))
	}

	err = validatePEM(key)
	if err != nil {
		validateErrors.Add(fmt.Errorf("key %s: \"%s\"", err, key))
	}

	if validateErrors.Length() > 0 {
		return validateErrors
	}

	var privateKey *rsa.PrivateKey
	tlsCertificateStruct, err := tls.X509KeyPair(cert, key)
	if err != nil {
		validateErrors.Add(err)
	} else {
		privateKey = tlsCertificateStruct.PrivateKey.(*rsa.PrivateKey)
	}

	loadKeyPairError := err
	certificate, err := parseCertificate(cert, loadKeyPairError)
	if err != nil {
		validateErrors.Add(err)
	}

	if privateKey != nil && certificate != nil {
		if err := validateCertAndKey(certificate, privateKey); err != nil {
			validateErrors.Add(err)
		}
	}

	if validateErrors.Length() > 0 {
		return validateErrors
	}

	return nil
}

func (v Validator) ReadAndValidatePKCS12(certPath, passwordPath string) (CertData, error) {
	certData, readErrors := v.ReadPKCS12(certPath, passwordPath)
	if readErrors != nil {
		return CertData{}, readErrors
	}

	validateErrors := v.ValidatePKCS12(certData.Cert, certData.Key)
	if validateErrors != nil {
		return CertData{}, validateErrors
	}

	return certData, nil
}

func (v Validator) ReadPKCS12(certPath, passwordPath string) (CertData, error) {
	validateErrors := multierror.NewMultiError("")

	certBytes, err := readFile("certificate", "--lb-cert", certPath)
	if err != nil {
		validateErrors.Add(err)
	}

	passwordBytes, err := readFile("key", "--lb-key", passwordPath)
	if err != nil {
		validateErrors.Add(err)
	}

	if validateErrors.Length() > 0 {
		return CertData{}, validateErrors
	}

	passwordString := strings.TrimSuffix(string(passwordBytes), "\n")

	return CertData{
		Cert: certBytes,
		Key:  []byte(passwordString),
	}, nil
}

func (v Validator) ValidatePKCS12(cert, password []byte) error {
	validateErrors := multierror.NewMultiError("")

	_, err := pkcs12.ToPEM(cert, string(password))
	if err != nil {
		validateErrors.Add(fmt.Errorf("failed to parse certificate: %s", err))
	}

	if validateErrors.Length() > 0 {
		return validateErrors
	}
	return nil
}

func validatePEM(data []byte) error {
	p, _ := pem.Decode(data)
	if p == nil {
		return errors.New("is not PEM encoded")
	}
	return nil
}

func readFile(propertyName string, flagName string, filePath string) ([]byte, error) {
	if filePath == "" {
		return []byte{}, fmt.Errorf("%s is required", flagName)
	}

	file, err := os.Open(filePath)
	if os.IsNotExist(err) {
		return []byte{}, fmt.Errorf(`%s file not found: %q`, propertyName, filePath)
	} else if err != nil {
		return []byte{}, err
	}

	fileInfo, err := os.Stat(file.Name())
	if err != nil {
		return []byte{}, fmt.Errorf("%s: %s", err, filePath)
	}

	if !fileInfo.Mode().IsRegular() {
		return []byte{}, fmt.Errorf(`%s is not a regular file: %q`, propertyName, filePath)
	}

	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		return []byte{}, fmt.Errorf("%s: %s", err, filePath)
	}

	return fileData, nil
}

func validateCertAndKey(certificate *x509.Certificate, privateKey *rsa.PrivateKey) error {
	publicKey := certificate.PublicKey.(*rsa.PublicKey)
	if privateKey.PublicKey.N.Cmp(publicKey.N) != 0 || privateKey.PublicKey.E != publicKey.E {
		return errors.New("certificate and key mismatch")
	}

	return nil
}

func parseCertificate(certificateData []byte, loadKeyPairError error) (*x509.Certificate, error) {
	pemCertData, _ := pem.Decode(certificateData)
	cert, err := x509.ParseCertificate(pemCertData.Bytes)
	if err != nil && err != loadKeyPairError {
		return nil, fmt.Errorf("failed to parse certificate: %s", err)
	}

	return cert, nil
}
