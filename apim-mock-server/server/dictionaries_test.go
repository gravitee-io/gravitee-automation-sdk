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
	"net/http"
	"testing"

	"github.com/gravitee-io/gravitee-automation-sdk/apim-sdk/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const dictionaryHrid = "dictionary"

// secretDictionaryBody has the plain properties plain and the encrypted properties secret.
func secretDictionaryBody(plain, secret map[string]string) jsonObject {
	manual := jsonObject{}
	if plain != nil {
		properties := jsonObject{}
		for key, value := range plain {
			properties[key] = value
		}
		manual["properties"] = properties
	}
	if secret != nil {
		encrypted := jsonObject{}
		for key, value := range secret {
			encrypted[key] = jsonObject{"value": value}
		}
		manual["encryptedProperties"] = encrypted
	}
	return jsonObject{"hrid": dictionaryHrid, "name": "Dictionary", "type": "MANUAL", "deployed": false, "manual": manual}
}

func maskedProperties(keys ...string) jsonObject {
	encrypted := jsonObject{}
	for _, key := range keys {
		encrypted[key] = jsonObject{"value": encryptedValueMask}
	}
	return encrypted
}

func encryptedProperties(got jsonObject) any {
	return got["manual"].(jsonObject)["encryptedProperties"]
}

// storedSecret is the plaintext the mock holds for key, which no response shows.
func storedSecret(t *testing.T, mock *MockAPIM, key string) string {
	t.Helper()
	stored, ok := mock.tenant("DEFAULT", "DEFAULT").Dictionaries.Get(dictionaryHrid)
	require.True(t, ok)
	return (*stored.Manual.EncryptedProperties)[key].Value
}

func assertBadRequest(t *testing.T, resp *http.Response, technicalCode string, parameters jsonObject) {
	t.Helper()
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	got := decodeTo[jsonObject](t, resp)
	assert.Equal(t, technicalCode, got["technicalCode"])
	assert.Equal(t, parameters, got["parameters"])
	assert.NotEmpty(t, got["message"])
}

func TestDictionary_EncryptedValuesAreMasked(t *testing.T) {
	mock, srv := createAPIMServer(t)
	url := collectionURL(srv, "/dictionaries")

	put := putJSON(t, url, secretDictionaryBody(map[string]string{"plain": "visible"}, map[string]string{"secret": "s3cr3t"}))

	assert.Equal(t, maskedProperties("secret"), encryptedProperties(put))
	assert.Equal(t, jsonObject{"plain": "visible"}, put["manual"].(jsonObject)["properties"])
	assert.Equal(t, put, getJSON(t, url, dictionaryHrid))
	assert.Equal(t, "s3cr3t", storedSecret(t, mock, "secret"))
}

func TestDictionary_MaskKeepsStoredValue(t *testing.T) {
	mock, srv := createAPIMServer(t)
	url := collectionURL(srv, "/dictionaries")
	putJSON(t, url, secretDictionaryBody(nil, map[string]string{"secret": "s3cr3t"}))

	got := putJSON(t, url, secretDictionaryBody(nil, map[string]string{"secret": encryptedValueMask}))

	assert.Equal(t, maskedProperties("secret"), encryptedProperties(got))
	assert.Equal(t, "s3cr3t", storedSecret(t, mock, "secret"))
}

func TestDictionary_NewValueReplacesStoredValue(t *testing.T) {
	mock, srv := createAPIMServer(t)
	url := collectionURL(srv, "/dictionaries")
	putJSON(t, url, secretDictionaryBody(nil, map[string]string{"secret": "s3cr3t"}))

	putJSON(t, url, secretDictionaryBody(nil, map[string]string{"secret": "rotated"}))

	assert.Equal(t, "rotated", storedSecret(t, mock, "secret"))
}

func TestDictionary_PlainKeyCanBecomeEncrypted(t *testing.T) {
	mock, srv := createAPIMServer(t)
	url := collectionURL(srv, "/dictionaries")
	putJSON(t, url, secretDictionaryBody(map[string]string{"key": "value"}, nil))

	got := putJSON(t, url, secretDictionaryBody(nil, map[string]string{"key": "value"}))

	assert.Equal(t, maskedProperties("key"), encryptedProperties(got))
	assert.Equal(t, "value", storedSecret(t, mock, "key"))
}

