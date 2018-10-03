package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http"
	"time"
)

// GetTLSConfig generate tls.Config from CAFile
func GetTLSConfig(CAFile string) (*tls.Config, error) {
	caCert, err := iotul.ReadFile(CAFile)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		return nil, errors.New("Failed to add CA to pool")
	}

	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	tlsConfig.BuildNameToCertificate()
	return tlsConfig, nil
}

//CreateHTTPClient returns a http.Client
func CreateHTTPClient(CAFile string) (*http.Client, error) {
	var tlsConfig *tls.Config
	var err error

	if CAFile != "" {
		tlsConfig, err = GetTLSConfig(CAFile)
		if err != nil {
			return nil, err
		}
	}

	tr := &http.Transport{
		MaxIdleConnsPerHost: 20,
		TLSClientConfig:     tlsConfig,
	}

	return &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}, nil
}
