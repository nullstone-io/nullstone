package all

import (
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
	"gopkg.in/nullstone-io/nullstone.v0/aws/beanstalk"
	"gopkg.in/nullstone-io/nullstone.v0/aws/ec2"
	"gopkg.in/nullstone-io/nullstone.v0/aws/ecs"
	"gopkg.in/nullstone-io/nullstone.v0/gcp/gke"
)

var (
	ecsContract = types.ModuleContractName{
		Category:    string(types.CategoryApp),
		Subcategory: string(types.SubcategoryAppContainer),
		Provider:    "aws",
		Platform:    "ecs",
		Subplatform: "*",
	}
	beanstalkContract = types.ModuleContractName{
		Category:    string(types.CategoryApp),
		Subcategory: string(types.SubcategoryAppServer),
		Provider:    "aws",
		Platform:    "ec2",
		Subplatform: "beanstalk",
	}
	ec2Contract = types.ModuleContractName{
		Category:    string(types.CategoryApp),
		Subcategory: string(types.SubcategoryAppServer),
		Provider:    "aws",
		Platform:    "ec2",
		Subplatform: "",
	}
	lambdaContract = types.ModuleContractName{
		Category:    string(types.CategoryApp),
		Subcategory: string(types.SubcategoryAppServerless),
		Provider:    "aws",
		Platform:    "lambda",
		Subplatform: "*",
	}
	s3SiteContract = types.ModuleContractName{
		Category:    string(types.CategoryApp),
		Subcategory: string(types.SubcategoryAppStaticSite),
		Provider:    "aws",
		Platform:    "s3",
		Subplatform: "",
	}
	gkeContract = types.ModuleContractName{
		Category:    string(types.CategoryApp),
		Subcategory: string(types.SubcategoryAppContainer),
		Provider:    "gcp",
		Platform:    "k8s",
		Subplatform: "gke",
	}

	Providers = admin.Providers{
		ecsContract: admin.Provider{
			NewStatuser: ecs.NewStatuser,
			NewRemoter:  ecs.NewRemoter,
		},
		beanstalkContract: admin.Provider{
			NewStatuser: nil, // TODO: beanstalk.NewStatuser
			NewRemoter:  beanstalk.NewRemoter,
		},
		ec2Contract: admin.Provider{
			NewStatuser: nil,
			NewRemoter:  ec2.NewRemoter,
		},
		lambdaContract: admin.Provider{
			NewStatuser: nil,
			NewRemoter:  nil, // TODO: lambda.NewRemoter,
		},
		s3SiteContract: admin.Provider{
			NewStatuser: nil,
			NewRemoter:  nil,
		},
		gkeContract: admin.Provider{
			NewStatuser: nil,
			NewRemoter:  gke.NewRemoter,
		},
	}
)
