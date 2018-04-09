package config

import (
	"encoding/json"
	"io/ioutil"
)

// Server contains the configuration information required to start a
// redoctober server.
type Server struct {
	// The host:port that the server should listen on
	Addr string `json:"address"`
	// The public-facing host + port part of the URL
	// (Used to construct Okta-related URLs)
	PublicHost string `json:"publichost"`
	// URL to use to get to the root of the application
	BaseURL string `json:"baseurl"`
	// Path to static content (callpicker2-ui et al.)
	StaticPath string `json:"staticpath"`

	// KeyPaths and CertPaths contains a list of paths to TLS key
	// pairs that should be used to secure connections to the
	// server. The paths should be comma-separated.
	KeyPaths  string `json:"private_keys"`
	CertPaths string `json:"certificates"`
	// CAPath contains the path to the TLS CA for client
	// authentication. This is an optional field.
	CAPath string `json:"capath,omitempty"`

	// DB configuration
	DBHost string `json:"dbhost"`
	DBName string `json:"dbname"`
	DBUserName string `json:"dbuser"`
	DBPassword string `json:"dbpass"`

	// Parent recordings folder
	RecordingsFolder string `'json:"recordingspath"`

	// Enable endpoint for direct key operations
	KeyEndpointEnabled bool `json:"keyepenabled"`
	// Secret API key to auth the key endpoint
	KeyEndpointAPIKey string `json:"keyepapikey"`

	// RedOctober server URL
	ROServer string `json:"roserver"`
	// Path to the RedOctober public certificate file
	ROCertFile string `json:"rocertfile"`
	// CallPicker2 user credentials in RedOctober
	ROUserName string `json:"rousername"`
	ROPassword string `json:"ropassword"`
	// Minimum users required to decrypt a key
	ROMinimumUsers int `json:"rominusers"`
	// Users that can delegate permission to RedOctober to decrypt data
	RODelegates []string `json:"rodelegates"`
}

// Transcoder information
type Transcoder struct {
	// The transcoder program/binary
	Executable string `json:"executable"`
	// Arguments for the transcoder
	Args []string `json:"args"`
}

// Okta contains Okta related configuration information
type Okta struct {
	// Okta host (without the 'https://')
	Host         string `json:"host"`
	// ID of the Okta application
	ClientID     string `json:"clientid"`
	// Secret of the Okta application
	ClientSecret string `json:"clientsecret"`
	// Relative path for login callback
	LoginPath    string `json:"loginpath"`
	// Relative path for auth callback
	AuthPath     string `json:"authpath"`
}

// Config contains all the configuration for a callpicker2 instance.
type Config struct {
	Server     *Server     `json:"server"`
	Okta       *Okta       `json:"okta"`
	Transcoder *Transcoder `json:"transcoder"`
}

// Valid returns true if the configuration is valid
func (c *Config) Valid() bool {
	// API uses TLS for security
	if len(c.Server.CertPaths) == 0 || len(c.Server.KeyPaths) == 0 {
		return false
	}

	// API needs a RedOctober server in order to decrypt recording encryption keys
	if len(c.Server.ROServer) == 0 {
		return false
	}

	return true
}

// New returns a freshly built config
func New() *Config {
	return &Config{
		Server:     &Server{},
		Okta:       &Okta{},
		Transcoder: &Transcoder{},
	}
}

// Load reads a JSON-encoded config file from disk.
func Load(path string) (*Config, error) {
	cfg := New()
	in, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(in, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
