package cmd

import (
	"github.com/AlecAivazis/survey/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/modules"
	"strings"
)

type moduleSurvey struct{}

func (m *moduleSurvey) Ask(cfg api.Config) (*modules.Manifest, error) {
	manifest := modules.Manifest{}

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
	}
	if err := survey.Ask(initialQuestions, &manifest); err != nil {
		return nil, err
	}

	// IsPublic
	isPublicPrompt := &survey.Confirm{
		Message: "Make this module available to everybody?",
		Default: false,
	}
	if err := survey.AskOne(isPublicPrompt, &manifest.IsPublic); err != nil {
		return nil, err
	}

	// Category
	categoryPrompt := &survey.Select{
		Message: "Category:",
		Options: types.AllCategoryNames,
	}
	if err := survey.AskOne(categoryPrompt, &manifest.Category); err != nil {
		return nil, err
	}

	// App Categories
	if strings.HasPrefix(manifest.Category, "capability/") {
		// Only capabilities are able to limit their targets to a set of app categories
		appCategoriesPrompt := &survey.MultiSelect{
			Message: "Supported App Category: (select none if all apps are supported)",
			Options: types.AllAppCategoryNames,
			Help:    "This allows you to limit which types of apps are allowed to use this capability module",
		}
		if err := survey.AskOne(appCategoriesPrompt, &manifest.AppCategories); err != nil {
			return nil, err
		}
	}

	// Layer
	// Attempt to find the layer from the chosen category
	// If ambiguous, the mapping will set layer to "" which means we need to prompt the user
	manifest.Layer = m.mapCategoryToLayer(manifest.Category)
	if manifest.Layer == "" {
		layerPrompt := &survey.Select{
			Message: "Layer:",
			Options: types.AllLayerNames,
		}
		if err := survey.AskOne(layerPrompt, &manifest.Layer); err != nil {
			return nil, err
		}
	}

	// Type
	typePrompt := &survey.Question{
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
	}
	if err := survey.Ask([]*survey.Question{typePrompt}, &manifest.Type); err != nil {
		return nil, err
	}

	allProviderTypes := []string{
		"aws",
		"gcp",
	}
	providerTypesPrompt := &survey.MultiSelect{
		Message: "Provider Types:",
		Options: allProviderTypes,
	}
	if err := survey.AskOne(providerTypesPrompt, &manifest.ProviderTypes); err != nil {
		return nil, err
	}

	return &manifest, nil
}

func (m *moduleSurvey) questionOrgName(cfg api.Config) *survey.Question {
	client := api.Client{Config: cfg}
	orgs, _ := client.Organizations().List()

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

func (m *moduleSurvey) mapCategoryToLayer(category string) string {
	if strings.HasPrefix(category, "app/") {
		return string(types.LayerService)
	} else if strings.HasPrefix(category, "capability/") {
		return string(types.LayerService)
	} else if category == types.CategoryDatastore {
		return string(types.LayerDatabase)
	} else if category == types.CategoryDomain {
		return string(types.LayerPublicEntry)
	} else if category == types.CategorySubdomain {
		return string(types.LayerPublicEntry)
	}
	return ""
}
