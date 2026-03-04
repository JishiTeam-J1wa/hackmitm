package api

import (
	"fmt"
	"net"
)

// NetworkInterface represents a network interface with its addresses
type NetworkInterface struct {
	Name  string   `json:"name"`
	IPs   []string `json:"ips"`
	Mac   string   `json:"mac"`
}

// ProxyConfig represents the proxy configuration
type ProxyConfig struct {
	// Listener settings
	ProxyPort    int    `json:"proxyPort"`
	BindAddress  string `json:"bindAddress"`
	APIPort      int    `json:"apiPort"`

	// Protocol support
	EnableHTTPS     bool `json:"enableHttps"`
	EnableHTTP2     bool `json:"enableHttp2"`
	EnableWebSocket bool `json:"enableWebSocket"`

	// Upstream proxy
	UpstreamEnabled  bool   `json:"upstreamEnabled"`
	UpstreamAddress  string `json:"upstreamAddress"`
	UpstreamUsername string `json:"upstreamUsername"`
	UpstreamPassword string `json:"upstreamPassword"`

	// Intercept settings
	InterceptMode bool `json:"interceptMode"`

	// SSL/TLS
	CertDir string `json:"certDir"`
}

// ProxyConfigAPI handles proxy configuration operations
type ProxyConfigAPI struct {
	config ProxyConfig
}

// NewProxyConfigAPI creates a new ProxyConfigAPI
func NewProxyConfigAPI() *ProxyConfigAPI {
	return &ProxyConfigAPI{
		config: ProxyConfig{
			ProxyPort:       4443,
			BindAddress:     "127.0.0.1",
			APIPort:         9090,
			EnableHTTPS:     true,
			EnableHTTP2:     true,
			EnableWebSocket: true,
			InterceptMode:   false,
			CertDir:         "./certs",
		},
	}
}

// GetNetworkInterfaces returns all network interfaces with their IP addresses
func (a *ProxyConfigAPI) GetNetworkInterfaces() ([]NetworkInterface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	var result []NetworkInterface
	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		netIface := NetworkInterface{
			Name: iface.Name,
			Mac:  iface.HardwareAddr.String(),
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			// Parse IP address
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			// Only include IPv4 addresses
			if ipNet.IP.To4() != nil {
				netIface.IPs = append(netIface.IPs, ipNet.IP.String())
			}
		}

		if len(netIface.IPs) > 0 {
			result = append(result, netIface)
		}
	}

	return result, nil
}

// GetBindAddressOptions returns available bind address options
func (a *ProxyConfigAPI) GetBindAddressOptions() ([]map[string]string, error) {
	options := []map[string]string{
		{"value": "127.0.0.1", "label": "127.0.0.1 (仅本机)"},
		{"value": "0.0.0.0", "label": "0.0.0.0 (所有接口)"},
	}

	interfaces, err := a.GetNetworkInterfaces()
	if err != nil {
		return options, nil
	}

	for _, iface := range interfaces {
		for _, ip := range iface.IPs {
			options = append(options, map[string]string{
				"value": ip,
				"label": fmt.Sprintf("%s (%s)", ip, iface.Name),
			})
		}
	}

	return options, nil
}

// GetConfig returns the current proxy configuration
func (a *ProxyConfigAPI) GetConfig() ProxyConfig {
	return a.config
}

// SaveConfig saves the proxy configuration
func (a *ProxyConfigAPI) SaveConfig(config ProxyConfig) error {
	a.config = config
	// TODO: Persist to database/config file
	return nil
}

// UpdateConfig updates specific fields of the proxy configuration
func (a *ProxyConfigAPI) UpdateConfig(updates map[string]interface{}) error {
	// TODO: Implement partial updates
	return nil
}

// ValidateConfig validates the proxy configuration
func (a *ProxyConfigAPI) ValidateConfig(config ProxyConfig) map[string]string {
	errors := make(map[string]string)

	if config.ProxyPort < 1 || config.ProxyPort > 65535 {
		errors["proxyPort"] = "端口号必须在 1-65535 之间"
	}

	if config.APIPort < 1 || config.APIPort > 65535 {
		errors["apiPort"] = "API 端口号必须在 1-65535 之间"
	}

	if config.ProxyPort == config.APIPort {
		errors["proxyPort"] = "代理端口和 API 端口不能相同"
	}

	// Validate bind address
	if config.BindAddress != "" {
		ip := net.ParseIP(config.BindAddress)
		if ip == nil {
			errors["bindAddress"] = "无效的 IP 地址"
		}
	}

	// Validate upstream proxy address if enabled
	if config.UpstreamEnabled && config.UpstreamAddress == "" {
		errors["upstreamAddress"] = "上游代理地址不能为空"
	}

	return errors
}
