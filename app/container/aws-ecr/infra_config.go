package aws_ecr

import (
	aws_ecr "gopkg.in/nullstone-io/nullstone.v0/contracts/aws-ecr"
	"log"
)

// InfraConfig provides a minimal understanding of the infrastructure provisioned for a module type=aws-fargate
type InfraConfig struct {
	Outputs aws_ecr.Outputs
}

func (c InfraConfig) Print(logger *log.Logger) {
	logger = log.New(logger.Writer(), "    ", 0)
	logger.Printf("repository image url: %q\n", c.Outputs.ImageRepoUrl)
}
