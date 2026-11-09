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
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/auth"
)

// FaultsPath is the control endpoint that sets, lists and clears the faults of a running mock.
const FaultsPath = "/_mock/faults"

// Fault makes the mock answer Status with Message to requests matching Method and Path, instead of
// serving them. Times > 0 answers that many requests, then the fault clears itself; 0 keeps it until
// cleared.
type Fault struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Status  int    `json:"status"`
	Message string `json:"message"`
	Times   int    `json:"times,omitempty"`
}

func (f Fault) validate() error {
	if f.Method == "" || f.Path == "" {
		return errors.New("a fault needs a method and a path")
	}
	if f.Status < 400 || f.Status > 599 {
		return fmt.Errorf("fault status %d is not an error status (400-599)", f.Status)
	}
	if f.Times < 0 {
		return fmt.Errorf("fault times %d is negative", f.Times)
	}
	return nil
}

type faultKey struct {
	method string
	path   string
}

func keyOf(method, path string) faultKey {
	return faultKey{method: strings.ToUpper(method), path: path}
}

// Faults is the set of faults a running mock injects. It is safe for concurrent use.
type Faults struct {
	mu     sync.Mutex
	faults map[faultKey]Fault
}

// Set injects fault, replacing any fault on the same method and path. It panics on an invalid fault.
func (f *Faults) Set(fault Fault) {
	if err := fault.validate(); err != nil {
		panic(err)
	}
	fault.Method = strings.ToUpper(fault.Method)
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.faults == nil {
		f.faults = make(map[faultKey]Fault)
	}
	f.faults[keyOf(fault.Method, fault.Path)] = fault
}

// Clear removes the fault on method and path, if any.
func (f *Faults) Clear(method, path string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.faults, keyOf(method, path))
}

// Reset removes every fault.
func (f *Faults) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	clear(f.faults)
}

// List returns the faults in place, ordered by path then method.
func (f *Faults) List() []Fault {
	f.mu.Lock()
	defer f.mu.Unlock()
	list := make([]Fault, 0, len(f.faults))
	for _, fault := range f.faults {
		list = append(list, fault)
	}
	slices.SortFunc(list, func(a, b Fault) int {
		return cmp.Or(strings.Compare(a.Path, b.Path), strings.Compare(a.Method, b.Method))
	})
	return list
}

// match returns the fault answering a request, counting it against Times.
func (f *Faults) match(method, path string) (Fault, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := keyOf(method, path)
	fault, ok := f.faults[key]
	if !ok {
		return Fault{}, false
	}
	if fault.Times > 0 {
		fault.Times--
		if fault.Times == 0 {
			delete(f.faults, key)
		} else {
			f.faults[key] = fault
		}
	}
	return fault, true
}

// injectFaults answers a faulted request with its fault, in the same error body AM sends.
func injectFaults(faults *Faults) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if fault, ok := faults.match(r.Method, r.URL.Path); ok {
				auth.WriteError(w, fault.Status, fault.Message)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// routeFaultsEndpoint serves FaultsPath on router. The endpoint drives the mock: it is not part of AM's API
// and needs no token.
func routeFaultsEndpoint(router chi.Router, faults *Faults) {
	router.Put(FaultsPath, func(w http.ResponseWriter, r *http.Request) {
		var fault Fault
		if err := json.NewDecoder(r.Body).Decode(&fault); err != nil {
			auth.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := fault.validate(); err != nil {
			auth.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		faults.Set(fault)
		w.WriteHeader(http.StatusNoContent)
	})
	router.Get(FaultsPath, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(faults.List())
	})
	router.Delete(FaultsPath, func(w http.ResponseWriter, r *http.Request) {
		method, path := r.URL.Query().Get("method"), r.URL.Query().Get("path")
		switch {
		case method == "" && path == "":
			faults.Reset()
		case method == "" || path == "":
			auth.WriteError(w, http.StatusBadRequest, "method and path clear one fault together; send neither to clear all")
			return
		default:
			faults.Clear(method, path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
