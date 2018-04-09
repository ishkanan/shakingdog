package webserver

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
)

func NewServer(addr, caPath, roURL string, certPaths, keyPaths []string) (*http.Server, net.Listener, error) {
	config := &tls.Config{
		PreferServerCipherSuites: true,
		SessionTicketsDisabled:   true,
	}
	for i, certPath := range certPaths {
		cert, err := tls.LoadX509KeyPair(certPath, keyPaths[i])
		if err != nil {
			return nil, nil, fmt.Errorf("Error loading certificate (%s, %s): %s", certPath, keyPaths[i], err)
		}
		config.Certificates = append(config.Certificates, cert)
	}
	config.BuildNameToCertificate()

	// If a caPath has been specified then a local CA is being used
	// and not the system configuration.

	if caPath != "" {
		pemCert, err := ioutil.ReadFile(caPath)
		if err != nil {
			return nil, nil, fmt.Errorf("Error reading %s: %s\n", caPath, err)
		}

		derCert, _ := pem.Decode(pemCert)
		if derCert == nil {
			return nil, nil, fmt.Errorf("No PEM data was found in the CA certificate file\n")
		}

		cert, err := x509.ParseCertificate(derCert.Bytes)
		if err != nil {
			return nil, nil, fmt.Errorf("Error parsing CA certificate: %s\n", err)
		}

		rootPool := x509.NewCertPool()
		rootPool.AddCert(cert)

		config.ClientAuth = tls.RequireAndVerifyClientCert
		config.ClientCAs = rootPool
	}

	// set up the listener
	var lstnr net.Listener
	conn, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("Error starting TCP listener on %s: %s\n", addr, err)
	}

	lstnr = tls.NewListener(conn, config)

	srv := http.Server{
		Addr:         addr,
		Handler:      nil,
		TLSConfig:    config,
		TLSNextProto: map[string]func(*http.Server, *tls.Conn, http.Handler){},
	}

	return &srv, lstnr, nil
}

// set up encryption key fetcher
// var keyFetcher = ??

// set up routes

// user auth protected routes

// API key protected routes
//  /recordings/uploadmap
//  /recordings/upload?tenant={}&uploadmap={}
