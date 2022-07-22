package MetaCfg

import "C"
import (
	"Clash.Meta_For_Magisk/log"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type Template struct {
	Port               int          `yaml:"port"`
	SocksPort          int          `yaml:"socks-port"`
	RedirPort          int          `yaml:"redir-port"`
	TProxyPort         int          `yaml:"tproxy-port"`
	MixedPort          int          `yaml:"mixed-port"`
	Authentication     []string     `yaml:"authentication"`
	AllowLan           bool         `yaml:"allow-lan"`
	BindAddress        string       `yaml:"bind-address"`
	Mode               TunnelMode   `yaml:"mode"`
	UnifiedDelay       bool         `yaml:"unified-delay"`
	LogLevel           log.LogLevel `yaml:"log-level"`
	IPv6               bool         `yaml:"ipv6"`
	ExternalController string       `yaml:"external-controller"`
	ExternalUI         string       `yaml:"external-ui"`
	Secret             string       `yaml:"secret"`
	Interface          string       `yaml:"interface-name"`
	RoutingMark        int          `yaml:"routing-mark"`
	GeodataMode        bool         `yaml:"geodata-mode"`
	GeodataLoader      string       `yaml:"geodata-loader"`
	TCPConcurrent      bool         `yaml:"tcp-concurrent" json:"tcp-concurrent"`
	EnableProcess      bool         `yaml:"enable-process" json:"enable-process"`

	Sniffer      RawSniffer        `yaml:"sniffer"`
	Hosts        map[string]string `yaml:"hosts"`
	DNS          RawDNS            `yaml:"dns"`
	Tun          RawTun            `yaml:"tun"`
	IPTables     IPTables          `yaml:"iptables"`
	Experimental Experimental      `yaml:"experimental"`
	Profile      Profile           `yaml:"profile"`
	GeoXUrl      RawGeoXUrl        `yaml:"geox-url"`
}

type RawSniffer struct {
	Enable      bool     `yaml:"enable" json:"enable"`
	Sniffing    []string `yaml:"sniffing" json:"sniffing"`
	ForceDomain []string `yaml:"force-domain" json:"force-domain"`
	SkipDomain  []string `yaml:"skip-domain" json:"skip-domain"`
	Ports       []string `yaml:"port-whitelist" json:"port-whitelist"`
}

// Experimental config
type Experimental struct {
	Fingerprints []string `yaml:"fingerprints"`
}

type RawGeoXUrl struct {
	GeoIp   string `yaml:"geoip" json:"geoip"`
	Mmdb    string `yaml:"mmdb" json:"mmdb"`
	GeoSite string `yaml:"geosite" json:"geosite"`
}

// Profile config
type Profile struct {
	StoreSelected bool `yaml:"store-selected"`
	StoreFakeIP   bool `yaml:"store-fake-ip"`
}

type RawDNS struct {
	Enable                bool              `yaml:"enable"`
	IPv6                  bool              `yaml:"ipv6"`
	UseHosts              bool              `yaml:"use-hosts"`
	NameServer            []string          `yaml:"nameserver"`
	Fallback              []string          `yaml:"fallback"`
	FallbackFilter        RawFallbackFilter `yaml:"fallback-filter"`
	Listen                string            `yaml:"listen"`
	EnhancedMode          DNSMode           `yaml:"enhanced-mode"`
	FakeIPRange           string            `yaml:"fake-ip-range"`
	FakeIPFilter          []string          `yaml:"fake-ip-filter"`
	DefaultNameserver     []string          `yaml:"default-nameserver"`
	NameServerPolicy      map[string]string `yaml:"nameserver-policy"`
	ProxyServerNameserver []string          `yaml:"proxy-server-nameserver"`
}

type RawFallbackFilter struct {
	GeoIP     bool     `yaml:"geoip"`
	GeoIPCode string   `yaml:"geoip-code"`
	IPCIDR    []string `yaml:"ipcidr"`
	Domain    []string `yaml:"domain"`
	GeoSite   []string `yaml:"geosite"`
}

type MetaConfig struct {
	*Template
	ProxyProvider map[string]map[string]any `yaml:"proxy-providers"`
	RuleProvider  map[string]map[string]any `yaml:"rule-providers"`
	Proxy         []map[string]any          `yaml:"proxies"`
	ProxyGroup    []map[string]any          `yaml:"proxy-groups"`
	Rule          []string                  `yaml:"rules"`
}

type RawTun struct {
	Enable              bool     `yaml:"enable" json:"enable"`
	Device              string   `yaml:"device" json:"device"`
	Stack               TUNStack `yaml:"stack" json:"stack"`
	DNSHijack           []string `yaml:"dns-hijack" json:"dns-hijack"`
	AutoRoute           bool     `yaml:"auto-route" json:"auto-route"`
	AutoDetectInterface bool     `yaml:"auto-detect-interface"`
}

// IPTables config
type IPTables struct {
	Enable           bool     `yaml:"enable" json:"enable"`
	InboundInterface string   `yaml:"inbound-interface" json:"inbound-interface"`
	Bypass           []string `yaml:"bypass" json:"bypass"`
}

func WriteTo(cfg *MetaConfig, path string) (err error) {
	path, err = filepath.Abs(path)
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o777); err != nil {
			return fmt.Errorf("can't create config directory %s: %s", dir, err.Error())
		}
	}
	bCfg, err := yaml.Marshal(cfg)
	err = os.WriteFile(path, bCfg, 0o644)
	return err
}

