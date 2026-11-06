package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSMSConfigRefusesMock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sms.env")
	if err := os.WriteFile(path, []byte("SMS_PROVIDER=mock\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SMS_CONFIG", path)
	if _, err := Load(); err == nil {
		t.Fatal("acceptance config must not start as mock")
	}
}

func TestSMSConfigRequiresSynermaxx(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sms.env")
	body := "SMS_PROVIDER=synermaxx\nSYNERMAXX_SMS_APPKEY=test-key\nSYNERMAXX_SMS_URL=https://example.invalid/send\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SMS_CONFIG", path)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.LiveSynermaxx() || cfg.SynermaxxAppKey != "test-key" {
		t.Fatal("locked config did not load synermaxx")
	}
}
