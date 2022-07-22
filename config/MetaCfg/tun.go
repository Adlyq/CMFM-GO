package MetaCfg

import (
	"encoding/json"
	"errors"
	"strings"
)

var StackTypeMapping = map[string]TUNStack{
	strings.ToUpper(TunGvisor.String()): TunGvisor,
	strings.ToUpper(TunSystem.String()): TunSystem,
}

const (
	TunGvisor TUNStack = iota
	TunSystem
)

type TUNStack int

// UnmarshalYAML unserialize TUNStack with yaml
func (e *TUNStack) UnmarshalYAML(unmarshal func(any) error) error {
	var tp string
	if err := unmarshal(&tp); err != nil {
		return err
	}
	mode, exist := StackTypeMapping[strings.ToUpper(tp)]
	if !exist {
		return errors.New("invalid tun stack")
	}
	*e = mode
	return nil
}

// MarshalYAML serialize TUNStack with yaml
func (e TUNStack) MarshalYAML() (any, error) {
	return e.String(), nil
}

// UnmarshalJSON unserialize TUNStack with json
func (e *TUNStack) UnmarshalJSON(data []byte) error {
	var tp string
	json.Unmarshal(data, &tp)
	mode, exist := StackTypeMapping[strings.ToUpper(tp)]
	if !exist {
		return errors.New("invalid tun stack")
	}
	*e = mode
	return nil
}

// MarshalJSON serialize TUNStack with json
func (e TUNStack) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e TUNStack) String() string {
	switch e {
	case TunGvisor:
		return "gVisor"
	case TunSystem:
		return "System"
	default:
		return "unknown"
	}
}

type TunnelMode int

// ModeMapping is a mapping for Mode enum
var ModeMapping = map[string]TunnelMode{
	Global.String(): Global,
	Rule.String():   Rule,
	Script.String(): Script,
	Direct.String(): Direct,
}

const (
	Global TunnelMode = iota
	Rule
	Script
	Direct
)

// UnmarshalJSON unserialize Mode
func (m *TunnelMode) UnmarshalJSON(data []byte) error {
	var tp string
	json.Unmarshal(data, &tp)
	mode, exist := ModeMapping[strings.ToLower(tp)]
	if !exist {
		return errors.New("invalid mode")
	}
	*m = mode
	return nil
}

// UnmarshalYAML unserialize Mode with yaml
func (m *TunnelMode) UnmarshalYAML(unmarshal func(any) error) error {
	var tp string
	unmarshal(&tp)
	mode, exist := ModeMapping[strings.ToLower(tp)]
	if !exist {
		return errors.New("invalid mode")
	}
	*m = mode
	return nil
}

// MarshalJSON serialize Mode
func (m TunnelMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

// MarshalYAML serialize TunnelMode with yaml
func (m TunnelMode) MarshalYAML() (any, error) {
	return m.String(), nil
}

func (m TunnelMode) String() string {
	switch m {
	case Global:
		return "global"
	case Rule:
		return "rule"
	case Script:
		return "script"
	case Direct:
		return "direct"
	default:
		return "Unknown"
	}
}
