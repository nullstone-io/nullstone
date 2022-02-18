package cmd

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"strings"
)

type moduleSurvey struct{}

func (m *moduleSurvey) Ask(cfg api.Config) (*types.Module, error) {
	module := types.Module{
		Status: types.ModuleStatusDraft,
	}

	initialQuestions := []*survey.Question{
		m.questionOrgName(cfg),
		{
			Name:     "Name",
			Validate: survey.Required,
			Prompt: &survey.Input{
				Message: "Module Name:",
				Help:    "A name that is used to uniquely identify the module in Nullstone. (Example: aws-rds-postgres)",
			},
		},
		{
			Name:     "FriendlyName",
			Validate: survey.Required,
			Prompt: &survey.Input{
				Message: "Friendly Name:",
				Help:    "A friendly name is what appears to users in the Nullstone UI. (Example: RDS Postgres)",
			},
		},
		{
			Name:     "Description",
			Validate: survey.Required,
			Prompt: &survey.Input{
				Message: "Description:",
				Help:    "A description helps users understand what the module does.",
			},
		},
		{
			Name:     "Type",
			Validate: survey.Required,
			Prompt: &survey.Input{
				Message: "Type:",
				Help: `Type is a generic identifier to make connections between modules.
For example, the aws-fargate module needs a network so it defines a connection to a network/aws.
Any module that is defined with the type network/aws can satisfy the aws-fargate needs when launched.
Typically, this looks like <generic-resource>/<provider-platform>.
Examples: subdomain/aws, server/ec2, service/aws-fargate, capability/postgres-access/aws`,
			},
		},
	}
	if err := survey.Ask(initialQuestions, &module); err != nil {
		return nil, err
	}

	// IsPublic
	isPublicPrompt := &survey.Confirm{
		Message: "Make this module available to everybody?",
		Default: false,
	}
	if err := survey.AskOne(isPublicPrompt, &module.IsPublic); err != nil {
		return nil, err
	}

	// Category
	categories := []string{
		string(types.CategoryAppServer),
		string(types.CategoryAppContainer),
		string(types.CategoryAppServerless),
		string(types.CategoryAppStaticSite),
		string(types.CategorySubdomain),
		string(types.CategoryDomain),
		string(types.CategoryCapability),
		string(types.CategoryBlock),
	}
	categoryPrompt := &survey.Select{
		Message: "Category:",
		Options: categories,
	}
	var category string
	if err := survey.AskOne(categoryPrompt, &category); err != nil {
		return nil, err
	}
	module.Category = types.CategoryName(category)
	fmt.Println(category, module.Category)

	// Layer
	layers := []string{
		string(types.LayerPublicEntry),
		string(types.LayerService),
		string(types.LayerDatabase),
		string(types.LayerCluster),
		string(types.LayerNetwork),
	}
	layerPrompt := &survey.Select{
		Message: "Layer:",
		Options: layers,
	}
	var layer string
	if err := survey.AskOne(layerPrompt, &layer); err != nil {
		return nil, err
	}
	module.Layer = types.Layer(layer)

	allProviderTypes := []string{
		"aws",
		"gcp",
	}
	providerTypesPrompt := &survey.MultiSelect{
		Message: "Provider Types:",
		Options: allProviderTypes,
	}
	providerTypes := types.ProviderTypes{}
	if err := survey.AskOne(providerTypesPrompt, &providerTypes); err != nil {
		return nil, err
	}

	return &module, nil
}

func (m *moduleSurvey) questionOrgName(cfg api.Config) *survey.Question {
	client := api.Client{Config: cfg}
	orgs, err := client.Organizations().List()
	fmt.Println(cfg, orgs, err)

	return &survey.Question{
		Name:     "OrgName",
		Validate: survey.Required,
		Prompt: &survey.Input{
			Message: "Which organizations owns this module?",
			Default: cfg.OrgName,
			Suggest: func(toComplete string) []string {
				matched := make([]string, 0)
				for _, org := range orgs {
					if strings.HasPrefix(org.Name, toComplete) {
						matched = append(matched, org.Name)
					}
				}
				return matched
			},
		},
	}
}
