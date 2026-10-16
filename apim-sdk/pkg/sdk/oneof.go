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
	"encoding/json"
	"fmt"
	"reflect"
)

// Polymorphic schemas (oneOf on `type`) hold one pointer per variant instead of oapi-codegen's opaque
// json.RawMessage union, so their content stays walkable by reflection. On the wire they are the variant
// object, with `type` set from the variant's oneof tag. Exactly one variant is expected to be set.

// OneOfListener is a Listener, typed by its `type`.
type OneOfListener struct {
	Http         *HttpListener         `oneof:"HTTP" json:"http,omitempty"`
	Tcp          *TcpListener          `oneof:"TCP" json:"tcp,omitempty"`
	Subscription *SubscriptionListener `oneof:"SUBSCRIPTION" json:"subscription,omitempty"`
	Kafka        *KafkaListener        `oneof:"KAFKA" json:"kafka,omitempty"`
}

func (o OneOfListener) MarshalJSON() ([]byte, error)     { return marshalOneOf(o) }
func (o *OneOfListener) UnmarshalJSON(data []byte) error { return unmarshalOneOf(data, o) }

// OneOfSelector is a flow Selector, typed by its `type`.
type OneOfSelector struct {
	Http      *HttpSelector      `oneof:"HTTP" json:"http,omitempty"`
	Channel   *ChannelSelector   `oneof:"CHANNEL" json:"channel,omitempty"`
	Condition *ConditionSelector `oneof:"CONDITION" json:"condition,omitempty"`
	Mcp       *McpSelector       `oneof:"MCP" json:"mcp,omitempty"`
}

func (o OneOfSelector) MarshalJSON() ([]byte, error)     { return marshalOneOf(o) }
func (o *OneOfSelector) UnmarshalJSON(data []byte) error { return unmarshalOneOf(data, o) }

// OneOfAimCatalogMcpServerAuth is an AimCatalogMcpServerAuth, typed by its `type`.
type OneOfAimCatalogMcpServerAuth struct {
	None   *AimCatalogMcpServerNoAuth     `oneof:"NONE" json:"none,omitempty"`
	Header *AimCatalogMcpServerHeaderAuth `oneof:"HEADER" json:"header,omitempty"`
	OAuth2 *AimCatalogMcpServerOAuth2Auth `oneof:"OAUTH2" json:"oauth2,omitempty"`
}

func (o OneOfAimCatalogMcpServerAuth) MarshalJSON() ([]byte, error) { return marshalOneOf(o) }
func (o *OneOfAimCatalogMcpServerAuth) UnmarshalJSON(data []byte) error {
	return unmarshalOneOf(data, o)
}

// OneOfAimMcpProxyUpstreamAuth is an AimMcpProxyUpstreamAuth, typed by its `type`.
type OneOfAimMcpProxyUpstreamAuth struct {
	None   *AimMcpProxyNoUpstreamAuth     `oneof:"NONE" json:"none,omitempty"`
	ApiKey *AimMcpProxyApiKeyUpstreamAuth `oneof:"API_KEY" json:"apiKey,omitempty"`
	Bearer *AimMcpProxyBearerUpstreamAuth `oneof:"BEARER" json:"bearer,omitempty"`
	Basic  *AimMcpProxyBasicUpstreamAuth  `oneof:"BASIC" json:"basic,omitempty"`
	OAuth2 *AimMcpProxyOAuth2UpstreamAuth `oneof:"OAUTH2" json:"oauth2,omitempty"`
}

func (o OneOfAimMcpProxyUpstreamAuth) MarshalJSON() ([]byte, error) { return marshalOneOf(o) }
func (o *OneOfAimMcpProxyUpstreamAuth) UnmarshalJSON(data []byte) error {
	return unmarshalOneOf(data, o)
}

