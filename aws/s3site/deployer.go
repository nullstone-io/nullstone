package s3site

import (
	"errors"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/deploy"
)

var (
	ErrMustUseSingleSource = errors.New("cannot deploy static site because --dir and --archive flags were specified")
	ErrMustSpecifySource   = errors.New("cannot deploy static site because no source flags were specified (--dir or --archive)")
)

var _ deploy.Deployer = Deployer{}

// Deployer will deploy an app/static-site, moduleType=site/aws-s3 using s3sync
type Deployer struct{}

func (d Deployer) Detect(app *types.Application, workspace *types.Workspace) bool {
	if workspace.Module.Category != types.CategoryAppStaticSite {
		return false
	}
	if workspace.Module.Type != "site/aws-s3" {
		return false
	}
	return true
}

func (d Deployer) Identify(nsConfig api.Config, app *types.Application, workspace *types.Workspace) (deploy.InfraConfig, error) {
	return newInfraConfig(nsConfig, workspace)
}

func (d Deployer) Deploy(app *types.Application, workspace *types.Workspace, config map[string]string, infraConfig interface{}) error {
	ic := infraConfig.(*InfraConfig)

	dir := config["dir"]
	archive := config["archive"]
	if dir != "" && archive != "" {
		return ErrMustUseSingleSource
	}

	if dir != "" {
		return ic.SyncDirectory(dir)
	}
	if archive != "" {
		return ic.SyncArchive(archive)
	}

	return ErrMustSpecifySource
}
