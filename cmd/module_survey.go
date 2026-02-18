package cmd

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/vcs"
)

type moduleSurvey struct{}

var (
	AllToolNames = []string{
		"terraform",
		"opentofu",
	}
)

func (m *moduleSurvey) Ask(cfg api.Config, defaults *types.ModuleManifest) (*types.ModuleManifest, error) {
	manifest := types.ModuleManifest{}
	if defaults != nil {
		manifest = *defaults
	}
	if manifest.ToolName == "" {
		manifest.ToolName = AllToolNames[0]
	}
	if manifest.SourceUrl == "" {
		// Detect repository URL through Git
		if commitInfo, err := vcs.GetCommitInfo(); err == nil {
			manifest.SourceUrl = commitInfo.Repository.Url
		}
	}

	initialQuestions := []*survey.Question{
		m.questionOrgName(cfg),
		{
			Name:      "Name",
			Validate:  survey.Required,
			Transform: survey.ToLower,
			Prompt: &survey.Input{
				Message: "Module Name:",
				Help:    "A name that is used to uniquely identify the module in Nullstone. (Example: aws-rds-postgres)",
				Default: manifest.Name,
			},
		},
		{
			Name:     "FriendlyName",
			Validate: survey.Required,
			Prompt: &survey.Input{
				Message: "Friendly Name:",
				Help:    "A friendly name is what appears to users in the Nullstone UI. (Example: RDS Postgres)",
				Default: manifest.FriendlyName,
			},
		},
		{
			Name:     "Description",
			Validate: survey.Required,
			Prompt: &survey.Input{
				Message: "Description:",
				Help:    "A description helps users understand what the module does.",
				Default: manifest.Description,
			},
		},
		{
			Name:     "ToolName",
			Validate: survey.Required,
			Prompt: &survey.Select{
				Message: "Tool Name:",
				Options: AllToolNames,
				Default: manifest.ToolName,
			},
		},
		{
			Name: "SourceUrl",
			Prompt: &survey.Input{
				Message: "Repository URL:",
				Help:    "The URL to the source repository (e.g. https://github.com/owner/repo)",
				Default: manifest.SourceUrl,
			},
		},
	}
	if err := survey.Ask(initialQuestions, &manifest); err != nil {
		return nil, err
	}

	// IsPublic
	isPublicPrompt := &survey.Confirm{
		Message: "Make this module available to everybody?",
		Default: manifest.IsPublic,
	}
	if err := survey.AskOne(isPublicPrompt, &manifest.IsPublic); err != nil {
		return nil, err
	}

	// Category
	categoryPrompt := &survey.Select{
		Message: "Category:",
		Options: types.AllCategoryNames,
	}
	if manifest.Category != "" && slices.Contains(types.AllCategoryNames, manifest.Category) {
		categoryPrompt.Default = manifest.Category
	}
	if err := survey.AskOne(categoryPrompt, &manifest.Category); err != nil {
		return nil, err
	}

	// [Optional] Subcategory
	subcategories, _ := types.AllSubcategoryNames[types.CategoryName(manifest.Category)]
	if len(subcategories) > 0 {
		subcategoryPrompt := &survey.Select{
			Message: "Subcategory:",
			Options: subcategories,
		}
		if manifest.Subcategory != "" && slices.Contains(subcategories, manifest.Subcategory) {
			subcategoryPrompt.Default = manifest.Subcategory
		}
		if err := survey.AskOne(subcategoryPrompt, &manifest.Subcategory); err != nil {
			return nil, err
		}
	}
	manifest.Subcategory = strings.ToLower(manifest.Subcategory)

	// App Categories
	if strings.HasPrefix(manifest.Category, "capability") {
		// We are splitting category and subcategory
		// We need to map existing app categories (e.g. app/container) to new format (e.g. container)
		curAppCategories := make([]string, 0)
		for _, ac := range manifest.AppCategories {
			curAppCategories = append(curAppCategories, strings.TrimPrefix(ac, "app/"))
		}

		appSubcategories := types.AllSubcategoryNames[types.CategoryApp]
		// Only capabilities are able to limit their targets to a set of app categories
		appCategoriesPrompt := &survey.MultiSelect{
			Message: "Supported App Category: (select none if all apps are supported)",
			Options: appSubcategories,
			Help:    "This allows you to limit which types of apps are allowed to use this capability module",
			Default: curAppCategories,
		}
		manifest.AppCategories = make([]string, 0)
		if err := survey.AskOne(appCategoriesPrompt, &manifest.AppCategories); err != nil {
			return nil, err
		}
	}

	allProviderTypes := []string{
		"aws",
		"gcp",
	}
	defaultProviderTypes := make([]string, 0)
	for _, cur := range manifest.ProviderTypes {
		if slices.Contains(allProviderTypes, cur) {
			defaultProviderTypes = append(defaultProviderTypes, cur)
		}
	}
	providerTypesPrompt := &survey.MultiSelect{
		Message: "Provider Types:",
		Options: allProviderTypes,
		Default: defaultProviderTypes,
	}
	providerTypes := make([]string, 0)
	if err := survey.AskOne(providerTypesPrompt, &providerTypes); err != nil {
		return nil, err
	}
	manifest.ProviderTypes = providerTypes

	var fullPlatform struct {
		Platform string
	}
	curPlatform := manifest.Platform
	if manifest.Subplatform != "" {
		curPlatform = fmt.Sprintf("%s:%s", curPlatform, manifest.Subplatform)
	}
	finalQuestions := []*survey.Question{
		{
			Name:     "Platform",
			Validate: survey.Required,
			Prompt: &survey.Input{
				Message: "Platform:",
				Help:    "The platform that the module targets. (e.g. ecs:fargate, eks, lambda:container, postgres:rds)",
				Default: curPlatform,
			},
		},
	}
	if err := survey.Ask(finalQuestions, &fullPlatform); err != nil {
		return nil, err
	}
	manifest.Platform = fullPlatform.Platform
	tokens := strings.SplitN(manifest.Platform, ":", 2)
	if len(tokens) > 1 {
		manifest.Platform, manifest.Subplatform = tokens[0], tokens[1]
	}

	return &manifest, nil
}

func (m *moduleSurvey) questionOrgName(cfg api.Config) *survey.Question {
	ctx := context.TODO()
	client := api.Client{Config: cfg}
	orgs, _ := client.Organizations().List(ctx)

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
