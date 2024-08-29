package config

// httpBehavior is the configuration for how and what the proxy does with HTTP traffic
type httpBehavior struct {
	Listen                string // Local address the proxy should listen on
	CertDir               string // Dir to the certificate, for TLS MITM
	InsecureSkipVerifyTLS bool   // if true, MITM will not verify the TLS certificate of the target server
	NoHTTPUpgrader        bool   // if true, the proxy will NOT upgrade http requests to https
}
