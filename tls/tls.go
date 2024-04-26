package tls

import (
	"crypto/tls"
	"crypto/x509"
)

func WithCertificate(ca, cert, key []byte, mutual bool) *tls.Config {
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(ca)

	keyPair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{keyPair},
		RootCAs:      caCertPool,
		MinVersion:   tls.VersionTLS12,
	}

	if mutual {
		tlsCfg.ClientCAs = caCertPool
		tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return tlsCfg
}

func WithCertificatePair(cert, key []byte) *tls.Config {
	keyPair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{keyPair},
		MinVersion:   tls.VersionTLS12,
	}

	return tlsCfg
}

func WithCA(ca []byte) *tls.Config {
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(ca)

	tlsCfg := &tls.Config{
		RootCAs:    caCertPool,
		MinVersion: tls.VersionTLS12,
	}

	return tlsCfg
}

func WithServerAndCA(serverName string, ca []byte) *tls.Config {
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(ca)

	tlsCfg := &tls.Config{
		RootCAs:    caCertPool,
		ServerName: serverName,
		MinVersion: tls.VersionTLS12,
	}

	return tlsCfg
}
