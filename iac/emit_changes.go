package iac

import (
	"fmt"
	"io"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/mitchellh/colorstring"
	"github.com/nullstone-io/iac/events"
	"github.com/nullstone-io/iac/workspace"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

// changeTypeNamespace mirrors the (unexported) namespace change type produced by
// workspace.DiffCapabilityConfig. There is no exported constant for it in go-api-client.
const changeTypeNamespace types.ChangeType = "namespace"

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
	} else if change.Identifier == "extra_subdomain" {
		changeType = "dns"
		identifier = ""
	} else if change.ChangeType == changeTypeNamespace {
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
		emitVariableValueDiff(w, indent, prevVar.Value, newVar.Value)
	case types.ChangeTypeEnvVariable:
		prevEnvVar, _ := change.Current.(types.EnvVariable)
		newEnvVar, _ := change.Desired.(types.EnvVariable)
		colorstring.Fprintf(w, "%s[red]%s[reset] => [green]%s[reset]\n", indent, prevEnvVar.Value, newEnvVar.Value)
	case types.ChangeTypeConnection:
		prevConn, _ := change.Current.(types.Connection)
		newConn, _ := change.Desired.(types.Connection)
		colorstring.Fprintf(w, "%s[red]%s[reset] => [green]%s[reset]\n", indent, prevConn.EffectiveTarget, newConn.EffectiveTarget)
	case types.ChangeTypeCapability:
		prevCap, _ := change.Current.(types.CapabilityConfig)
		newCap, _ := change.Desired.(types.CapabilityConfig)
		changes := workspace.DiffCapabilityConfig(prevCap, newCap).ToSlice()
		slices.SortFunc(changes, compareWorkspaceChange)
		for _, subChange := range changes {
			emitChangeLabel(w, indent, subChange, false)
			if subChange.Action == types.ChangeActionUpdate {
				emitUpdateChangeDiff(w, indent, subChange)
			}
		}
	case changeTypeNamespace:
		prevNs, _ := change.Current.(string)
		newNs, _ := change.Desired.(string)
		colorstring.Fprintf(w, "%s[red]%s[reset] => [green]%s[reset]\n", indent, prevNs, newNs)
	case types.ChangeTypeExtraSubdomain:
		emitSubdomainChangeDiff(w, indent, change)
	}
}

