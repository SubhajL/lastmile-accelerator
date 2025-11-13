package security

import (
    "crypto/tls"
    "crypto/x509"
    "fmt"
    "os"

    "example.com/lma/secrets-env-service/internal/config"
)

// LoadServerTLS returns *tls.Config if cfg.EnableMTLS is true, else nil.
func LoadServerTLS(cfg config.TLSConfig) (*tls.Config, error) {
	if !cfg.EnableMTLS {
		return nil, nil
	}
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil { return nil, err }
    caBytes, err := os.ReadFile(cfg.ClientCAFile)
	if err != nil { return nil, err }
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caBytes) { return nil, fmt.Errorf("failed to append client CA certs") }
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    pool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS12,
	}, nil
}
