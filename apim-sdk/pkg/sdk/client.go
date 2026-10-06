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

package sdk

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/apicontext"
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/errors"
)

// APIMClient is the generated Automation API client, sharing one base URL, org/env, auth, and HTTP client.
type APIMClient struct {
	ClientWithResponsesInterface
	server string
	editor RequestEditorFn
}

// NewAPIMClient builds an APIMClient from ac. Trailing slashes are stripped from BaseURL.
// OrgID and EnvID are baked into the server URL (empty becomes "DEFAULT").
// Auth must be exactly one of bearer or basic.
// The HTTP client has no timeout; use WithHTTPClient to set one or any other HTTP setting.
func NewAPIMClient(ac apicontext.APIContext) (*APIMClient, error) {
	baseUrl, err := url.Parse(strings.TrimRight(ac.BaseURL, "/"))
	if err != nil {
		return nil, errors.NewClientError(err)
	}

	editor, err := ac.AuthInterceptor()
	if err != nil {
		return nil, err
	}

	server, err := NewScopedServerURL(
		ScopedServerURLBaseUrlVariable(baseUrl.String()),
		ScopedServerURLEnvIdVariable(ac.GetEnvIdOrDefault()),
		ScopedServerURLOrgIdVariable(ac.GetOrgIdOrDefault()),
	)
	if err != nil {
		return nil, err
	}

	return (&APIMClient{server: server, editor: editor}).WithHTTPClient(&http.Client{})
}

// WithHTTPClient returns a new APIMClient with the same server URL and auth, sending requests through httpClient.
// The receiver is left unchanged.
func (c *APIMClient) WithHTTPClient(httpClient *http.Client) (*APIMClient, error) {
	client, err := NewClientWithResponses(
		c.server,
		WithHTTPClient(httpClient),
		WithRequestEditorFn(c.editor),
	)
	if err != nil {
		return nil, err
	}
	return &APIMClient{ClientWithResponsesInterface: client, server: c.server, editor: c.editor}, nil
}
