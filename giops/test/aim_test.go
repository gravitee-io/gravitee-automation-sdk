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

package test

import "testing"

func TestAIMCheck(t *testing.T) {
	newGammaCLI(t, "aim").mustRun("aim", "check")
}

func TestAIMMcpProxy(t *testing.T) {
	testGammaResource(t, gammaResource{
		module:   "aim",
		resource: "mcp-proxies",
		file:     "aim/mcp-proxy.yaml",
		hrid:     "giops-mcp-proxy",
		field:    "name",
		value:    "Giops MCP proxy",
		header:   `HRID\s+ENTITYID\s+MODE\s+CONTEXTPATH\s+STATE`,
		row:      `giops-mcp-proxy\s+mcp-proxy.giops\s+PROXY\s+/giops-mcp`,
	})
}

func TestAIMMcpServer(t *testing.T) {
	testGammaResource(t, gammaResource{
		module:   "aim",
		resource: "mcp-servers",
		file:     "aim/mcp-server.yaml",
		hrid:     "giops-mcp-server",
		field:    "entityId",
		value:    "mcp-server.giops",
		header:   `HRID\s+ENTITYID\s+ENDPOINT\s+LASTSYNCEDAT`,
		row:      `giops-mcp-server\s+mcp-server.giops\s+https://mcp.example.com/mcp`,
	})
}
