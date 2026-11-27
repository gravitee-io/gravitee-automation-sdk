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
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/errors"
)

// APIMClient is the generated APIM Automation API client, sharing one base URL, org/env, auth, and HTTP client.
type APIMClient struct {
	ClientWithResponsesInterface
	baseURL    string
	server     string
	editor     RequestEditorFn
	httpClient *http.Client
}

// NewAPIMClient builds an APIMClient from ac, resolved first (see APIContext.Resolve).
// Trailing slashes are stripped from BaseURL.
// OrgID and EnvID are baked into the server URL (empty becomes "DEFAULT").
// Auth must be exactly one of bearer or basic.
// The HTTP client has no timeout; use WithHTTPClient to set one or any other HTTP setting.
func NewAPIMClient(ac APIContext) (*APIMClient, error) {
	resolved, err := ac.Resolve()
	if err != nil {
		return nil, err
	}

	baseUrl, err := url.Parse(strings.TrimRight(resolved.BaseURL, "/"))
	if err != nil {
		return nil, errors.NewClientError(err)
	}

	editor, err := resolved.AuthInterceptor()
	if err != nil {
		return nil, err
	}

	server, err := NewScopedServerURL(
		ScopedServerURLBaseUrlVariable(baseUrl.String()),
		ScopedServerURLEnvIdVariable(resolved.GetEnvIdOrDefault()),
		ScopedServerURLOrgIdVariable(resolved.GetOrgIdOrDefault()),
	)
	if err != nil {
		return nil, err
	}

	return (&APIMClient{baseURL: baseUrl.String(), server: server, editor: editor}).WithHTTPClient(&http.Client{})
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
	return &APIMClient{
		ClientWithResponsesInterface: client,
		baseURL:                      c.baseURL,
		server:                       c.server,
		editor:                       c.editor,
		httpClient:                   httpClient,
	}, nil
}

// probeHrid names a Group nobody should create: Probe only needs APIM to look it up.
var probeHrid = "gravitee-automation-sdk-probe"

// Probe checks that BaseURL serves the APIM Automation API, and that APIM accepts the credentials, org and env.
// It reads the OpenAPI document (404 means no Automation API at BaseURL), then looks up a Group:
// found, or "HRID not found", both mean access is granted. Any other answer is returned as an errors.HttpError.
// Both requests carry the credentials: APIM serves the document without auth, but a gateway in front of it,
// such as a Gravitee Cloud gate, may not.
func (c *APIMClient) Probe(ctx context.Context) error {
	if err := c.probeOpenAPI(ctx); err != nil {
		return err
	}
	resp, err := c.GetGroupWithResponse(ctx, probeHrid, nil)
	if err != nil {
		return errors.NewClientError(err)
	}
	if resp.StatusCode() == http.StatusOK || isProbeHridNotFound(resp) {
		return nil
	}
	return errors.HttpError{Status: resp.StatusCode(), Body: string(resp.Body)}
}

func (c *APIMClient) probeOpenAPI(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/open-api.yaml", nil)
	if err != nil {
		return errors.NewClientError(err)
	}
	if err := c.editor(ctx, req); err != nil {
		return errors.NewClientError(err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.NewClientError(err)
	}
	defer func() { _ = resp.Body.Close() }()
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return errors.HttpError{
			Status: resp.StatusCode,
			Body:   fmt.Sprintf("no APIM Automation API at %s: GET open-api.yaml returned 404", c.baseURL),
		}
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxProbeBody))
		return errors.HttpError{Status: resp.StatusCode, Body: string(body)}
	}
}

// maxProbeBody bounds the error body Probe keeps: the OpenAPI document URL may answer with a large page.
const maxProbeBody = 4096

// isProbeHridNotFound tells a missing probe Group apart from other 404s, such as an unknown environment.
func isProbeHridNotFound(resp *GetGroupResponse) bool {
	return resp.JSON404 != nil && resp.JSON404.Message != nil && strings.Contains(*resp.JSON404.Message, probeHrid)
}
