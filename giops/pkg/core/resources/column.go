// Copyright (C) 2015 The Gravitee team (http://gravitee.io)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"
)

// Column is a column of the table get prints: a header and how to read a cell from a resource.
type Column struct {
	header  string
	extract func(item any) (string, bool)
}

// NewColumnFromField reads the struct field named field (the Go name, such as "Hrid") of each resource, a
// struct, or a pointer to one. A nil pointer field is an empty cell. The header is the field name in upper case.
// A missing field, or a resource that is not a struct, is a bug in the columns, so it panics.
func NewColumnFromField(field string) *Column {
	return NewColumnFromExtractor(field, fieldExtractor(field))
}

// NewColumnFromExtractor reads each cell with extractor, which returns the cell and true, or false to leave
// the whole row out of the table. The header is the upper case.
func NewColumnFromExtractor(header string, extractor func(item any) (string, bool)) *Column {
	return &Column{header: strings.ToUpper(header), extract: extractor}
}

// printTable prints one row per resource under the column headers.
// A resource that a column cannot read is left out.
func printTable(w io.Writer, resources []any, columns []*Column) error {
	var rows []string
next:
	for _, resource := range resources {
		cells := make([]string, len(columns))
		for i, column := range columns {
			cell, ok := column.extract(resource)
			if !ok {
				continue next
			}
			cells[i] = cell
		}
		rows = append(rows, strings.Join(cells, "\t"))
	}
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "No resources found")
		return err
	}

	headers := make([]string, len(columns))
	for i, column := range columns {
		headers[i] = column.header
	}
	table := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(table, strings.Join(headers, "\t"))
	for _, row := range rows {
		_, _ = fmt.Fprintln(table, row)
	}
	return table.Flush()
}

// fieldExtractor reads field from a struct or a pointer to one; a nil resource leaves its row out.
func fieldExtractor(field string) func(item any) (string, bool) {
	return func(item any) (string, bool) {
		v := reflect.ValueOf(item)
		if !v.IsValid() || v.IsNil() {
			return "", false
		}
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			panic(fmt.Sprintf("cannot extract field [%s] from type [%s], must be a struct", field, v.Type().String()))
		}
		f := v.FieldByName(field)
		if !f.IsValid() {
			panic(fmt.Sprintf("could not find field [%s] in struct [%s]", field, v.Type().String()))
		}
		if f.Kind() == reflect.Pointer {
			if f.IsNil() {
				return "", true
			}
			f = f.Elem()
		}
		return fmt.Sprint(f.Interface()), true
	}
}
