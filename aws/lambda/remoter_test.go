package lambda

import (
	"context"
	"testing"

	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/stretchr/testify/assert"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
)

func TestRemoter_Run_Unsupported(t *testing.T) {
	remoter := Remoter{OsWriters: logging.StandardOsWriters{}}

	t.Run("rejects a command", func(t *testing.T) {
		err := remoter.Run(context.Background(), admin.RunOptions{}, []string{"rake", "db:migrate"}, triggerOnlyEnvVars())
		assert.ErrorContains(t, err, "--payload")
	})

	t.Run("rejects user-specified env vars", func(t *testing.T) {
		envVars := triggerOnlyEnvVars()
		envVars["LOG_LEVEL"] = "debug"
		envVars["DRY_RUN"] = "true"

		err := remoter.Run(context.Background(), admin.RunOptions{}, nil, envVars)
		assert.ErrorContains(t, err, "DRY_RUN, LOG_LEVEL")
	})
}

func TestUserEnvVarNames(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    []string
	}{
		{
			name:    "ignores the env vars attached to every run",
			envVars: triggerOnlyEnvVars(),
			want:    []string{},
		},
		{
			name:    "reports user-specified env vars in a stable order",
			envVars: map[string]string{"ZED": "1", "ALPHA": "2", admin.TriggerEnvVar: admin.TriggerManual},
			want:    []string{"ALPHA", "ZED"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, userEnvVarNames(test.envVars))
		})
	}
}

func triggerOnlyEnvVars() map[string]string {
	return map[string]string{
		admin.TriggerEnvVar:     admin.TriggerManual,
		admin.TriggerNameEnvVar: "Fake User",
	}
}
