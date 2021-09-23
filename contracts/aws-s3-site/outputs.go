package aws_s3_site

import "gopkg.in/nullstone-io/nullstone.v0/contracts/aws"

type Outputs struct {
	Region     string   `ns:"region"`
	BucketName string   `ns:"bucket_name"`
	BucketArn  string   `ns:"bucket_arn"`
	Deployer   aws.User `ns:"deployer"`
	CdnIds     []string `ns:"cdn_ids"`
}
