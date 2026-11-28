package main

import (
	"testing"
)

func TestXbpsProviderCategory(t *testing.T) {
	p := xbpsProvider{}
	cat := p.Category()
	if cat.ID != "apps" {
		t.Fatalf("expected category ID 'apps', got %q", cat.ID)
	}
	if cat.Group != "find" {
		t.Fatalf("expected category Group 'find', got %q", cat.Group)
	}
}

func TestXbpsPackageNameValidation(t *testing.T) {
	valid := []string{"vlc", "obs-studio", "libx11", "gcc-14.2.0_1", "linux6.12"}
	for _, v := range valid {
		if !xbpsPkgName.MatchString(v) {
			t.Errorf("expected %q to be valid xbps package name", v)
		}
	}
	invalid := []string{"-vlc", "vlc;rm -rf /", "vlc&foo", "vlc|bar", ""}
	for _, inv := range invalid {
		if xbpsPkgName.MatchString(inv) {
			t.Errorf("expected %q to be rejected as xbps package name", inv)
		}
	}
}

func TestFlatpakIDGrammar(t *testing.T) {
	for _, valid := range []string{"org.mozilla.firefox", "com.discordapp.Discord", "md.obsidian.Obsidian"} {
		if !isFlatpakID(valid) {
			t.Errorf("expected %q to be valid flatpak ID", valid)
		}
	}
	for _, invalid := range []string{"firefox", "vlc", "../bad.id", "two.parts"} {
		if isFlatpakID(invalid) {
			t.Errorf("expected %q to be rejected as flatpak ID", invalid)
		}
	}
}
