//go:build linux && cgo

package linux

import (
	"testing"
)

func TestIsUinputVirtualDevice(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"kanata", true},
		{"Kanata Virtual Keyboard", true},
		{"KANATA", true},
		{"AT Translated Set 2 keyboard", false},
		{"Glove80 Keyboard", false},
		{"ThinkPad Extra Buttons", false},
		{"", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isUinputVirtualDevice(-1, tc.name)
			if got != tc.want {
				t.Errorf("isUinputVirtualDevice(-1, %q) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

func TestIsNeruInjectionDevice(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"neru-keyboard", true},
		{"neru-scroll", true},
		{"Neru-Keyboard", true},
		{"kanata", false},
		{"Glove80 Keyboard", false},
		{"", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isNeruInjectionDevice(tc.name)
			if got != tc.want {
				t.Errorf("isNeruInjectionDevice(%q) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}
