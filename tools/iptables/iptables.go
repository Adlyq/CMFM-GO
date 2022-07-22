package iptables

import (
	"Clash.Meta_For_Magisk/config"
	"Clash.Meta_For_Magisk/tools/cmd"
	pkg "Clash.Meta_For_Magisk/tools/package"
	"fmt"
	"net"
)

func SetTproxy(cfg *config.RawConfig) (errs []error) {
	cmdHErr := func(mCmd string, a ...any) {
		_, err := cmd.ExecCmd(fmt.Sprintf(mCmd, a...))
		if err != nil {
			errs = append(errs, err)
		}
	}

	_, dnsPort, err := net.SplitHostPort(cfg.Kernel.Template.DNS.Listen)
	if err != nil {
		return append(errs, err)
	}

	cmdHErr("ip -4 rule add fwmark %d table %d pref %d", cfg.Iptable.MarkID, cfg.Iptable.TableId, cfg.Iptable.PrefID)
	cmdHErr("ip -4 route add local default dev lo table %d", cfg.Iptable.TableId)

	if cfg.IPv6 {
		cmdHErr("ip -6 rule add fwmark %d table %d pref %d", cfg.Iptable.MarkID, cfg.Iptable.TableId, cfg.Iptable.PrefID)
		cmdHErr("ip -6 route add local default dev lo table %d", cfg.Iptable.TableId)
	}

	stp := func(tp string) {
		cmdHErr("%s -t mangle -N DIVERT", tp)
		cmdHErr("%s -t mangle -A DIVERT -j MARK --set-mark %d", tp, cfg.Iptable.MarkID)
		cmdHErr("%s -t mangle -A DIVERT -j ACCEPT", tp)
		cmdHErr("%s -t mangle -A PREROUTING -p tcp -m socket -j DIVERT", tp)
		cmdHErr("%s -t mangle -A PREROUTING -p udp -m socket -j DIVERT", tp)

		cmdHErr("%s -t mangle -N FILTER_LOCAL_IP", tp)
		//TODO
		cmdHErr("%s -t mangle -A PREROUTING -j FILTER_LOCAL_IP", tp)
		cmdHErr("%s -t mangle -A OUTPUT -j FILTER_LOCAL_IP", tp)

		cmdHErr("%s -t mangle -N CLASH_PRE", tp)
		cmdHErr("%s -t mangle -A CLASH_PRE -p tcp -j TPROXY --on-port %d --tproxy-mark %d", tp, cfg.Kernel.Template.TProxyPort, cfg.Iptable.MarkID)
		cmdHErr("%s -t mangle -A CLASH_PRE --p udp ! --dport 53 -j TPROXY --on-port %d --tproxy-mark %d", tp, cfg.Kernel.Template.TProxyPort, cfg.Iptable.MarkID)

		cmdHErr("%s -t mangle -N FILTER_PRE_CLASH", tp)
		for subnet := range cfg.Iptable.Bypass4 {
			cmdHErr("%s -t mangle -A FILTER_PRE_CLASH -d %s -j ACCEPT", tp, subnet)
		}
		cmdHErr("%s -t mangle -A FILTER_PRE_CLASH -j CLASH_PRE", tp)
		cmdHErr("%s -t mangle -A PREROUTING -j FILTER_PRE_CLASH", tp)

		cmdHErr("%s -t mangle -N FILTER_OUT_CLASH", tp)
		for subnet := range cfg.Iptable.Bypass4 {
			cmdHErr("%s -t mangle -A FILTER_OUT_CLASH -d %s -j ACCEPT", tp, subnet)
		}
		for _, mPkg := range cfg.Iptable.Mode.PkgList {
			uid, err := pkg.GetUIDByPkg(mPkg, cfg.SysPkgPth)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			if cfg.Iptable.Mode.Type == "bl" {
				cmdHErr("%s -t mangle -A FILTER_OUT_CLASH -m owner --uid-owner %d -j ACCEPT", tp, uid)
			} else {
				cmdHErr("%s -t mangle -A FILTER_OUT_CLASH -m owner --uid-owner %d -j CLASH_OUT", tp, uid)
			}
		}
		if cfg.Iptable.Mode.Type == "bl" {
			cmdHErr("%s -t mangle -A FILTER_OUT_CLASH -m owner ! --uid-owner %s ! --gid-owner %s -j CLASH_OUT", tp, cfg.Permissions.User, cfg.Permissions.Group)
		}
		cmdHErr("%s -t mangle -A OUTPUT -j FILTER_OUT_CLASH", tp)
	}

	stp("iptables -w 100")
	cmdHErr("iptables -w 100 -t nat -N DNS_PRE")
	cmdHErr("iptables -w 100 -t nat -A DNS_PRE -p udp --dport 53 -j REDIRECT --to-ports %s", dnsPort)
	cmdHErr("iptables -w 100 -t nat -A PREROUTING -j DNS_PRE")

	cmdHErr("iptables -w 100 -t nat -N DNS_OUT")
	cmdHErr("iptables -w 100 -t nat -A DNS_OUT -m owner --gid-owner ${Clash_group} -j ACCEPT")
	cmdHErr("iptables -w 100 -t nat -A DNS_OUT -p udp --dport 53 -j REDIRECT --to-ports %S", dnsPort)
	cmdHErr("iptables -w 100 -t nat -A OUTPUT -j DNS_OUT")

	stp("ip6tables -w 100")
	cmdHErr("ip6tables -I OUTPUT -p udp --dport 53 -j DROP")

	return
}