func ReadConfig(path string) ([]byte, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("configuration file %s is empty", path)
	}

	return data, err
}

func ParseByte(buf []byte) (*MetaConfig, error) {
	rawCfg, err := UnmarshalRawConfig(buf)
	if err != nil {
		return nil, err
	}
	return rawCfg, nil
}

func ParsePath(path string) (*MetaConfig, error) {
	buf, err := ReadConfig(path)
	if err != nil {
		return nil, err
	}
	return ParseByte(buf)
}

func UnmarshalRawConfig(buf []byte) (*MetaConfig, error) {
	// config with default value
	metaCfg := &MetaConfig{
		Template: &Template{
			AllowLan:       false,
			BindAddress:    "*",
			IPv6:           true,
			Mode:           Rule,
			GeodataMode:    true,
			GeodataLoader:  "memconservative",
			UnifiedDelay:   false,
			Authentication: []string{},
			LogLevel:       log.INFO,
			Hosts:          map[string]string{},
			TCPConcurrent:  false,
			EnableProcess:  false,
			Tun: RawTun{
				Enable:              false,
				Device:              "",
				Stack:               TunGvisor,
				DNSHijack:           []string{"0.0.0.0:53"}, // default hijack all dns query
				AutoRoute:           false,
				AutoDetectInterface: false,
			},
			IPTables: IPTables{
				Enable:           false,
				InboundInterface: "lo",
				Bypass:           []string{},
			},
			DNS: RawDNS{
				Enable:       false,
				IPv6:         false,
				UseHosts:     true,
				EnhancedMode: DNSMapping,
				FakeIPRange:  "198.18.0.1/16",
				FallbackFilter: RawFallbackFilter{
					GeoIP:     true,
					GeoIPCode: "CN",
					IPCIDR:    []string{},
					GeoSite:   []string{},
				},
				DefaultNameserver: []string{
					"114.114.114.114",
					"223.5.5.5",
					"8.8.8.8",
					"1.0.0.1",
				},
				NameServer: []string{
					"https://doh.pub/dns-query",
					"tls://223.5.5.5:853",
				},
				FakeIPFilter: []string{
					"dns.msftnsci.com",
					"www.msftnsci.com",
					"www.msftconnecttest.com",
				},
			},
			Sniffer: RawSniffer{
				Enable:      false,
				Sniffing:    []string{},
				ForceDomain: []string{},
				SkipDomain:  []string{},
				Ports:       []string{},
			},
			Profile: Profile{
				StoreSelected: true,
			},
			GeoXUrl: RawGeoXUrl{
				GeoIp:   "https://ghproxy.com/https://raw.githubusercontent.com/Loyalsoldier/v2ray-rules-dat/release/geoip.dat",
				Mmdb:    "https://ghproxy.com/https://raw.githubusercontent.com/Loyalsoldier/geoip/release/Country.mmdb",
				GeoSite: "https://ghproxy.com/https://raw.githubusercontent.com/Loyalsoldier/v2ray-rules-dat/release/geosite.dat",
			},
		},
		Rule:       []string{},
		Proxy:      []map[string]any{},
		ProxyGroup: []map[string]any{},
	}

	if err := yaml.Unmarshal(buf, metaCfg); err != nil {
		return nil, err
	}

	return metaCfg, nil
}
