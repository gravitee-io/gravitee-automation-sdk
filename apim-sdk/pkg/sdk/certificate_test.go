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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	past   = time.Now().Add(-time.Hour)
	future = time.Now().Add(time.Hour)
	plus2  = time.FixedZone("UTC+2", 2*60*60)
)

func TestClientCertificateExpired(t *testing.T) {
	tests := []struct {
		name   string
		endsAt *string
		want   bool
	}{
		{"unset", nil, false},
		{"empty", new(""), false},
		{"garbage", new("not-a-date"), false},
		{"date only", new(past.Format(time.DateOnly)), false},
		{"past RFC3339", new(past.UTC().Format(time.RFC3339)), true},
		{"past RFC3339 with offset", new(past.In(plus2).Format(time.RFC3339)), true},
		{"past RFC3339Nano", new(past.UTC().Format(time.RFC3339Nano)), true},
		{"future RFC3339", new(future.UTC().Format(time.RFC3339)), false},
		{"future RFC3339 with offset", new(future.In(plus2).Format(time.RFC3339)), false},
		{"future RFC3339Nano", new(future.UTC().Format(time.RFC3339Nano)), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ClientCertificate{EndsAt: tt.endsAt}.Expired())
		})
	}
}

func TestClientCertificateScheduled(t *testing.T) {
	tests := []struct {
		name     string
		startsAt *string
		want     bool
	}{
		{"unset", nil, false},
		{"empty", new(""), false},
		{"garbage", new("not-a-date"), false},
		{"date only", new(future.Format(time.DateOnly)), false},
		{"future RFC3339", new(future.UTC().Format(time.RFC3339)), true},
		{"future RFC3339 with offset", new(future.In(plus2).Format(time.RFC3339)), true},
		{"future RFC3339Nano", new(future.UTC().Format(time.RFC3339Nano)), true},
		{"past RFC3339", new(past.UTC().Format(time.RFC3339)), false},
		{"past RFC3339 with offset", new(past.In(plus2).Format(time.RFC3339)), false},
		{"past RFC3339Nano", new(past.UTC().Format(time.RFC3339Nano)), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ClientCertificate{StartsAt: tt.startsAt}.Scheduled())
		})
	}
}
