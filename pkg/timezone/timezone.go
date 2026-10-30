// Package timezone configures process-local civil time at startup. Persisted
// timestamps and wire contracts continue to use UTC.
package timezone

import (
	"fmt"
	"os"
	"strings"
	"time"
	_ "time/tzdata" // Keep IANA zones available in standalone binaries as well.
)

const Default = "Asia/Shanghai"

func Load(name string) (*time.Location, error) {
	if name == "" || name == "Local" || strings.HasPrefix(name, "/") || strings.Contains(name, "..") {
		return nil, fmt.Errorf("TZ must be an IANA timezone, for example %s or UTC", Default)
	}
	location, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("invalid TZ %q: %w", name, err)
	}
	return location, nil
}

// Configure must run once in main, before services or background goroutines
// start. Explicitly setting Local also avoids a previously cached host zone.
// Business services receive the resulting location through their constructors.
func Configure() error {
	name, exists := os.LookupEnv("TZ")
	if !exists {
		name = Default
	}
	location, err := Load(name)
	if err != nil {
		return err
	}
	if err = os.Setenv("TZ", name); err != nil {
		return err
	}
	time.Local = location
	return nil
}
