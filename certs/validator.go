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

	"github.com/cloudfoundry/multierror"
)

type CertData struct {
	Cert  []byte
	Key   []byte
	Chain []byte
}

type Validator struct{}

func NewValidator() Validator {
	return Validator{}
}

func (v Validator) ReadAndValidate(certPath, keyPath, chainPath string) (CertData, error) {
	certData, readErrors := v.Read(certPath, keyPath, chainPath)
	if readErrors != nil {
		return CertData{}, readErrors
	}

	validateErrors := v.Validate(certData.Cert, certData.Key, certData.Chain)
	if validateErrors != nil {
		return CertData{}, validateErrors
	}

	return certData, nil
}

func (v Validator) Read(certPath, keyPath, chainPath string) (CertData, error) {
	var err error
	var certBytes []byte
	var keyBytes []byte
	var chainBytes []byte
	validateErrors := multierror.NewMultiError("")

	if certBytes, err = readFile("certificate", "--cert", certPath); err != nil {
		validateErrors.Add(err)
	}

	if keyBytes, err = readFile("key", "--key", keyPath); err != nil {
		validateErrors.Add(err)
	}

	if chainPath != "" {
		if chainBytes, err = readFile("chain", "--chain", chainPath); err != nil {
			validateErrors.Add(err)
		}
	}

	if validateErrors.Length() > 0 {
		return CertData{}, validateErrors
	}

	return CertData{
		Cert:  certBytes,
		Key:   keyBytes,
		Chain: chainBytes,
	}, nil
}

func (v Validator) Validate(cert, key, chain []byte) error {
	var err error
	var certificate *x509.Certificate
	var privateKey *rsa.PrivateKey
	validateErrors := multierror.NewMultiError("")

	err = validatePEM(cert)
	if err != nil {
		validateErrors.Add(fmt.Errorf("certificate %s: \"%s\"", err, cert))
	}
	err = validatePEM(key)
	if err != nil {
		validateErrors.Add(fmt.Errorf("key %s: \"%s\"", err, key))
	}
	if len(chain) > 0 {
		err = validatePEM(chain)
		if err != nil {
			validateErrors.Add(fmt.Errorf("chain %s: \"%s\"", err, chain))
		}
	}

	if validateErrors.Length() > 0 {
		return validateErrors
	}

	tlsCertificateStruct, err := tls.X509KeyPair(cert, key)
	if err != nil {
		validateErrors.Add(err)
	} else {
		privateKey = tlsCertificateStruct.PrivateKey.(*rsa.PrivateKey)
	}
	if certificate == nil {
		loadKeyPairError := err
		certificate, err = parseCertificate(cert, loadKeyPairError)
		if err != nil {
			validateErrors.Add(err)
		}
	}

	var certPool *x509.CertPool
	if len(chain) > 0 {
		certPool, err = parseChain(chain)
		if err != nil {
			validateErrors.Add(err)
		}
	}

	if privateKey != nil && certificate != nil {
		if err := validateCertAndKey(certificate, privateKey); err != nil {
			validateErrors.Add(err)
		}
	}

	if certPool != nil && certificate != nil {
		if err := validateCertAndChain(certificate, certPool); err != nil {
			validateErrors.Add(err)
		}
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

func validateCertAndChain(certificate *x509.Certificate, certPool *x509.CertPool) error {
	opts := x509.VerifyOptions{
		Roots: certPool,
	}

	if _, err := certificate.Verify(opts); err != nil {
		return fmt.Errorf("certificate and chain mismatch: %s", err.Error())
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

func parseChain(chainData []byte) (*x509.CertPool, error) {
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(chainData)
	if !ok {
		return nil, fmt.Errorf("failed to parse chain")
	}

	return roots, nil
}
