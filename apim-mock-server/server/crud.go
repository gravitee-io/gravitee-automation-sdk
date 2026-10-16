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

package server

import (
	"encoding/json"
	"maps"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/store"
)

const dryRunMessage = "Mock server is in dry-run-reject mode, all PUT request are rejected on purpose when ?dryRun=true"

// scope locates a resource: its tenant, its kind and, for a sub-resource, its parent.
type scope struct {
	org, env, kind string
	// parentKey is the status field naming the parent (apiHrid, portalHrid), parent its hrid.
	parentKey, parent string
}

func at(org, env, kind string) scope {
	return scope{org: org, env: env, kind: kind}
}

func (s scope) under(parentKey, parent string) scope {
	s.parentKey, s.parent = parentKey, parent
	return s
}

// id is stable per org, env, kind, parent and hrid, like APIM's HRIDToUUID (but not byte-identical to it).
func (s scope) id(hrid string) string {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(s.org+"/"+s.env+"/"+s.kind+"/"+s.parent+"/"+hrid)).String()
}

// crossID is the given crossId (legacy resources set it on PUT), or one derived like id.
func (s scope) crossID(hrid string, given *string) string {
	if given != nil {
		return *given
	}
	s.kind = "cross-" + s.kind
	return s.id(hrid)
}

// status returns the read-only fields APIM computes on the State of hrid. A dry run in reject mode
// also gets the rejection in errors.severe.
func (m *MockAPIM) status(at scope, hrid string, dryRun *bool) map[string]any {
	status := map[string]any{"id": at.id(hrid), "organizationId": at.org, "environmentId": at.env}
	if at.parentKey != "" {
		status[at.parentKey] = at.parent
	}
	if isDryRun(dryRun) && m.dryRunReject {
		status["errors"] = map[string]any{"severe": []string{dryRunMessage}}
	}
	return status
}

// state builds a State from a spec, its defaults already applied, and its status.
// A JSON round trip is enough: every State is allOf(Spec, Status).
func state[T any](spec any, status map[string]any) T {
	var object map[string]any
	mustRoundTrip(spec, &object)
	maps.Copy(object, status)
	var s T
	mustRoundTrip(object, &s)
	return s
}

// mustRoundTrip panics on failure: generated models always encode, and decode what they encoded.
func mustRoundTrip(from, to any) {
	data, err := json.Marshal(from)
	if err == nil {
		err = json.Unmarshal(data, to)
	}
	if err != nil {
		panic(err)
	}
}

// save stores v, unless the request is a dry run: APIM then persists nothing.
func save[T store.Identifiable](s *store.Store[T], v T, dryRun *bool) {
	if !isDryRun(dryRun) {
		s.Put(v)
	}
}

// remove deletes key and reports whether it existed.
func remove[T store.Identifiable](s *store.Store[T], key string) bool {
	_, ok := s.Get(key)
	s.DeleteByKey(key)
	return ok
}

func isDryRun(dryRun *bool) bool {
	return dryRun != nil && *dryRun
}

// startingAt keeps the stored subscription start, otherwise it starts now.
func startingAt(stored Child[SubscriptionState], ok bool) string {
	if ok && stored.Self.StartingAt != nil {
		return *stored.Self.StartingAt
	}
	return time.Now().UTC().Format(time.RFC3339)
}

func hridNotFoundMessage(hrid string) string {
	return "HRID [" + hrid + "] cannot be found."
}

func hridNotFound(hrid string) HttpError {
	return HttpError{HttpStatus: new(int32(http.StatusNotFound)), Message: new(hridNotFoundMessage(hrid))}
}

func aimHridNotFound(hrid string) AimHttpError {
	return AimHttpError{HttpStatus: new(int32(http.StatusNotFound)), Message: new(hridNotFoundMessage(hrid))}
}

// Child is a sub-resource, keyed under its parent's hrid.
type Child[T store.Identifiable] struct {
	ParentIdentity string
	Self           T
}

func (c Child[T]) Identity() string {
	return c.ParentIdentity + "/" + c.Self.Identity()
}

func newChild[T store.Identifiable](self T, parentIdentity string) Child[T] {
	return Child[T]{ParentIdentity: parentIdentity, Self: self}
}

func isChildOf[T store.Identifiable](s *store.Store[Child[T]], parentIdentity, hrid string) (Child[T], bool) {
	return s.Get(parentIdentity + "/" + hrid)
}

func removeChild[T store.Identifiable](s *store.Store[Child[T]], parentIdentity, hrid string) bool {
	return remove(s, parentIdentity+"/"+hrid)
}

func removeChildren[T store.Identifiable](s *store.Store[Child[T]], parentIdentity string) {
	for _, child := range s.GetAll() {
		if child.ParentIdentity == parentIdentity {
			s.Delete(child)
		}
	}
}