func emitModuleUpdateChangeDiff(w io.Writer, indent string, change types.WorkspaceChange) {
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
			if prevVar.Type != newVar.Type {
				colorstring.Fprintf(w, "%s[yellow]~ type[reset]: [red]%s[reset] => [green]%s[reset]\n", indent, prevVar.Type, newVar.Type)
			}
			if prevVar.Sensitive != newVar.Sensitive {
				colorstring.Fprintf(w, "%s[yellow]~ sensitive[reset]: [red]%t[reset] => [green]%t[reset]\n", indent, prevVar.Sensitive, newVar.Sensitive)
			}
			if prevDefault, newDefault := variableValToString(prevVar.Default), variableValToString(newVar.Default); prevDefault != newDefault {
				colorstring.Fprintf(w, "%s[yellow]~ default[reset]: [red]%s[reset] => [green]%s[reset]\n", indent, prevDefault, newDefault)
			}
			if prevVar.Description != newVar.Description {
				colorstring.Fprintf(w, "%s[yellow]~ description[reset]\n", indent)
			}
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

func emitSubdomainChangeDiff(w io.Writer, indent string, change types.WorkspaceChange) {
	prev, _ := change.Current.(*types.ExtraSubdomainConfig)
	if prev == nil {
		prev = &types.ExtraSubdomainConfig{}
	}
	cur, _ := change.Desired.(*types.ExtraSubdomainConfig)
	if cur == nil {
		cur = &types.ExtraSubdomainConfig{}
	}

	emit := func(field string, p, c string) {
		if p == c {
			return
		}
		if p == "" {
			p = "(empty)"
		}
		if c == "" {
			c = "(empty)"
		}
		colorstring.Fprintf(w, "%s[yellow]~ %s[reset]: [red]%s[reset] => [green]%s[reset]\n", indent, field, p, c)
	}

	emit("template", prev.SubdomainNameTemplate, cur.SubdomainNameTemplate)
	emit("template", prev.SubdomainName, cur.SubdomainName)
	emit("template", prev.DomainName, cur.DomainName)
	emit("template", prev.Fqdn, cur.Fqdn)
}

// emitVariableValueDiff renders a diff between two variable values. Maps and lists are
// expanded Terraform-style so the reader sees exactly which items changed, were added, or
// were removed (with unchanged items kept inline for context); scalars render as "old => new".
func emitVariableValueDiff(w io.Writer, indent string, prev, next any) {
	if pm, nm, ok := bothMaps(prev, next); ok {
		colorstring.Fprintf(w, "%s{\n", indent)
		emitMapDiffBody(w, indent+indentStep, pm, nm)
		colorstring.Fprintf(w, "%s}\n", indent)
		return
	}
	if pl, nl, ok := bothLists(prev, next); ok {
		colorstring.Fprintf(w, "%s[\n", indent)
		emitListDiffBody(w, indent+indentStep, pl, nl)
		colorstring.Fprintf(w, "%s]\n", indent)
		return
	}
	colorstring.Fprintf(w, "%s[red]%s[reset] => [green]%s[reset]\n", indent, variableValToString(prev), variableValToString(next))
}

// emitMapDiffBody renders the per-key diff lines for a map (no surrounding braces).
func emitMapDiffBody(w io.Writer, indent string, prev, next map[string]any) {
	for _, k := range sortedUnionKeys(prev, next) {
		pv, pok := prev[k]
		nv, nok := next[k]
		switch {
		case pok && !nok:
			colorstring.Fprintf(w, "%s[red]- %s: %s[reset]\n", indent, k, variableValToString(pv))
		case !pok && nok:
			colorstring.Fprintf(w, "%s[green]+ %s: %s[reset]\n", indent, k, variableValToString(nv))
		case variableValToString(pv) == variableValToString(nv):
			// unchanged - keep as context
			colorstring.Fprintf(w, "%s  %s: %s\n", indent, k, variableValToString(pv))
		default:
			emitChangedItem(w, indent, k+": ", pv, nv)
		}
	}
}

// emitListDiffBody renders the per-index diff lines for a list (no surrounding brackets).
func emitListDiffBody(w io.Writer, indent string, prev, next []any) {
	n := len(prev)
	if len(next) > n {
		n = len(next)
	}
	for i := 0; i < n; i++ {
		switch {
		case i >= len(next):
			colorstring.Fprintf(w, "%s[red]- %s[reset]\n", indent, variableValToString(prev[i]))
		case i >= len(prev):
			colorstring.Fprintf(w, "%s[green]+ %s[reset]\n", indent, variableValToString(next[i]))
		case variableValToString(prev[i]) == variableValToString(next[i]):
			// unchanged - keep as context
			colorstring.Fprintf(w, "%s  %s\n", indent, variableValToString(prev[i]))
		default:
			emitChangedItem(w, indent, "", prev[i], next[i])
		}
	}
}

// emitChangedItem renders a single changed item, recursing into nested maps/lists.
// label is the key prefix ("key: ") for map entries, or empty for list elements.
func emitChangedItem(w io.Writer, indent, label string, prev, next any) {
	if pm, nm, ok := bothMaps(prev, next); ok {
		colorstring.Fprintf(w, "%s[yellow]~ %s{[reset]\n", indent, label)
		emitMapDiffBody(w, indent+indentStep, pm, nm)
		colorstring.Fprintf(w, "%s[yellow]}[reset]\n", indent)
		return
	}
	if pl, nl, ok := bothLists(prev, next); ok {
		colorstring.Fprintf(w, "%s[yellow]~ %s[[reset]\n", indent, label)
		emitListDiffBody(w, indent+indentStep, pl, nl)
		colorstring.Fprintf(w, "%s[yellow]][reset]\n", indent)
		return
	}
	colorstring.Fprintf(w, "%s[yellow]~ %s[reset][red]%s[reset] => [green]%s[reset]\n", indent, label, variableValToString(prev), variableValToString(next))
}

func bothMaps(a, b any) (map[string]any, map[string]any, bool) {
	am, aok := a.(map[string]any)
	bm, bok := b.(map[string]any)
	return am, bm, aok && bok
}

func bothLists(a, b any) ([]any, []any, bool) {
	al, aok := a.([]any)
	bl, bok := b.([]any)
	return al, bl, aok && bok
}

func sortedUnionKeys(a, b map[string]any) []string {
	keys := make([]string, 0, len(a)+len(b))
	for k := range a {
		keys = append(keys, k)
	}
	for k := range b {
		if _, ok := a[k]; !ok {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

func variableValToString(val any) string {
	switch v := val.(type) {
	case nil:
		return ""
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case map[string]any:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(v))
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s: %s", k, variableValToString(v[k])))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			parts = append(parts, variableValToString(item))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	}
	if s, ok := numberToString(val); ok {
		return s
	}
	return fmt.Sprintf("%v", val)
}

// numberToString formats any numeric type without scientific notation or float
// artifacts, normalizing int and float representations of the same value (e.g.
// float64(60) and int(60) both render as "60").
func numberToString(val any) (string, bool) {
	switch v := val.(type) {
	case int:
		return strconv.FormatInt(int64(v), 10), true
	case int8:
		return strconv.FormatInt(int64(v), 10), true
	case int16:
		return strconv.FormatInt(int64(v), 10), true
	case int32:
		return strconv.FormatInt(int64(v), 10), true
	case int64:
		return strconv.FormatInt(v, 10), true
	case uint:
		return strconv.FormatUint(uint64(v), 10), true
	case uint8:
		return strconv.FormatUint(uint64(v), 10), true
	case uint16:
		return strconv.FormatUint(uint64(v), 10), true
	case uint32:
		return strconv.FormatUint(uint64(v), 10), true
	case uint64:
		return strconv.FormatUint(v, 10), true
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 64), true
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), true
	}
	return "", false
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
