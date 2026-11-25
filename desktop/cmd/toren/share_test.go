package main

import (
	"testing"
)

func TestShareUnavailableWhenTailscaleOff(t *testing.T) {
	// By default /var/service/tailscaled is not present in test environment
	code := asuraShare([]string{"devices"})
	if code != exitFail {
		t.Errorf("expected exitFail when tailscale is off, got %d", code)
	}
}

func TestShareSendMissingFile(t *testing.T) {
	oldStatus := shareTailscaleStatus
	defer func() { shareTailscaleStatus = oldStatus }()
	shareTailscaleStatus = func() ([]byte, error) {
		return []byte(`{"Self":{"HostName":"box","DNSName":"box.ts.net","Online":true},"Peer":{}}`), nil
	}

	// Send missing file
	code := shareSend("/nonexistent/file.txt", "myphone")
	if code != exitFail {
		t.Errorf("expected exitFail for nonexistent file, got %d", code)
	}
}

func TestShareDevicesParse(t *testing.T) {
	oldStatus := shareTailscaleStatus
	defer func() { shareTailscaleStatus = oldStatus }()
	shareTailscaleStatus = func() ([]byte, error) {
		return []byte(`{
			"Self": {"HostName": "asura-laptop", "DNSName": "asura-laptop.net", "Online": true},
			"Peer": {
				"peer1": {"HostName": "phone", "DNSName": "phone.net", "Online": true, "OS": "iOS"},
				"peer2": {"HostName": "workstation", "DNSName": "workstation.net", "Online": false, "OS": "linux"}
			}
		}`), nil
	}

	code := shareDevices()
	if code != exitOK {
		t.Errorf("shareDevices exited %d, want exitOK", code)
	}
}
