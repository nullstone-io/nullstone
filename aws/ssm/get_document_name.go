package ssm

import "github.com/aws/aws-sdk-go-v2/aws"

var (
	DocNameStartSSH                = "AWS-StartSSHSession"
	DocNamePortForward             = "AWS-StartPortForwardingSession"
	DocNamePortForwardToRemoteHost = "AWS-StartPortForwardingSessionToRemoteHost"
)

func GetDocumentName(parameters map[string][]string) *string {
	if _, forward := parameters["portNumber"]; forward {
		if _, remote := parameters["host"]; remote {
			return aws.String(DocNamePortForwardToRemoteHost)
		}
		return aws.String(DocNamePortForward)
	}
	return aws.String(DocNameStartSSH)
}
