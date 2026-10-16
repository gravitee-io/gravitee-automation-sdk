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
	"fmt"
	"slices"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/apicontext"
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/errors"
)

// APIContext is the APIM client input. With CloudToken, Auth.BearerToken is a Gravitee Cloud token:
// org and env come from its claims, and an empty BaseURL becomes the token's cloud gate.
type APIContext struct {
	apicontext.APIContext
	CloudToken bool
}

const cloudGateURL = "https://%s.cloudgate.gravitee.io/apim/automation"

// Resolve returns the common APIContext to build the client from.
// The cloud token signature is not verified: APIM does it on every request.
func (c APIContext) Resolve() (apicontext.APIContext, error) {
	if !c.CloudToken {
		return c.APIContext, nil
	}
	if c.Auth.BearerToken == nil || c.Auth.BasicAuth != nil {
		return apicontext.APIContext{}, errors.NewClientError(fmt.Errorf("cloud token requires bearer auth only"))
	}

	claims, err := cloudTokenClaims(*c.Auth.BearerToken)
	if err != nil {
		return apicontext.APIContext{}, errors.NewClientError(err)
	}

	resolved := c.APIContext
	if resolved.BaseURL == "" {
		resolved.BaseURL = fmt.Sprintf(cloudGateURL, claims.cpg)
	}
	resolved.OrgID = claims.org
	if resolved.EnvID, err = cloudEnv(c.EnvID, claims.envs); err != nil {
		return apicontext.APIContext{}, errors.NewClientError(err)
	}
	return resolved, nil
}

type cloudClaims struct {
	org  string
	envs []string
	cpg  string
}

func cloudTokenClaims(token string) (cloudClaims, error) {
	mapClaims := jwt.MapClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(token, mapClaims); err != nil {
		return cloudClaims{}, fmt.Errorf("invalid cloud token: %w", err)
	}

	org, _ := mapClaims["org"].(string)
	cpg, _ := mapClaims["cpg"].(string)
	var envs []string
	if values, ok := mapClaims["envs"].([]any); ok {
		for _, value := range values {
			if env, ok := value.(string); ok {
				envs = append(envs, env)
			}
		}
	}

	switch {
	case org == "":
		return cloudClaims{}, fmt.Errorf("cloud token has no org claim")
	case len(envs) == 0:
		return cloudClaims{}, fmt.Errorf("cloud token has no envs claim")
	case cpg == "":
		return cloudClaims{}, fmt.Errorf("cloud token has no cpg claim")
	}
	return cloudClaims{org: org, envs: envs, cpg: cpg}, nil
}

// cloudEnv returns the env to use: the given one if the token grants it, else the token's only env.
func cloudEnv(env string, envs []string) (string, error) {
	if env == "" {
		if len(envs) != 1 {
			return "", fmt.Errorf("cloud token grants several envs %v: set EnvID", envs)
		}
		return envs[0], nil
	}
	if !slices.Contains(envs, env) {
		return "", fmt.Errorf("cloud token does not grant env %q (envs %v)", env, envs)
	}
	return env, nil
}
