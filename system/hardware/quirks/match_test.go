package quirks_test

import (
	"testing"

	"toren/hardware/probe"
	"toren/hardware/quirks"
)

func TestModuleMatchIsExact(t *testing.T) {
	def := quirks.QuirkDef{ID: "t", Layer: quirks.LayerComponent, Match: quirks.QuirkMatch{Module: []string{"hid_sensor_hub"}}}
	if ok, _ := quirks.MatchQuirk(def, probe.Fields{}, nil, nil, probe.ACPIState{}, []string{"hid", "usbhid"}); ok {
		t.Fatal("hid_sensor_hub matched a machine that only has hid")
	}
	if ok, _ := quirks.MatchQuirk(def, probe.Fields{}, nil, nil, probe.ACPIState{}, []string{"hid", "hid_sensor_hub"}); !ok {
		t.Fatal("hid_sensor_hub did not match a machine that has it")
	}
	glob := quirks.QuirkDef{ID: "g", Layer: quirks.LayerComponent, Match: quirks.QuirkMatch{Module: []string{"snd_sof*"}}}
	if ok, _ := quirks.MatchQuirk(glob, probe.Fields{}, nil, nil, probe.ACPIState{}, []string{"snd_sof_pci"}); !ok {
		t.Fatal("snd_sof* did not match snd_sof_pci")
	}
}

func TestFieldPatternDoesNotMatchAShorterValue(t *testing.T) {
	if quirks.MatchFieldPattern("ThinkPad X1 Carbon", "ThinkPad") {
		t.Fatal("a pattern matched a value that is only its beginning")
	}
	if !quirks.MatchFieldPattern("ThinkPad", "ThinkPad T14 Gen 3") {
		t.Fatal("a pattern did not match a longer value that starts with it")
	}
}