func TestDictionary_MaskWithoutStoredValueIsRejected(t *testing.T) {
	_, srv := createAPIMServer(t)
	url := collectionURL(srv, "/dictionaries")

	resp := httpPut(t, url, secretDictionaryBody(nil, map[string]string{"secret": encryptedValueMask}))

	assertBadRequest(t, resp, "dictionary.property.maskedValue", jsonObject{"property": "secret"})
	assertGet404(t, url, dictionaryHrid)
}

func TestDictionary_EncryptedKeyCannotBecomePlain(t *testing.T) {
	_, srv := createAPIMServer(t)
	url := collectionURL(srv, "/dictionaries")
	existing := putJSON(t, url, secretDictionaryBody(nil, map[string]string{"secret": "s3cr3t"}))

	resp := httpPut(t, url, secretDictionaryBody(map[string]string{"secret": "s3cr3t"}, nil))

	assertBadRequest(t, resp, "dictionary.property.encryptedToPlain",
		jsonObject{"dictionary": existing["id"], "property": "secret"})
	assert.Equal(t, existing, getJSON(t, url, dictionaryHrid))
}

func TestDictionary_KeyDeclaredTwiceIsRejected(t *testing.T) {
	_, srv := createAPIMServer(t)
	url := collectionURL(srv, "/dictionaries")

	resp := httpPut(t, url, secretDictionaryBody(map[string]string{"key": "a"}, map[string]string{"key": "b"}))

	assertBadRequest(t, resp, "dictionary.propertyOptions.invalid", jsonObject{
		"property": "key", "detail": "it is declared in both 'properties' and 'encryptedProperties'",
	})
}

func TestDictionary_DryRunRunsTheSameChecks(t *testing.T) {
	_, srv := createAPIMServer(t)
	url := collectionURL(srv, "/dictionaries")

	resp := httpPut(t, url+"?dryRun=true", secretDictionaryBody(nil, map[string]string{"secret": encryptedValueMask}))

	assertBadRequest(t, resp, "dictionary.property.maskedValue", jsonObject{"property": "secret"})
}

func TestDictionary_DryRunMasksAndKeepsStoredValue(t *testing.T) {
	mock, srv := createAPIMServer(t)
	url := collectionURL(srv, "/dictionaries")
	putJSON(t, url, secretDictionaryBody(nil, map[string]string{"secret": "s3cr3t"}))

	got := putJSON(t, url+"?dryRun=true", secretDictionaryBody(nil, map[string]string{"secret": "rotated"}))

	assert.Equal(t, maskedProperties("secret"), encryptedProperties(got))
	assert.Equal(t, "s3cr3t", storedSecret(t, mock, "secret"))
}

func TestDictionary_EmptyEncryptedValueIsInvalid(t *testing.T) {
	_, srv := createAPIMServer(t)

	resp := httpPut(t, collectionURL(srv, "/dictionaries"), secretDictionaryBody(nil, map[string]string{"secret": ""}))
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// A client can PUT back what it read: the SDK round trip keeps the stored value.
func TestDictionary_SDKReadThenWriteKeepsStoredValue(t *testing.T) {
	mock, srv := createAPIMServer(t)
	c := newAPIMClient(t, srv)
	ctx := t.Context()
	putJSON(t, collectionURL(srv, "/dictionaries"), secretDictionaryBody(nil, map[string]string{"secret": "s3cr3t"}))

	resp, err := c.GetDictionaryWithResponse(ctx, dictionaryHrid)
	assertSDKOK(t, resp, err)
	var read sdk.Dictionary
	require.NoError(t, json.Unmarshal(resp.Body, &read), "a State decodes into its spec")
	assert.Equal(t, encryptedValueMask, read.Manual.EncryptedProperties["secret"].Value)

	upsert, err := c.UpsertDictionaryWithResponse(ctx, nil, read)
	assertSDKOK(t, upsert, err)
	assert.Equal(t, "s3cr3t", storedSecret(t, mock, "secret"))
}
