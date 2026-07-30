package lambda

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
)

func TestNormalizePayload(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    string
		wantErr bool
	}{
		{
			name:    "empty payload sends an empty json document",
			payload: "",
			want:    EmptyPayload,
		},
		{
			name:    "whitespace payload sends an empty json document",
			payload: "  \n",
			want:    EmptyPayload,
		},
		{
			name:    "json object is passed through",
			payload: `{"name":"world"}`,
			want:    `{"name":"world"}`,
		},
		{
			name:    "json scalar is passed through",
			payload: `"hello"`,
			want:    `"hello"`,
		},
		{
			name:    "invalid json is rejected",
			payload: `{"name":`,
			wantErr: true,
		},
		{
			name:    "raw string is rejected",
			payload: `hello`,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := normalizePayload([]byte(test.payload))
			if test.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.want, string(got))
		})
	}
}

func TestCreateClientContext(t *testing.T) {
	got, err := createClientContext("Fake User")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.LessOrEqual(t, len(*got), MaxClientContextLength)

	raw, err := base64.StdEncoding.DecodeString(*got)
	require.NoError(t, err, "client context must be base64-encoded")

	var decoded struct {
		Custom map[string]string `json:"custom"`
	}
	require.NoError(t, json.Unmarshal(raw, &decoded))
	assert.Equal(t, admin.TriggerManual, decoded.Custom[admin.TriggerEnvVar])
	assert.Equal(t, "Fake User", decoded.Custom[admin.TriggerNameEnvVar])
}

func TestEmitLogs(t *testing.T) {
	requestId := "8bd3e1d0-3f31-4b62-9b0e-1f0a5e5d9a11"
	completeLog := fmt.Sprintf("START RequestId: %s Version: $LATEST\nhello world\nEND RequestId: %s\n", requestId, requestId)

	tests := []struct {
		name          string
		logResult     *string
		wantMessages  []string
		wantStream    string
		wantTruncated bool
	}{
		{
			name:         "emits every line and scopes them to the request id",
			logResult:    encodeLog(completeLog),
			wantMessages: []string{fmt.Sprintf("START RequestId: %s Version: $LATEST", requestId), "hello world", fmt.Sprintf("END RequestId: %s", requestId)},
			wantStream:   requestId,
		},
		{
			name:          "warns when lambda truncated the log output",
			logResult:     encodeLog(fmt.Sprintf("of a really long line\nEND RequestId: %s\n", requestId)),
			wantMessages:  []string{"of a really long line", fmt.Sprintf("END RequestId: %s", requestId)},
			wantStream:    requestId,
			wantTruncated: true,
		},
		{
			name:         "falls back to the function name when there is no request id",
			logResult:    encodeLog("START RequestId: unknown\n"),
			wantMessages: []string{"START RequestId: unknown"},
			wantStream:   "fake-function",
		},
		{
			name:         "tolerates an invocation with no logs",
			logResult:    nil,
			wantMessages: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			infra := Outputs{LambdaName: "fake-function"}
			messages := make([]app.LogMessage, 0)
			emitter := func(message app.LogMessage) {
				messages = append(messages, message)
			}
			var stderr bytes.Buffer

			emitLogs(infra, test.logResult, emitter, &stderr)

			gotMessages := make([]string, 0)
			for _, message := range messages {
				gotMessages = append(gotMessages, message.Message)
				assert.Equal(t, test.wantStream, message.Stream)
				assert.Equal(t, infra.LambdaName, message.Source)
			}
			assert.Equal(t, test.wantMessages, gotMessages)
			if test.wantTruncated {
				assert.Contains(t, stderr.String(), "nullstone logs")
			} else {
				assert.Empty(t, stderr.String())
			}
		})
	}
}

func encodeLog(contents string) *string {
	encoded := base64.StdEncoding.EncodeToString([]byte(contents))
	return &encoded
}
