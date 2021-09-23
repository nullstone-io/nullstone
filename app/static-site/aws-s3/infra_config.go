package aws_s3

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gopkg.in/nullstone-io/nullstone.v0/artifacts"
	nsaws "gopkg.in/nullstone-io/nullstone.v0/aws"
	"gopkg.in/nullstone-io/nullstone.v0/contracts/aws-s3-site"
	"log"
	"os"
	"strings"
)

// InfraConfig provides a minimal understanding of the infrastructure provisioned for a module type=aws-fargate
type InfraConfig struct {
	Outputs aws_s3_site.Outputs
}

func (c InfraConfig) Print(logger *log.Logger) {
	logger = log.New(logger.Writer(), "    ", 0)
	logger.Printf("bucket arn: %q\n", c.Outputs.BucketArn)
	logger.Printf("bucket name: %q\n", c.Outputs.BucketName)
	logger.Printf("CDNs: %s\n", strings.Join(c.Outputs.CdnIds, ", "))
}

func (c InfraConfig) UploadArtifact(ctx context.Context, source artifacts.Walker, version string) error {
	s3Client := s3.NewFromConfig(nsaws.NewConfig(c.Outputs.Deployer, c.Outputs.Region))

	err := source.Walk(func(file *os.File) error {

	})

	return err
}

func (c InfraConfig) UpdateCdnVersion(ctx context.Context, version string) error {
	cdns, err := c.GetCdns(ctx)
	if err != nil {
		return err
	}

	cfClient := nsaws.NewCloudfrontClient(c.Outputs.Deployer, c.Outputs.Region)
	for _, cdnRes := range cdns {
		_, err := cfClient.UpdateDistribution(ctx, &cloudfront.UpdateDistributionInput{
			DistributionConfig: c.replaceOriginPath(cdnRes, version),
			Id:                 cdnRes.Distribution.Id,
			IfMatch:            cdnRes.ETag,
		})
		if err != nil {
			return fmt.Errorf("error updating distribution %q: %w", *cdnRes.Distribution.Id, err)
		}
	}

	return err
}

func (c InfraConfig) GetCdns(ctx context.Context) ([]*cloudfront.GetDistributionOutput, error) {
	cfClient := nsaws.NewCloudfrontClient(c.Outputs.Deployer, c.Outputs.Region)
	cdns := make([]*cloudfront.GetDistributionOutput, 0)
	for _, cdnId := range c.Outputs.CdnIds {
		out, err := cfClient.GetDistribution(ctx, &cloudfront.GetDistributionInput{Id: aws.String(cdnId)})
		if err != nil {
			return nil, fmt.Errorf("error getting distribution %q: %w", cdnId, err)
		}
		cdns = append(cdns, out)
	}
	return cdns, nil
}

func (c InfraConfig) replaceOriginPath(cdn *cloudfront.GetDistributionOutput, newOriginPath string) *cftypes.DistributionConfig {
	dc := cdn.Distribution.DistributionConfig
	for i := range dc.Origins.Items {
		dc.Origins.Items[i].OriginPath = aws.String(newOriginPath)
	}
	return dc
}
