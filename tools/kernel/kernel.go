package kernel

import (
	"Clash.Meta_For_Magisk/config"
	"Clash.Meta_For_Magisk/config/MetaCfg"
	"Clash.Meta_For_Magisk/log"
	"Clash.Meta_For_Magisk/tools/cmd"
	"fmt"
	"path/filepath"
)

func StartClash(cfg *config.RawConfig) error {
	metaCfg, err := perHandleCfg(cfg)
	if err != nil {
		return err
	}
	err = MetaCfg.WriteTo(metaCfg, filepath.Join(cfg.Kernel.DataDir, "./run/config.yaml"))
	if err != nil {
		return err
	}

	_, err = cmd.ExecCmdP(cfg.Busybox, "")
	if err != nil {
		return err
	}
	return nil
}

func TestClashConfig(cfg *config.RawConfig) error {
	metaCfg, err := perHandleCfg(cfg)
	if err != nil {
		return err
	}
	err = MetaCfg.WriteTo(metaCfg, filepath.Join(cfg.Kernel.DataDir, "./run/config.yaml"))
	if err != nil {
		return err
	}

	output, err := cmd.ExecCmdP(cfg.Kernel.Bin, fmt.Sprintf("-d %s -f %s -t", cfg.Kernel.DataDir, filepath.Join(cfg.Kernel.DataDir, "./run/config.yaml")))
	if err != nil {
		return fmt.Errorf("%s: %s", err, output)
	}

	return nil
}

func perHandleCfg(cfg *config.RawConfig) (*MetaCfg.MetaConfig, error) {
	kCfg := cfg.Kernel
	mCfg, err := MetaCfg.ParsePath(kCfg.Config)
	if err != nil {
		return nil, err
	}
	if kCfg.UseTpl {
		mCfg.Template = kCfg.Template
	} else {
		kCfg.Template = mCfg.Template
	}

	mCfg.IPv6 = cfg.IPv6
	mCfg.DNS.IPv6 = cfg.IPv6

	switch {
	case mCfg.Tun.Enable:
		mCfg.TProxyPort = 0
		mCfg.Tun.AutoRoute = true
	case mCfg.TProxyPort == 0:
		return nil, fmt.Errorf("must choose one in Tun or Tproxy")
	case !mCfg.DNS.Enable:
		return nil, fmt.Errorf("dns must be enabled when use Tproxy")
	case mCfg.DNS.EnhancedMode == MetaCfg.DNSFakeIP:
		log.Warnln("Tproxy mod + Fake-IP is not support blacklist or whitelist")
		cfg.Iptable.Mode.PkgList = nil
	}

	mCfg.IPTables.Enable = false
	kCfg.Template = mCfg.Template

	return mCfg, nil
}
