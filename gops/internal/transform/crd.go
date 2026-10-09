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

package transform

import (
	"embed"
	"fmt"
	"io/fs"
	"slices"
	"sort"
	"strings"

	"github.com/gravitee-io/gravitee-automation-sdk/apim-sdk/openapi"
	"github.com/gravitee-io/gravitee-automation-sdk/gops/internal/resource"
	"github.com/gravitee-io/gravitee-automation-sdk/gops/internal/schema"
	"gopkg.in/yaml.v3"
)

// GKO CRDs, copied from gravitee-kubernetes-operator/crds/gravitee.io.
//
//go:embed crd/manifests/*.yaml
var manifests embed.FS

// crd maps an Automation API state onto the GKO manifest of the matching kind. The mapping is
// driven by the two schemas rather than by per-kind code: the OpenAPI state schema says which
// properties are status (readOnly), the CRD schema says which properties exist and which
// containers are maps rather than lists.
type crd struct {
	doc  *schema.Document
	crds map[string]*schema.CRD
	opts Options
}

func newCRD(opts Options) (*crd, error) {
	doc, err := schema.ParseOpenAPI(openapi.Spec())
	if err != nil {
		return nil, err
	}
	t := &crd{doc: doc, crds: map[string]*schema.CRD{}, opts: opts}
	entries, err := fs.ReadDir(manifests, "crd/manifests")
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		raw, err := manifests.ReadFile("crd/manifests/" + e.Name())
		if err != nil {
			return nil, err
		}
		c, err := schema.ParseCRD(raw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		t.crds[c.Kind] = c
	}
	return t, nil
}

func (crd) Format() string { return "crd" }
func (crd) Ext() string    { return "yaml" }

// manifest keeps the conventional key order of a Kubernetes object.
type manifest struct {
	APIVersion string         `yaml:"apiVersion"`
	Kind       string         `yaml:"kind"`
	Metadata   map[string]any `yaml:"metadata"`
	Spec       map[string]any `yaml:"spec"`
}

func (t *crd) Transform(r resource.Resource, name string) ([]byte, []string, error) {
	state, err := t.doc.Schema(r.Kind.StateSchema)
	if err != nil {
		return nil, nil, err
	}
	c, ok := t.crds[r.Kind.CRDKind]
	if !ok {
		return nil, nil, fmt.Errorf("no CRD embedded for kind %s", r.Kind.CRDKind)
	}
	m := &mapper{stripIDs: t.opts.StripIDs}
	spec := m.object(r.State, state, c.Spec, "spec")
	out, err := yaml.Marshal(manifest{
		APIVersion: c.APIVersion(),
		Kind:       c.Kind,
		Metadata:   map[string]any{"name": name},
		Spec:       spec,
	})
	if err != nil {
		return nil, nil, err
	}
	return out, m.notes, nil
}

type mapper struct {
	stripIDs bool
	notes    []string
}

// object maps one object node. oas describes what the API returned, crd what the manifest accepts.
func (m *mapper) object(tree map[string]any, oas, crd *schema.Schema, path string) map[string]any {
	out := map[string]any{}
	typed := crd != nil && !crd.PreserveUnknown && len(crd.Properties) > 0
	if typed {
		tree = fold(tree, crd)
	}
	for _, k := range sortedKeys(tree) {
		v := tree[k]
		p := path + "." + k
		oasProp := oas.Property(k)
		if oasProp != nil && oasProp.ReadOnly {
			continue // status, stamped by the API: id, environmentId, organizationId, crossId, errors
		}
		if k == "hrid" && isEmpty(v) {
			continue // not created through the Automation API: GKO derives the hrid from the object name
		}
		if k == "id" && m.stripIDs {
			continue // the user chose a portable manifest over adopting this environment's plans and pages
		}
		if !typed {
			out[k] = v
			continue
		}
		crdProp := crd.Property(k)
		if crdProp == nil {
			m.notes = append(m.notes, fmt.Sprintf("%s: dropped, not in the CRD", p))
			continue
		}
		if s, ok := v.(string); ok && len(crdProp.Enum) > 0 && !slices.Contains(crdProp.Enum, s) {
			m.notes = append(m.notes, fmt.Sprintf("%s: dropped %q, the CRD accepts %s", p, s, strings.Join(crdProp.Enum, ", ")))
			continue
		}
		out[k] = m.value(v, oasProp, crdProp, p)
	}
	return out
}

