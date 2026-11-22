package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"log"
	"os"
	"regexp"
	"strings"
)

// vueInterpolation matches Mustache-style "{{ ... }}" sequences that the
// VitePress compiler would otherwise try to evaluate as Vue interpolation.
var vueInterpolation = regexp.MustCompile(`\{\{[^}]*\}\}`)

// escapeVueInterpolation wraps any "{{ ... }}" sequences in <code v-pre>...</code>
// so the docs site renders them literally instead of failing to compile them.
func escapeVueInterpolation(s string) string {
	return vueInterpolation.ReplaceAllString(s, "<code v-pre>$0</code>")
}

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
		outputCommandTree(f, "", command)
	}
}

// outputCommandTree documents a command and everything below it. A command group with
// subcommands is documented too -- it carries its own description -- so nested groups
// like `envs apps` don't lose their leaves.
func outputCommandTree(f *os.File, prefix string, command *cli.Command) {
	name := command.Name
	if prefix != "" {
		name = fmt.Sprintf("%s %s", prefix, command.Name)
	}
	if len(command.Subcommands) == 0 {
		outputCommandDocs(f, name, command)
		return
	}
	// A top-level group with no description of its own adds nothing but a stub heading.
	if prefix != "" || command.Description != "" {
		outputCommandDocs(f, name, command)
	}
	for _, subcommand := range command.Subcommands {
		outputCommandTree(f, name, subcommand)
	}
}

func formatUsageText(usageText string) string {
	result := strings.Replace(usageText, "\n", "", -1)
	result = strings.Replace(result, "<", "`", -1)
	result = strings.Replace(result, ">", "`", -1)
	return escapeVueInterpolation(result)
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

// formatFlagRow renders one row of a command's options table.
// It reads the flag through the interfaces every flag type of the cli library implements,
// so a flag type that is new to this CLI (int, generic, ...) is documented without a change here.
func formatFlagRow(flag cli.Flag) (string, bool) {
	doc, ok := flag.(cli.DocGenerationFlag)
	if !ok {
		return "", false
	}
	names := flag.Names()
	if len(names) == 0 {
		return "", false
	}
	required := false
	if rf, ok := flag.(cli.RequiredFlag); ok {
		required = rf.IsRequired()
	}
	return fmt.Sprintf("| %s | %s | %s |\n", formatFlagName(names[0], names[1:]), formatUsageText(doc.GetUsage()), formatRequired(required)), true
}

func outputCommandDescription(f *os.File, name string, description string) {
	if name == "version" {
		f.WriteString("Prints the version of the CLI.\n\n")
	} else {
		f.WriteString(escapeVueInterpolation(description) + "\n\n")
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
			row, ok := formatFlagRow(flag)
			if !ok {
				log.Printf("Skipping flag: %+v\n", flag)
				continue
			}
			f.WriteString(row)
		}
		f.WriteString("\n\n")
	}
}

func outputCommandDocs(f *os.File, name string, command *cli.Command) {
	log.Printf("Generating docs for %s", name)
	f.WriteString(fmt.Sprintf("## %s\n", name))
	outputCommandDescription(f, name, command.Description)
	outputCommandUsage(f, name, command.UsageText)
	outputCommandOptions(f, command.Flags)
}