// OneOfAimMcpProxyPlanSecurity is an AimMcpProxyPlanSecurity, typed by its `type`.
type OneOfAimMcpProxyPlanSecurity struct {
	KeyLess *AimMcpProxyKeyLessPlanSecurity `oneof:"KEY_LESS" json:"keyLess,omitempty"`
	ApiKey  *AimMcpProxyApiKeyPlanSecurity  `oneof:"API_KEY" json:"apiKey,omitempty"`
	OAuth2  *AimMcpProxyOAuth2PlanSecurity  `oneof:"OAUTH2" json:"oauth2,omitempty"`
}

func (o OneOfAimMcpProxyPlanSecurity) MarshalJSON() ([]byte, error) { return marshalOneOf(o) }
func (o *OneOfAimMcpProxyPlanSecurity) UnmarshalJSON(data []byte) error {
	return unmarshalOneOf(data, o)
}

// OneOfAimMcpProxyFlowSelector is an AimMcpProxyFlowSelector, typed by its `type`.
type OneOfAimMcpProxyFlowSelector struct {
	Mcp       *AimMcpProxyMcpFlowSelector       `oneof:"MCP" json:"mcp,omitempty"`
	Condition *AimMcpProxyConditionFlowSelector `oneof:"CONDITION" json:"condition,omitempty"`
}

func (o OneOfAimMcpProxyFlowSelector) MarshalJSON() ([]byte, error) { return marshalOneOf(o) }
func (o *OneOfAimMcpProxyFlowSelector) UnmarshalJSON(data []byte) error {
	return unmarshalOneOf(data, o)
}

// OneOfAimMcpProxyIdentityProvider is an AimMcpProxyIdentityProvider, typed by its `type`.
type OneOfAimMcpProxyIdentityProvider struct {
	GraviteeAm    *AimMcpProxyGraviteeAmIdentityProvider `oneof:"GRAVITEE_AM" json:"graviteeAm,omitempty"`
	OAuth2Generic *AimMcpProxyGenericIdentityProvider    `oneof:"OAUTH2_GENERIC" json:"oauth2Generic,omitempty"`
	OAuth2Auth0   *AimMcpProxyAuth0IdentityProvider      `oneof:"OAUTH2_AUTH0" json:"oauth2Auth0,omitempty"`
}

func (o OneOfAimMcpProxyIdentityProvider) MarshalJSON() ([]byte, error) { return marshalOneOf(o) }
func (o *OneOfAimMcpProxyIdentityProvider) UnmarshalJSON(data []byte) error {
	return unmarshalOneOf(data, o)
}

// marshalOneOf writes the only non-nil variant of v, with "type" set from that field's oneof tag.
func marshalOneOf(v any) ([]byte, error) {
	rv := reflect.ValueOf(v)
	var data []byte
	var kind string
	for i := range rv.NumField() {
		f := rv.Field(i)
		if f.IsNil() {
			continue
		}
		if kind != "" {
			return nil, fmt.Errorf("%s: more than one variant set", rv.Type().Name())
		}
		kind = rv.Type().Field(i).Tag.Get("oneof")
		var err error
		if data, err = json.Marshal(f.Interface()); err != nil {
			return nil, err
		}
	}
	if kind == "" {
		return []byte("null"), nil
	}
	var object map[string]any
	if err := json.Unmarshal(data, &object); err != nil {
		return nil, err
	}
	object["type"] = kind
	return json.Marshal(object)
}

// unmarshalOneOf reads "type" from data and decodes data into the field whose oneof tag matches it.
func unmarshalOneOf(data []byte, v any) error {
	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &head); err != nil {
		return err
	}
	rv := reflect.ValueOf(v).Elem()
	for i := range rv.NumField() {
		if rv.Type().Field(i).Tag.Get("oneof") == head.Type {
			variant := reflect.New(rv.Field(i).Type().Elem())
			if err := json.Unmarshal(data, variant.Interface()); err != nil {
				return err
			}
			rv.Field(i).Set(variant)
			return nil
		}
	}
	return fmt.Errorf("%s: unknown type %q", rv.Type().Name(), head.Type)
}
