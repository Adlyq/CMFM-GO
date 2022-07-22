package config

import (
	"Clash.Meta_For_Magisk/config/MetaCfg"
	"Clash.Meta_For_Magisk/log"
	"Clash.Meta_For_Magisk/tools"
	"fmt"
	"github.com/dlclark/regexp2"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

var (
	Cfg *RawConfig

	DefBypass4 = []string{
		"0.0.0.0/8",
		"10.0.0.0/8",
		"100.64.0.0/10",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"172.16.0.0/12",
		"192.0.0.0/24",
		"192.0.2.0/24",
		"192.168.0.0/16",
		"198.51.100.0/24",
		"203.0.113.0/24",
		"224.0.0.0/4",
		"255.255.255.255/32",
		"240.0.0.0/4",
	}

	DefBypass6 = []string{
		"::/128",
		"::1/128",
		"::ffff:0:0/96",
		"100::/64",
		"64:ff9b::/96",
		"2001::/32",
		"2001:10::/28",
		"2001:20::/28",
		"2001:db8::/32",
		"2002::/16",
		"fc00::/7",
		"fe80::/10",
		"ff00::/8",
	}
)

type RawKernel struct {
	Bin      string            `yaml:"bin"`
	DataDir  string            `yaml:"data-dir"`
	UseTpl   bool              `yaml:"use-template"`
	Template *MetaCfg.Template `yaml:"template"`
	Config   string            `yaml:"config"`
	Log      string            `yaml:"log"`
}

type RawUpdate struct {
	Corn            string `yaml:"corn"`
	SubscriptionUrl string `yaml:"subscription-url"`
}

type RawCgroups struct {
	MemoryPath  string `yaml:"memory-path"`
	MemoryLimit string `yaml:"memory-limit"`
	CupPath     string `yaml:"cpu-path"`
	CupLimit    int    `yaml:"cpu-limit"`
}

type RawPermissions struct {
	User  string `yaml:"user"`
	Group string `yaml:"group"`
}

type RawMode struct {
	Type    string   `yaml:"type"`
	PkgList []string `yaml:"pkg-list"`
}

type RawIptables struct {
	PrefID  int      `yaml:"pref-id"`
	MarkID  int      `yaml:"mark-id"`
	TableId int      `yaml:"table-id"`
	Mode    RawMode  `yaml:"mode"`
	UseDef4 bool     `yaml:"use-def4"`
	UseDef6 bool     `yaml:"use-def6"`
	Bypass4 []string `yaml:"bypass4"`
	Bypass6 []string `yaml:"bypass6"`
}

type RawConfig struct {
	Busybox     string         `yaml:"busybox-path"`
	SysPkgPth   string         `yaml:"system-pkg-file"`
	IPv6        bool           `yaml:"ipv6"`
	LogLevel    log.LogLevel   `yaml:"log-level"`
	Kernel      RawKernel      `yaml:"kernel"`
	Update      RawUpdate      `yaml:"update"`
	Cgroups     RawCgroups     `yaml:"cgroups"`
	Permissions RawPermissions `yaml:"permissions"`
	Iptable     RawIptables    `yaml:"iptables"`
}

// ParseByte Parse config
func ParseByte(buf []byte) (*RawConfig, error) {
	rawCfg, err := UnmarshalRawConfig(buf)
	if err != nil {
		return nil, err
	}
	return rawCfg, nil
}

func ParsePath(path string) (*RawConfig, error) {
	buf, err := MetaCfg.ReadConfig(path)
	if err != nil {
		return nil, err
	}
	return ParseByte(buf)
}

func UnmarshalRawConfig(buf []byte) (*RawConfig, error) {
	rawCfg := &RawConfig{
		Busybox:   "/data/adb/magisk/busybox",
		SysPkgPth: "/data/system/packages.list",
		LogLevel:  log.INFO,
		IPv6:      true,
		Kernel: RawKernel{
			Bin:      "Clash.Meta",
			DataDir:  "/data/clash",
			UseTpl:   false,
			Config:   "",
			Log:      "",
			Template: nil,
		},
		Update: RawUpdate{
			Corn:            "0 5 * * *",
			SubscriptionUrl: "",
		},
		Cgroups: RawCgroups{
			MemoryPath:  "",
			MemoryLimit: "",
			CupPath:     "",
			CupLimit:    0,
		},
		Permissions: RawPermissions{
			User:  "root",
			Group: "net_admin",
		},
		Iptable: RawIptables{
			PrefID:  5000,
			MarkID:  2022,
			TableId: 2022,
			Mode: RawMode{
				Type:    "bl",
				PkgList: []string{},
			},
			UseDef4: true,
			UseDef6: true,
			Bypass4: []string{},
			Bypass6: []string{},
		},
	}

	if err := yaml.Unmarshal(buf, rawCfg); err != nil {
		return nil, err
	} else {
		if len(DefBypass4) > 0 && rawCfg.Iptable.UseDef4 {
			rawCfg.Iptable.Bypass4 = tools.Set(append(rawCfg.Iptable.Bypass4, DefBypass4...))
		}
		if len(DefBypass6) > 0 && rawCfg.Iptable.UseDef6 {
			rawCfg.Iptable.Bypass6 = tools.Set(append(rawCfg.Iptable.Bypass6, DefBypass6...))
		}
		if rawCfg.Kernel.Config == "" {
			rawCfg.Kernel.Config = filepath.Join(rawCfg.Kernel.DataDir, "config.yaml")
		}
		return rawCfg, nil
	}
}

func VerifyCfg(cfg *RawConfig) []error {
	var errs []error
	isExist := func(name, path string) {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("could not find %s at %s", name, path))
		}
	}
	isAbs := func(name, path string) {
		if path != "" && !filepath.IsAbs(path) {
			errs = append(errs, fmt.Errorf("%s mast be abstract: %s", name, path))
		}
	}
	if runtime.GOARCH == "android" {
		isExist("BusyBox", cfg.Busybox)
	}
	if filepath.IsAbs(cfg.Kernel.Bin) {
		isExist("kernel", cfg.Kernel.Bin)
	} else {
		isExist("kernel", filepath.Join("/bin", cfg.Kernel.Bin))
	}
	isExist("clash data dir", cfg.Kernel.DataDir)
	isExist("clash config file", cfg.Kernel.Config)
	tp := cfg.Iptable.Mode.Type
	if tp != "bl" && tp != "wl" && tp != "core" {
		errs = append(errs, fmt.Errorf("iptable.mode.type mast be bl, wl or core"))
	}

	isAbs("kernel.log", cfg.Kernel.Log)
	isAbs("memory-path", cfg.Cgroups.MemoryPath)
	isAbs("cuu-path", cfg.Cgroups.CupPath)

	urlVerifyReg := regexp2.MustCompile("(https?|ftp|file)://[-A-Za-z0-9+&@#/%?=~_|!:,.;]+[-A-Za-z0-9+&@#/%=~_|]", 0)
	if mat, _ := urlVerifyReg.FindStringMatch(cfg.Update.SubscriptionUrl); mat == nil && cfg.Update.SubscriptionUrl != "" {
		errs = append(errs, fmt.Errorf("url: %s is not support", cfg.Update.SubscriptionUrl))
	}

	ml := cfg.Cgroups.MemoryLimit
	if ml != "" && ml != "-1" {
		_, err := strconv.Atoi(ml[:len(ml)-1])
		if err != nil || ml[len(ml)-1:] != "M" {
			errs = append(errs, fmt.Errorf("cgroup.memory-limit: unsupported formats %s", ml))
		}
	}

	return errs
}
