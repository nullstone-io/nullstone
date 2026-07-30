package lambda

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/nullstone-io/deployment-sdk/app"
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
	"github.com/nullstone-io/deployment-sdk/logging"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
)

const (
	// MaxClientContextLength is the maximum length of the base64-encoded client context accepted by the Invoke API
	MaxClientContextLength = 3583

	// EmptyPayload is sent when the user does not specify a payload
	// Many lambda runtimes fail to deserialize an empty event, so we always send a valid json document
	EmptyPayload = "{}"

	logSourceType = "cloudwatch"
)

var (
	triggerEnvVars = map[string]bool{
		admin.TriggerEnvVar:     true,
		admin.TriggerNameEnvVar: true,
	}

	// requestIdRegex matches the request id that lambda reports in the START/END/REPORT log lines
	requestIdRegex = regexp.MustCompile(`RequestId: ([0-9a-fA-F-]{36})`)
)

// Invoke performs a synchronous invocation of the lambda function
// The logs for the invocation are emitted through options.LogEmitter and the response payload is written to stdout
func Invoke(ctx context.Context, osWriters logging.OsWriters, infra Outputs, options admin.RunOptions) error {
	stdout, stderr := osWriters.Stdout(), osWriters.Stderr()

	payload, err := normalizePayload(options.Payload)
	if err != nil {
		return err
	}
	clientContext, err := createClientContext(options.Username)
	if err != nil {
		return err
	}

	awsConfig := nsaws.NewConfig(infra.Runner, infra.Region)
	// An invocation is not necessarily idempotent -- a retried request can execute the function a second time
	// The default retryer retries on timeouts, so we disable retries entirely
	lambdaClient := lambda.NewFromConfig(awsConfig, func(o *lambda.Options) {
		o.Retryer = aws.NopRetryer{}
	})

	out, err := lambdaClient.Invoke(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(infra.LambdaName),
		InvocationType: types.InvocationTypeRequestResponse,
		LogType:        types.LogTypeTail,
		ClientContext:  clientContext,
		Payload:        payload,
	})
	if err != nil {
		return fmt.Errorf("error invoking function: %w", err)
	}

	emitLogs(infra, out.LogResult, options.LogEmitter, stderr)

	if out.FunctionError != nil && *out.FunctionError != "" {
		writePayload(stdout, out.Payload, true)
		return fmt.Errorf("function invocation failed (%s)", *out.FunctionError)
	}
	if out.StatusCode >= 300 {
		writePayload(stdout, out.Payload, true)
		return fmt.Errorf("function invocation failed with status code %d", out.StatusCode)
	}

	writePayload(stdout, out.Payload, false)
	fmt.Fprintln(stderr, "Function invocation completed successfully")
	return nil
}

func normalizePayload(payload []byte) ([]byte, error) {
	if len(strings.TrimSpace(string(payload))) == 0 {
		return []byte(EmptyPayload), nil
	}
	if !json.Valid(payload) {
		return nil, fmt.Errorf("invalid --payload, a lambda function requires the payload to be a valid json document")
	}
	return payload, nil
}

// createClientContext passes the nullstone trigger metadata to the function
// Lambda has no mechanism to override environment variables for a single invocation,
// so this metadata is delivered through the client context instead
// (available on the `client_context` field of the function's context object)
func createClientContext(username string) (*string, error) {
	custom := map[string]string{
		admin.TriggerEnvVar:     admin.TriggerManual,
		admin.TriggerNameEnvVar: username,
	}
	raw, err := json.Marshal(map[string]any{"custom": custom})
	if err != nil {
		return nil, fmt.Errorf("unable to construct the client context for this invocation: %w", err)
	}
	encoded := base64.StdEncoding.EncodeToString(raw)
	if len(encoded) > MaxClientContextLength {
		return nil, fmt.Errorf("unable to construct the client context for this invocation: exceeds the maximum length of %d bytes", MaxClientContextLength)
	}
	return aws.String(encoded), nil
}

// emitLogs decodes the tail of the log output that lambda returns with the invocation
// Lambda only returns the last 4KB of logs, use `nullstone logs` to see the entire log output
func emitLogs(infra Outputs, logResult *string, emitter app.LogEmitter, stderr io.Writer) {
	if emitter == nil || logResult == nil || *logResult == "" {
		return
	}
	decoded, err := base64.StdEncoding.DecodeString(*logResult)
	if err != nil {
		fmt.Fprintf(stderr, "Unable to read the logs for this invocation: %s\n", err)
		return
	}

	contents := string(decoded)
	stream := infra.LambdaName
	if matches := requestIdRegex.FindStringSubmatch(contents); len(matches) > 1 {
		stream = matches[1]
	}
	// Lambda emits a START line at the beginning of every invocation
	// If it's missing, we received the tail of a log that was too large to return in full
	if !strings.HasPrefix(contents, "START RequestId:") {
		fmt.Fprintln(stderr, "This invocation produced more logs than lambda returns (4KB); run `nullstone logs` to see the entire log output")
	}

	for _, line := range strings.Split(strings.TrimRight(contents, "\n"), "\n") {
		emitter(app.LogMessage{
			SourceType: logSourceType,
			Source:     infra.LambdaName,
			Stream:     stream,
			Message:    strings.TrimRight(line, "\r"),
		})
	}
}

func writePayload(stdout io.Writer, payload []byte, always bool) {
	contents := strings.TrimSpace(string(payload))
	if contents == "" {
		return
	}
	// A function that returns nothing responds with `null`; that's noise unless we're reporting a failure
	if !always && contents == "null" {
		return
	}
	fmt.Fprintln(stdout, contents)
}
