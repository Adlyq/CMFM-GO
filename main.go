//go:build linux || android

package main

import (
	"Clash.Meta_For_Magisk/config"
	C "Clash.Meta_For_Magisk/constant"
	"Clash.Meta_For_Magisk/log"
	"Clash.Meta_For_Magisk/tools/kernel"
	"flag"
	"fmt"
	"runtime"
)

var (
	flagSet    map[string]bool
	version    bool
	configPath string
	testConfig bool
	demo       bool
	start      bool
	stop       bool
)

func init() {
	flag.BoolVar(&demo, "u", false, "更新")
	flag.BoolVar(&start, "s", false, "启动")
	flag.BoolVar(&stop, "k", false, "停止")
	flag.BoolVar(&version, "v", false, "显示版本号")
	flag.StringVar(&configPath, "f", "/data/clash/cmfm.yaml", "配置文件路径")
	flag.BoolVar(&testConfig, "t", false, "测试配置文件，然后退出")
	flag.Parse()

	flagSet = map[string]bool{}
	flag.Visit(func(f *flag.Flag) {
		flagSet[f.Name] = true
	})
}

func main() {
	if version {
		fmt.Printf("%s %s %s %s with %s %s\n", C.Name, C.Version, runtime.GOOS, runtime.GOARCH, runtime.Version(), C.BuildTime)
		return
	}
	rawCfg, err := config.ParsePath(configPath)
	if err != nil {
		log.Errorln("%s", err)
		return
	}

	if testConfig {
		if errs := config.VerifyCfg(rawCfg); len(errs) > 0 {
			for _, err := range errs {
				log.Errorln("%s", err)
			}
			log.Errorln("test fail")
			return
		}
		if err := kernel.TestClashConfig(rawCfg); err != nil {
			log.Errorln("%s", err)
			log.Errorln("test fail")
			return
		}
		log.Infoln("all test pass")
		return
	}

	if errs := config.VerifyCfg(rawCfg); len(errs) > 0 {
		for _, err := range errs {
			log.Errorln("%s", err)
		}
		return
	}

	log.SetLevel(rawCfg.LogLevel)
	config.Cfg = rawCfg

	switch {
	case start:
		err := kernel.StartClash(rawCfg)
		if err != nil {
			log.Errorln("%s", err)
			return
		}
	}

}
