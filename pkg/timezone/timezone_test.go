package timezone

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestConfigureProcess(t *testing.T) {
	for _, tc := range []struct {
		name, zone, want string
		unset, invalid   bool
	}{
		{name: "default", unset: true, want: "Asia/Shanghai|2026-10-31T00:05:00+08:00"},
		{name: "utc", zone: "UTC", want: "UTC|2026-10-30T16:05:00Z"},
		{name: "new-york", zone: "America/New_York", want: "America/New_York|2026-10-30T12:05:00-04:00"},
		{name: "invalid", zone: "Asia/NotAPlace", invalid: true},
		{name: "empty", zone: "", invalid: true},
		{name: "host-alias", zone: "Local", invalid: true},
		{name: "path", zone: "/etc/localtime", invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Configure changes process globals only at boot. Use a fresh process
			// instead of racing time.Local or its lazy initialization in tests.
			command := exec.Command(os.Args[0], "-test.run=^TestTimezoneProcessHelper$")
			for _, value := range os.Environ() {
				if !strings.HasPrefix(value, "TZ=") && !strings.HasPrefix(value, "GOSEED_TZ_TEST_HELPER=") {
					command.Env = append(command.Env, value)
				}
			}
			command.Env = append(command.Env, "GOSEED_TZ_TEST_HELPER=1")
			if !tc.unset {
				command.Env = append(command.Env, "TZ="+tc.zone)
			}
			output, err := command.CombinedOutput()
			if tc.invalid {
				if err == nil || !strings.Contains(string(output), "TZ") {
					t.Fatalf("invalid timezone accepted: %s (%v)", output, err)
				}
			} else if err != nil || strings.TrimSpace(string(output)) != tc.want {
				t.Fatalf("got %q (%v), want %q", output, err, tc.want)
			}
		})
	}
}

func TestTimezoneProcessHelper(t *testing.T) {
	if os.Getenv("GOSEED_TZ_TEST_HELPER") != "1" {
		return
	}
	if err := Configure(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%s|%s\n", os.Getenv("TZ"), time.Date(2026, 10, 30, 16, 5, 0, 0, time.UTC).In(time.Local).Format(time.RFC3339))
	os.Exit(0)
}
