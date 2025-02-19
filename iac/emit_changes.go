package iac

import (
	"fmt"
	"github.com/mitchellh/colorstring"
	"github.com/nullstone-io/iac/events"
	"github.com/nullstone-io/iac/workspace"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"io"
	"slices"
)

func emitWorkspaceChanges(w io.Writer, block types.Block, changes workspace.IndexedChanges) {
	if len(changes) == 0 {
		return
	}

	s := "s"
	if len(changes) == 1 {
		s = ""
	}
	indent := indentStep
	colorstring.Fprintf(w, "%s[bold]%s[reset] => %d change%s\n", indent, block.Name, len(changes), s)
	indent += indentStep
	for _, change := range changes {
		emitChangeLabel(w, indent, *change, false)
		if change.Action == types.ChangeActionUpdate {
			emitUpdateChangeDiff(w, indent, *change)
		}
	}
}

func emitChangeLabel(w io.Writer, indent string, change types.WorkspaceChange, inModule bool) {
	changeType := change.ChangeType
	identifier := fmt.Sprintf(".%s", change.Identifier)
	if change.Identifier == types.ChangeIdentifierModuleVersion {
		changeType = "module"
		identifier = ""
	} else if change.ChangeType == types.ChangeTypeCapability {
		if cur, ok := change.Current.(types.CapabilityConfig); ok {
			index := cur.Source
			if cur.Name != "" {
				index = cur.Name
			}
			identifier = fmt.Sprintf("[%s]", index)
		}
		if desired, ok := change.Desired.(types.CapabilityConfig); identifier == "" && ok {
			index := desired.Source
			if desired.Name != "" {
				index = desired.Name
			}
			identifier = fmt.Sprintf("[%s]", index)
		}
	}

	switch change.Action {
	case types.ChangeActionAdd:
		colorstring.Fprintf(w, "%s[green]+ %s%s[reset]\n", indent, changeType, identifier)
	case types.ChangeActionDelete:
		colorstring.Fprintf(w, "%s[red]- %s%s[reset]\n", indent, changeType, identifier)
	case types.ChangeActionUpdate:
		colorstring.Fprintf(w, "%s[yellow]~ %s%s[reset]\n", indent, changeType, identifier)
	}
}

func emitUpdateChangeDiff(w io.Writer, indent string, change types.WorkspaceChange) {
	indent += indentStep
	switch change.ChangeType {
	case types.ChangeTypeModuleVersion:
		prevModuleConfig, _ := change.Current.(types.ModuleConfig)
		newModuleConfig, _ := change.Desired.(types.ModuleConfig)
		changes := workspace.DiffModuleConfig(prevModuleConfig, newModuleConfig).ToSlice()
		slices.SortFunc(changes, compareWorkspaceChange)
		for _, subChange := range changes {
			emitModuleUpdateChangeDiff(w, indent, subChange)
		}
	case types.ChangeTypeVariable:
		prevVar, _ := change.Current.(types.Variable)
		newVar, _ := change.Desired.(types.Variable)
		colorstring.Fprintf(w, "%s[red]%s[reset] => [green]%s[reset]\n", indent, prevVar.Value, newVar.Value)
	case types.ChangeTypeEnvVariable:
		prevEnvVar, _ := change.Current.(types.EnvVariable)
		newEnvVar, _ := change.Desired.(types.EnvVariable)
		colorstring.Fprintf(w, "%s[red]%s[reset] => [green]%s[reset]\n", indent, prevEnvVar.Value, newEnvVar.Value)
	case types.ChangeTypeConnection:
		prevConn, _ := change.Current.(types.Connection)
		newConn, _ := change.Desired.(types.Connection)
		colorstring.Fprintf(w, "%s[red]%s[reset] => [green]%s[reset]\n", indent, prevConn.EffectiveTarget, newConn.EffectiveTarget)
	case types.ChangeTypeCapability:
		// TODO: Implement
	}
}

func emitModuleUpdateChangeDiff(w io.Writer, indent string, change types.WorkspaceChange) {
	indent += indentStep

	if change.Action == types.ChangeActionUpdate {
		switch change.ChangeType {
		case types.ChangeTypeModuleVersion:
			prevVersion, _ := change.Current.(string)
			newVersion, _ := change.Desired.(string)
			colorstring.Fprintf(w, "%s[red]%s[reset] => [green]%s[reset]\n", indent, prevVersion, newVersion)
		case types.ChangeTypeVariable:
			emitChangeLabel(w, indent, change, true)
			indent += indentStep
			prevVar, _ := change.Current.(types.Variable)
			newVar, _ := change.Desired.(types.Variable)
			colorstring.Fprintf(w, "%s[yellow]~ type[reset]: [red]%s[reset] => [green]%s[reset]\n", indent, prevVar.Type, newVar.Type)
			colorstring.Fprintf(w, "%s[yellow]~ sensitive[reset]: [red]%t[reset] => [green]%t[reset]\n", indent, prevVar.Sensitive, newVar.Sensitive)
			colorstring.Fprintf(w, "%s[yellow]~ default[reset]: [red]%s[reset] => [green]%s[reset]\n", indent, prevVar.Default, newVar.Default)
		case types.ChangeTypeConnection:
			emitChangeLabel(w, indent, change, true)
			indent += indentStep
			prevConn, _ := change.Current.(types.Connection)
			newConn, _ := change.Desired.(types.Connection)
			colorstring.Fprintf(w, "%s[yellow]~ contract[reset]: [red]%s[reset] => [green]%s[reset]\n", indent, prevConn.Contract, newConn.Contract)
			colorstring.Fprintf(w, "%s[yellow]~ optional[reset]: [red]%t[reset] => [green]%t[reset]\n", indent, prevConn.Optional, newConn.Optional)
		}
	}
}

func compareWorkspaceChange(a, b types.WorkspaceChange) int {
	if a.Identifier == types.ChangeIdentifierModuleVersion {
		return -1
	}
	if b.Identifier == types.ChangeIdentifierModuleVersion {
		return 1
	}
	if a.Identifier == b.Identifier {
		return 0
	}
	if a.Identifier > b.Identifier {
		return 1
	}
	return -1

}

func emitEventChanges(w io.Writer, changes events.Changes) {
	if len(changes) == 0 {
		return
	}
	s := "s"
	if len(changes) == 1 {
		s = ""
	}
	indent := indentStep
	colorstring.Fprintf(w, "%s[bold]events[reset] => %d change%s\n", indent, len(changes), s)
	indent += indentStep
	for _, change := range changes {
		emitEventChangeLabel(w, indent, change)
		if change.Action == events.ChangeActionUpdate {
			// TODO: emitEventUpdateChangeDiff(w, indent, change)
		}
	}
}

func emitEventChangeLabel(w io.Writer, indent string, change events.Change) {
	switch change.Action {
	case events.ChangeActionAdd:
		colorstring.Fprintf(w, "%s[green]+ %s[reset]\n", indent, change.Desired.Name)
	case events.ChangeActionDelete:
		colorstring.Fprintf(w, "%s[red]- %s[reset]\n", indent, change.Current.Name)
	case events.ChangeActionUpdate:
		colorstring.Fprintf(w, "%s[yellow]~ %s[reset]\n", indent, change.Desired.Name)
	}
}
