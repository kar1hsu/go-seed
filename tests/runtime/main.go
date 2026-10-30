// runtime verifies the actual final service image without external services.
// It deliberately does not embed time/tzdata: the image must provide it.
package main

import (
	"crypto/x509"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func check() error {
	if os.Getuid() != 65532 || os.Getgid() != 65532 {
		return fmt.Errorf("unexpected runtime identity: %d:%d", os.Getuid(), os.Getgid())
	}
	for _, name := range []string{"/bin/sh", "/bin/bash", "/usr/bin/apt-get", "/usr/local/go/bin/go"} {
		if _, err := os.Stat(name); !os.IsNotExist(err) {
			return fmt.Errorf("unexpected runtime tool: %s (%v)", name, err)
		}
	}
	certs, err := x509.SystemCertPool()
	if err != nil {
		return err
	}
	if len(certs.Subjects()) == 0 {
		return fmt.Errorf("missing CA roots")
	}
	expected := map[string]string{
		"Asia/Shanghai":    "20261031 +0800",
		"UTC":              "20261030 +0000",
		"America/New_York": "20261030 -0400",
	}
	zone, err := time.LoadLocation(os.Getenv("TZ"))
	if err != nil {
		return err
	}
	moment := time.Date(2026, 10, 30, 16, 5, 0, 0, time.UTC)
	if want, ok := expected[zone.String()]; !ok || moment.In(zone).Format("20060102 -0700") != want || moment.In(time.Local).Format("20060102 -0700") != want {
		return fmt.Errorf("unexpected configured/local timezone: %s/%s", zone, time.Local)
	}
	dir, err := os.MkdirTemp("/app/data/files", ".runtime-check-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	file := filepath.Join(dir, "sample.txt")
	if err = os.WriteFile(file, []byte("upload-volume-check"), 0600); err != nil {
		return err
	}
	data, err := os.ReadFile(file)
	if err != nil || string(data) != "upload-volume-check" {
		return fmt.Errorf("file readback failed: %v", err)
	}
	for _, binary := range []string{"server", "probe", "systemctl", "memberctl"} {
		if out, err := exec.Command("/app/"+binary, "-h").CombinedOutput(); err != nil {
			return fmt.Errorf("cannot execute %s without shell: %w (%s)", binary, err, out)
		}
	}
	fmt.Printf("PASS: nonroot, shellless, CA roots, %s, writable fresh upload volume, service/probe/domain binaries\n", zone)
	return nil
}

func main() {
	if err := check(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
