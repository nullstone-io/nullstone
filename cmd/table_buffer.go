package cmd

import (
	"fmt"
	"github.com/ryanuber/columnize"
	"strings"
)

// TableBuffer builds a table of data to display on the terminal
// The TableBuffer guarantees safe merging of rows with potentially different field names
// Example: If a user is migrating an app from container to serverless,
//
//	it's possible that the infrastructure has not fully propagated
type TableBuffer struct {
	Fields   []string
	HasField map[string]bool
	Data     []map[string]interface{}
}

func (b *TableBuffer) AddFields(fields ...string) {
	if b.HasField == nil {
		b.HasField = map[string]bool{}
	}

	for _, field := range fields {
		if _, ok := b.HasField[field]; !ok {
			b.Fields = append(b.Fields, field)
			b.HasField[field] = true
		}
	}
}

func (b *TableBuffer) AddRow(data map[string]interface{}) {
	cur := map[string]interface{}{}
	for k, v := range data {
		b.AddFields(k)
		cur[k] = v
	}
	b.Data = append(b.Data, cur)
}

func (b *TableBuffer) Values() [][]string {
	all := make([][]string, 0)
	for _, arr := range b.Data {
		cur := make([]string, 0)
		for _, field := range b.Fields {
			if val, ok := arr[field]; ok {
				cur = append(cur, fmt.Sprintf("%s", val))
			} else {
				cur = append(cur, "")
			}
		}
		all = append(all, cur)
	}
	return all
}

func (b *TableBuffer) String() string {
	colConfig := columnize.DefaultConfig()
	values := b.Values()
	lines := make([]string, len(values)+1)
	lines[0] = strings.Join(b.Fields, colConfig.Delim)
	for i, row := range values {
		lines[i+1] = strings.Join(row, colConfig.Delim)
	}
	return columnize.Format(lines, colConfig)
}
