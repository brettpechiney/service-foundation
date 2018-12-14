package config_test

import (
	"fmt"
	"testing"

	"github.com/brettpechiney/workout-service/config"
)

// Test getters by comparing against default values.
func TestGetters(t *testing.T) {
	const Source = "postgresql://maxroach@localhost:26257/workout?sslmode=disable"
	cfg := config.Defaults()
	testCases := []struct {
		Name         string
		TestFunction func() string
		Expected     string
	}{
		{
			"DataSource",
			cfg.DataSource,
			Source,
		},
		{
			"LoggingLevel",
			cfg.LoggingLevel,
			"INFO",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s", tc.Name), func(t *testing.T) {
			if actual := tc.TestFunction(); actual != tc.Expected {
				t.Errorf("expected '%s', got '%s'", tc.Expected, actual)
			}
		})
	}
}