func (m *mapper) value(v any, oas, crd *schema.Schema, path string) any {
	switch val := v.(type) {
	case []any:
		if crd.IsMap() {
			// The API lists what the CRD keys by name: plans, pages.
			out := map[string]any{}
			for i, item := range val {
				im, ok := item.(map[string]any)
				if !ok {
					out[fmt.Sprintf("item-%d", i+1)] = item
					continue
				}
				key := uniqueKey(out, keyOf(im, i))
				out[key] = m.value(im, oas.Items(), crd.AdditionalProperties, path+"["+key+"]")
			}
			return out
		}
		if crd.IsArray() {
			out := make([]any, 0, len(val))
			for i, item := range val {
				out = append(out, m.value(item, oas.Items(), crd.Items(), fmt.Sprintf("%s[%d]", path, i)))
			}
			return out
		}
		return val
	case map[string]any:
		if crd.IsMap() {
			out := map[string]any{}
			for _, k := range sortedKeys(val) {
				out[k] = m.value(val[k], oas.AdditionalProps(), crd.AdditionalProperties, path+"."+k)
			}
			return out
		}
		return m.object(val, oas, crd, path)
	}
	return v
}

// fold nests the fields of a discriminated object the way the CRD declares them. The Automation API
// writes {type: API_KEY, source: HEADER}; the CRD writes {type: API_KEY, apiKey: {source: HEADER}},
// one sub-object per type value. When the CRD object has a `type` enum and a property named after
// the value (API_KEY → apiKey, OAUTH2_AUTH0 → auth0), the fields the CRD does not declare at this
// level move under that property.
func fold(tree map[string]any, crd *schema.Schema) map[string]any {
	typeProp := crd.Property("type")
	value, ok := tree["type"].(string)
	if typeProp == nil || len(typeProp.Enum) == 0 || !ok {
		return tree
	}
	target := branchFor(value, crd)
	if target == "" {
		return tree
	}
	if _, already := tree[target]; already {
		return tree
	}
	out, nested := map[string]any{}, map[string]any{}
	for k, v := range tree {
		if crd.Property(k) != nil {
			out[k] = v
		} else {
			nested[k] = v
		}
	}
	if len(nested) > 0 {
		out[target] = nested
	}
	return out
}

// branchFor names the CRD property holding the fields of a type value: the property whose name,
// lower-cased and without separators, equals the value or ends it (OAUTH2_AUTH0 → auth0).
func branchFor(value string, crd *schema.Schema) string {
	want := normalize(value)
	best := ""
	for name, prop := range crd.Properties {
		if name == "type" || prop.Type != "object" || len(prop.Properties) == 0 {
			continue
		}
		n := normalize(name)
		if n == want {
			return name
		}
		if strings.HasSuffix(want, n) && len(n) > len(normalize(best)) {
			best = name
		}
	}
	return best
}

func normalize(s string) string {
	return strings.ToLower(strings.NewReplacer("_", "", "-", "").Replace(s))
}

// keyOf picks the map key of a listed item: its name, else its hrid, else its position.
func keyOf(item map[string]any, i int) string {
	if name, ok := item["name"].(string); ok && name != "" {
		return name
	}
	if hrid, ok := item["hrid"].(string); ok && hrid != "" {
		return hrid
	}
	return fmt.Sprintf("item-%d", i+1)
}

func uniqueKey(m map[string]any, key string) string {
	k, n := key, 1
	for {
		if _, taken := m[k]; !taken {
			return k
		}
		n++
		k = fmt.Sprintf("%s-%d", key, n)
	}
}

func isEmpty(v any) bool {
	switch val := v.(type) {
	case nil:
		return true
	case string:
		return val == ""
	}
	return false
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