func CleanTproxy(cfg *config.RawConfig) (errs []error) {
	cmdHErr := func(mCmd string, a ...any) {
		_, err := cmd.ExecCmd(fmt.Sprintf(mCmd, a...))
		if err != nil {
			errs = append(errs, err)
		}
	}

	//TODO()

	ctp := func(tp string) {
		cmdHErr("%s -t mangle -D PREROUTING -p tcp -m socket -j DIVERT", tp)
		cmdHErr("%s -t mangle -D PREROUTING -p udp -m socket -j DIVERT", tp)
		cmdHErr("%s -t mangle -D PREROUTING -j FILTER_LOCAL_IP", tp)
		cmdHErr("%s -t mangle -D OUTPUT -j FILTER_LOCAL_IP", tp)
		cmdHErr("%s -t mangle -D PREROUTING -j FILTER_PRE_CLASH", tp)
		cmdHErr("%s -t mangle -D OUTPUT -j FILTER_OUT_CLASH", tp)

		cmdHErr("%s -t mangle -F DIVERT", tp)
		cmdHErr("%s -t mangle -F FILTER_LOCAL_IP", tp)
		cmdHErr("%s -t mangle -F CLASH_PRE", tp)
		cmdHErr("%s -t mangle -F FILTER_PRE_CLASH", tp)
		cmdHErr("%s -t mangle -F FILTER_OUT_CLASH", tp)

		cmdHErr("%s -t mangle -X DIVERT", tp)
		cmdHErr("%s -t mangle -X FILTER_LOCAL_IP", tp)
		cmdHErr("%s -t mangle -X CLASH_PRE", tp)
		cmdHErr("%s -t mangle -X FILTER_PRE_CLASH", tp)
		cmdHErr("%s -t mangle -X FILTER_OUT_CLASH", tp)

	}

	ctp("iptables -w 100")
	cmdHErr("iptables -w 100 -t nat -D PREROUTING -j DNS_PRE")
	cmdHErr("iptables -w 100 -t nat -D OUTPUT -j DNS_OUT")
	cmdHErr("iptables -w 100 -t nat -F DNS_PRE")
	cmdHErr("iptables -w 100 -t nat -F DNS_OUT")
	cmdHErr("iptables -w 100 -t nat -X DNS_PRE")
	cmdHErr("iptables -w 100 -t nat -X DNS_OUT")

	ctp("ip6tables -w 100")
	cmdHErr("ip6tables -D OUTPUT -p udp --dport 53 -j DROP")

	return
}

func SetTun(cfg *config.RawConfig) (errs []error) {
	cmdHErr := func(mCmd string, a ...any) {
		_, err := cmd.ExecCmd(fmt.Sprintf(mCmd, a...))
		if err != nil {
			errs = append(errs, err)
		}
	}

	cmdHErr("ip -4 rule add fwmark %d table %d pref %d", cfg.Iptable.MarkID, cfg.Iptable.TableId, cfg.Iptable.PrefID)
	cmdHErr("ip -4 route add default dev %s table %d", cfg.Kernel.Template.Tun.Device, cfg.Iptable.TableId)

	if cfg.IPv6 {
		cmdHErr("ip -6 rule add fwmark %d table %d pref %d", cfg.Iptable.MarkID, cfg.Iptable.TableId, cfg.Iptable.PrefID)
		cmdHErr("ip -6 route add default dev %s table %d", cfg.Kernel.Template.Tun.Device, cfg.Iptable.TableId)
	}
	return
}

func CleanTun(cfg *config.RawConfig) (errs []error) {
	return nil
}
