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

package resources

import (
	"fmt"
	"reflect"

	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/response"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
)

// HandleResponse turns an SDK "…WithResponse" call into what a GetFunc returns: the call's error, an error for
// a status outside 2xx, or a pointer to the JSON200 body. Pass the call's two results as they are:
// HandleResponse(client.GetDomainWithResponse(ctx, key)).
func HandleResponse(response response.BodyResponse, err error) (any, error) {
	if err != nil {
		return nil, err
	}
	if err := core.ExpectSuccess(response.StatusCode(), response.GetBody()); err != nil {
		return nil, err
	}
	return body(response)
}

// HandleResponseSlice does what HandleResponse does for a list call, whose JSON200 is a slice: it returns a
// pointer to each element, as a ListFunc does.
func HandleResponseSlice(response response.BodyResponse, err error) ([]any, error) {
	if err != nil {
		return nil, err
	}
	if err := core.ExpectSuccess(response.StatusCode(), response.GetBody()); err != nil {
		return nil, err
	}
	return bodySlice(response)
}

// body returns a pointer to the JSON200 body.
func body(bodyResponse response.BodyResponse) (any, error) {
	json200, err := json200(bodyResponse)
	if err != nil {
		return nil, err
	}
	return json200.Addr().Interface(), nil

}

// bodySlice returns a pointer to each element of the JSON200 slice.
func bodySlice(bodyResponse response.BodyResponse) ([]any, error) {
	json200, err := json200(bodyResponse)
	if err != nil {
		return nil, err
	}
	slice := make([]any, json200.Len())
	for i := range slice {
		slice[i] = json200.Index(i).Addr().Interface()
	}
	return slice, nil

}

// json200 returns the value the JSON200 pointer field of an oapi-codegen response points to.
func json200(bodyResponse response.BodyResponse) (reflect.Value, error) {

	val := reflect.ValueOf(bodyResponse)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	json200Ptr := val.FieldByName("JSON200")
	if !json200Ptr.IsValid() {
		return reflect.Value{}, fmt.Errorf("invalid JSON200 field")
	}
	if json200Ptr.IsNil() {
		return reflect.Value{}, fmt.Errorf("no JSON body in the response")
	}
	return json200Ptr.Elem(), nil
}
