package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"log"
	"os"
	"strings"
)

func main() {
	log.Println("Generating CLI docs...")
	cliApp := app.Build()

	filename := "docs/CLI.md"
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("unable to open %s for writing: %v", filename, err)
	}
	defer f.Close()

	f.WriteString("# CLI Docs\n")
	f.WriteString("Cheat sheet and reference for the Nullstone CLI.\n\n")
	f.WriteString("This document contains a list of all the commands available in the Nullstone CLI along with:\n")
	f.WriteString("- descriptions\n")
	f.WriteString("- when to use them\n")
	f.WriteString("- examples\n")
	f.WriteString("- options\n\n")

	for _, command := range cliApp.Commands {
		if len(command.Subcommands) > 0 {
			for _, subcommand := range command.Subcommands {
				outputCommandDocs(f, &command.Name, subcommand)
			}
		} else {
			outputCommandDocs(f, nil, command)
		}
	}
}

func formatUsageText(usageText string) string {
	result := strings.Replace(usageText, "\n", "", -1)
	result = strings.Replace(result, "<", "`", -1)
	result = strings.Replace(result, ">", "`", -1)
	return result
}

func formatFlagName(name string, aliases []string) string {
	if len(aliases) > 0 {
		return fmt.Sprintf("`--%s, -%s`", name, strings.Join(aliases, ", "))
	}
	return fmt.Sprintf("`--%s`", name)
}

func formatRequired(required bool) string {
	if required {
		return "required"
	}
	return ""
}

func outputCommandDescription(f *os.File, name string, description string) {
	if name == "version" {
		f.WriteString("Prints the version of the CLI.\n\n")
	} else {
		f.WriteString(description + "\n\n")
	}
}

func outputCommandUsage(f *os.File, name, usage string) {
	f.WriteString("#### Usage\n")
	f.WriteString("```shell\n")
	if name == "version" {
		f.WriteString("nullstone -v\n")
	} else {
		f.WriteString(fmt.Sprintf("$ %s\n", usage))
	}
	f.WriteString("```\n\n")
}

func outputCommandOptions(f *os.File, flags []cli.Flag) {
	if len(flags) > 0 {
		f.WriteString("#### Options\n")
		f.WriteString("| Option | Description | |\n")
		f.WriteString("| --- | --- | --- |\n")
		for _, flag := range flags {
			if sf, ok := flag.(*cli.StringFlag); ok {
				f.WriteString(fmt.Sprintf("| %s | %s | %s |\n", formatFlagName(sf.Name, sf.Aliases), formatUsageText(sf.Usage), formatRequired(sf.Required)))
			} else if bf, ok := flag.(*cli.BoolFlag); ok {
				f.WriteString(fmt.Sprintf("| %s | %s | %s |\n", formatFlagName(bf.Name, bf.Aliases), formatUsageText(bf.Usage), formatRequired(bf.Required)))
			} else if df, ok := flag.(*cli.DurationFlag); ok {
				f.WriteString(fmt.Sprintf("| %s | %s | %s |\n", formatFlagName(df.Name, df.Aliases), formatUsageText(df.Usage), formatRequired(df.Required)))
			} else if ssf, ok := flag.(*cli.StringSliceFlag); ok {
				f.WriteString(fmt.Sprintf("| %s | %s | %s |\n", formatFlagName(ssf.Name, ssf.Aliases), formatUsageText(ssf.Usage), formatRequired(ssf.Required)))
			} else {
				log.Printf("Skipping flag: %+v\n", flag)
			}
		}
		f.WriteString("\n\n")
	}
}

func outputCommandDocs(f *os.File, prefix *string, command *cli.Command) {
	name := command.Name
	if prefix != nil {
		name = fmt.Sprintf("%s %s", *prefix, name)
	}
	log.Printf("Generating docs for %s", name)
	f.WriteString(fmt.Sprintf("## %s\n", name))
	outputCommandDescription(f, name, command.Description)
	outputCommandUsage(f, name, command.UsageText)
	outputCommandOptions(f, command.Flags)
}
