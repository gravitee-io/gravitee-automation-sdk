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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// APIM defaults these required fields when absent and rejects an empty value: an unset one is left out, a set one kept.
func TestRequiredFieldsAPIMDefaults_Omitempty(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name:  "HTTP selector without path and operator",
			value: OneOfSelector{Http: &HttpSelector{Type: "HTTP"}},
			want:  `{"type":"HTTP"}`,
		},
		{
			name:  "HTTP selector with path and operator",
			value: OneOfSelector{Http: &HttpSelector{Type: "HTTP", Path: "/a", PathOperator: "EQUALS"}},
			want:  `{"type":"HTTP","path":"/a","pathOperator":"EQUALS"}`,
		},
		{
			name:  "channel selector without channel and operator",
			value: OneOfSelector{Channel: &ChannelSelector{Type: "CHANNEL"}},
			want:  `{"type":"CHANNEL"}`,
		},
		{
			name:  "channel selector with channel and operator",
			value: OneOfSelector{Channel: &ChannelSelector{Type: "CHANNEL", Channel: "/a", ChannelOperator: "EQUALS"}},
			want:  `{"type":"CHANNEL","channel":"/a","channelOperator":"EQUALS"}`,
		},
		{
			name:  "listener path without path",
			value: PathV4{Host: new("api.example.com")},
			want:  `{"host":"api.example.com"}`,
		},
		{
			name:  "listener path with path",
			value: PathV4{Path: "/orders"},
			want:  `{"path":"/orders"}`,
		},
		{
			name:  "response template without status",
			value: ResponseTemplate{Body: new("oops")},
			want:  `{"body":"oops"}`,
		},
		{
			name:  "response template with status",
			value: ResponseTemplate{Status: 503},
			want:  `{"status":503}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.value)
			require.NoError(t, err)
			assert.JSONEq(t, tt.want, string(got))
		})
	}
}
