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
	"maps"
	"net/http"
	"slices"
)

// encryptedValueMask is what APIM answers in place of an encrypted dictionary value (EncryptedValueMask).
const encryptedValueMask = "••••••••••••"

// resolveSecrets checks the manual properties of a dictionary PUT against the stored dictionary and
// replaces each masked value with the stored one, as APIM does. Keys are checked in sorted order.
func resolveSecrets(body Dictionary, stored DictionaryState, exists bool) (Dictionary, *HttpError) {
	if body.Manual == nil {
		return body, nil
	}
	plain := deref(body.Manual.Properties)
	secret := deref(body.Manual.EncryptedProperties)
	storedSecret := map[string]EncryptableValue{}
	if exists && stored.Manual != nil {
		storedSecret = deref(stored.Manual.EncryptedProperties)
	}

	for _, key := range slices.Sorted(maps.Keys(secret)) {
		if _, ok := plain[key]; ok {
			detail := "it is declared in both 'properties' and 'encryptedProperties'"
			return body, badRequest(
				"Options for dictionary property ["+key+"] are invalid: "+detail+".",
				"dictionary.propertyOptions.invalid",
				map[string]string{"property": key, "detail": detail},
			)
		}
	}
	for _, key := range slices.Sorted(maps.Keys(plain)) {
		if _, ok := storedSecret[key]; ok {
			return body, badRequest(
				"Dictionary property ["+key+"] on dictionary ["+deref(stored.Id)+"] is already encrypted; "+
					"it cannot be made plain again. Renew it with a new value instead.",
				"dictionary.property.encryptedToPlain",
				map[string]string{"dictionary": deref(stored.Id), "property": key},
			)
		}
	}

	resolved := make(map[string]EncryptableValue, len(secret))
	for _, key := range slices.Sorted(maps.Keys(secret)) {
		value := secret[key]
		if value.Value == encryptedValueMask {
			storedValue, ok := storedSecret[key]
			if !ok {
				return body, badRequest(
					"Dictionary property ["+key+"] holds the placeholder that stands in for an encrypted value, "+
						"but no encrypted value is stored under that key. Send the real value instead.",
					"dictionary.property.maskedValue",
					map[string]string{"property": key},
				)
			}
			value = storedValue
		}
		resolved[key] = value
	}
	if body.Manual.EncryptedProperties != nil {
		manual := *body.Manual
		manual.EncryptedProperties = &resolved
		body.Manual = &manual
	}
	return body, nil
}

// masked returns d with every encrypted value replaced by the mask: APIM never returns them.
func masked(d DictionaryState) DictionaryState {
	if d.Manual == nil || d.Manual.EncryptedProperties == nil {
		return d
	}
	values := make(map[string]EncryptableValue, len(*d.Manual.EncryptedProperties))
	for key := range *d.Manual.EncryptedProperties {
		values[key] = EncryptableValue{Value: encryptedValueMask}
	}
	manual := *d.Manual
	manual.EncryptedProperties = &values
	d.Manual = &manual
	return d
}

func badRequest(message, technicalCode string, parameters map[string]string) *HttpError {
	return &HttpError{
		HttpStatus:    new(int32(http.StatusBadRequest)),
		Message:       new(message),
		TechnicalCode: new(technicalCode),
		Parameters:    &parameters,
	}
}

// deref returns the value p points to, or its zero value.
func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}
