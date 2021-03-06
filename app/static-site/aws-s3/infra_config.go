package aws_s3

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	nsaws "gopkg.in/nullstone-io/nullstone.v0/aws"
	"gopkg.in/nullstone-io/nullstone.v0/contracts/aws-s3-site"
	"log"
	"os"
	"strings"
	"time"
)

// InfraConfig provides a minimal understanding of the infrastructure provisioned for a module type=aws-s3-site
type InfraConfig struct {
	Outputs aws_s3_site.Outputs
}

func (c InfraConfig) Print(logger *log.Logger) {
	logger = log.New(logger.Writer(), "    ", 0)
	logger.Printf("bucket arn: %q\n", c.Outputs.BucketArn)
	logger.Printf("bucket name: %q\n", c.Outputs.BucketName)
	logger.Printf("CDNs: %s\n", strings.Join(c.Outputs.CdnIds, ", "))
}

func (c InfraConfig) UploadArtifact(ctx context.Context, source string, filepaths []string, version string) error {
	logger := log.New(os.Stderr, "", 0)
	uploader := nsaws.S3Uploader{
		BucketName:      c.Outputs.BucketName,
		ObjectDirectory: version,
		OnObjectUpload: func(objectKey string) {
			logger.Println(fmt.Sprintf("Uploaded %s", objectKey))
		},
	}
	return uploader.UploadDir(ctx, nsaws.NewConfig(c.Outputs.Deployer, c.Outputs.Region), source, filepaths)
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

func (c InfraConfig) InvalidateCdnPaths(ctx context.Context, urlPaths []string) error {
	cfClient := nsaws.NewCloudfrontClient(c.Outputs.Deployer, c.Outputs.Region)
	for _, cdnId := range c.Outputs.CdnIds {
		_, err := cfClient.CreateInvalidation(ctx, &cloudfront.CreateInvalidationInput{
			DistributionId: aws.String(cdnId),
			InvalidationBatch: &cftypes.InvalidationBatch{
				CallerReference: aws.String(time.Now().String()),
				Paths: &cftypes.Paths{
					Quantity: aws.Int32(int32(len(urlPaths))),
					Items:    urlPaths,
				},
			},
		})
		if err != nil {
			return fmt.Errorf("error invalidating cdn %s: %w", cdnId, err)
		}
	}
	return nil
}

func (c InfraConfig) replaceOriginPath(cdn *cloudfront.GetDistributionOutput, newOriginPath string) *cftypes.DistributionConfig {
	dc := cdn.Distribution.DistributionConfig
	for i := range dc.Origins.Items {
		dc.Origins.Items[i].OriginPath = aws.String(fmt.Sprintf("/%s", newOriginPath))
	}
	return dc
}
