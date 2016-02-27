package ec2

import "errors"

type KeyPairManager struct {
	validator keypairValidator
	generator keypairGenerator
	uploader  keypairUploader
	retriever keypairRetriever
	verifier  keypairVerifier
}

type keypairValidator interface {
	Validate(pemData []byte) error
}

type keypairGenerator interface {
	Generate() (KeyPair, error)
}

type keypairUploader interface {
	Upload(Session, KeyPair) error
}

type keypairRetriever interface {
	Retrieve(session Session, keypairName string) (KeyPairInfo, bool, error)
}

type keypairVerifier interface {
	Verify(fingerprint string, pemData []byte) error
}

func NewKeyPairManager(validator keypairValidator, generator keypairGenerator, uploader keypairUploader, retriever keypairRetriever, verifier keypairVerifier) KeyPairManager {
	return KeyPairManager{
		validator: validator,
		generator: generator,
		uploader:  uploader,
		retriever: retriever,
		verifier:  verifier,
	}
}

func (m KeyPairManager) Sync(ec2Session Session, keypair KeyPair) (KeyPair, error) {
	hasLocalKeyPair := !keypair.IsEmpty()
	_, hasRemoteKeyPair, err := m.retriever.Retrieve(ec2Session, keypair.Name)
	if err != nil {
		return KeyPair{}, err
	}

	if hasLocalKeyPair {
		err = m.validator.Validate(keypair.PrivateKey)
		if err != nil {
			return KeyPair{}, err
		}
	}

	switch {
	case !hasLocalKeyPair:
		keypair, err = m.generator.Generate()
		if err != nil {
			return KeyPair{}, err
		}

		err = m.uploader.Upload(ec2Session, keypair)
		if err != nil {
			return KeyPair{}, err
		}
	case hasLocalKeyPair && !hasRemoteKeyPair:
		err = m.uploader.Upload(ec2Session, keypair)
		if err != nil {
			return KeyPair{}, err
		}
	}

	keypairInfo, hasRemoteKeyPair, err := m.retriever.Retrieve(ec2Session, keypair.Name)
	if err != nil {
		return KeyPair{}, err
	}
	if !hasRemoteKeyPair {
		return KeyPair{}, errors.New("could not retrieve keypair for verification")
	}

	err = m.verifier.Verify(keypairInfo.Fingerprint, keypair.PrivateKey)
	if err != nil {
		return KeyPair{}, err
	}

	return keypair, nil
}
