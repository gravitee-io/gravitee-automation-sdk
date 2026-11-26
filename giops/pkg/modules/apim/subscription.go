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

package apim

import (
	"context"

	apimsdk "github.com/gravitee-io/gravitee-automation-sdk/apim-sdk/pkg/sdk"
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/store"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
	res "github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core/resources"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/modules/apim/internal"
)

// Subscription is the file format of an APIM subscription: the subscription itself, plus the hrid of the API
// it subscribes to, which the API takes in the path. Subscriptions have no dry run in the API: they register
// no DryRun.
type Subscription struct {
	Api string `json:"api"`
	apimsdk.Subscription
}

func SubscriptionApi(object store.Identifiable) string {
	return object.(*Subscription).Api
}

func SubscriptionColumns() []*res.Column {
	return []*res.Column{
		res.NewColumnFromField("Hrid"),
		subscriptionColumn("Api", func(s *apimsdk.SubscriptionState) string {
			if s.ApiHrid == nil {
				return ""
			}
			return *s.ApiHrid
		}),
		subscriptionColumn("Plan", func(s *apimsdk.SubscriptionState) string { return s.PlanHrid }),
		subscriptionColumn("Application", func(s *apimsdk.SubscriptionState) string { return s.ApplicationHrid }),
	}
}

// subscriptionColumn is a column whose header is not a field name, such as API for the field ApiHrid.
func subscriptionColumn(header string, cell func(*apimsdk.SubscriptionState) string) *res.Column {
	return res.NewColumnFromExtractor(header, func(item any) (string, bool) {
		subscription, ok := item.(*apimsdk.SubscriptionState)
		if !ok || subscription == nil {
			return "", false
		}
		return cell(subscription), true
	})
}

func GetSubscription(ctx context.Context, _ core.Module, kind res.Kind, id string) (any, error) {
	api, err := kind.GetParent()
	if err != nil {
		return nil, err
	}
	return res.HandleResponse(internal.Client(ctx).GetApiSubscriptionWithResponse(ctx, api, id, nil))
}

func PutSubscription(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	api, err := kind.GetParent()
	if err != nil {
		return err
	}
	body := kind.GetObject().(*Subscription).Subscription
	response, err := internal.Client(ctx).UpsertApiSubscriptionWithResponse(ctx, api, nil, body)
	if err != nil {
		return err
	}
	return core.ExpectSuccess(response.StatusCode(), response.Body)
}

func DeleteSubscription(ctx context.Context, _ core.Module, kind res.Kind, id string) error {
	api, err := kind.GetParent()
	if err != nil {
		return err
	}
	response, err := internal.Client(ctx).DeleteApiSubscriptionWithResponse(ctx, api, id, nil)
	if err != nil {
		return err
	}
	return core.ExpectSuccess(response.StatusCode(), response.Body)
}
