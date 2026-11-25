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

package core

import (
	"context"
	"fmt"

	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/config"
	"github.com/spf13/cobra"
)

// ConfigFlag is the root flag giving the configuration file.
const ConfigFlag = "config"

// Module is a group of commands, "giops <name> …", that share one API client. Several modules can use the
// same product: portal and aim use the APIM client.
//
// ClientFactory builds that client from the configuration, before any command of the module runs. It returns
// a pointer, such as *sdk.AMClient, that the module's functions read with Client, or an error when the
// configuration has no section for the product.
type Module struct {
	// Name is the module's command: "giops <name> …".
	Name string
	// ClientFactory builds the module client from the configuration. Required.
	ClientFactory func(config config.CLIConfig) (any, error)
}

// clientKey is the context key of the module client.
type clientKey struct{}

// Command returns the module command under root. Before any of its subcommands runs,
// it loads the configuration and puts the module client in the command context.
func (m Module) Command(root *cobra.Command) *cobra.Command {
	cmd := FindOrCreate(root, m.Name)
	cmd.PersistentPreRunE = m.connect
	return cmd
}

func (m Module) connect(cmd *cobra.Command, _ []string) error {
	file, err := cmd.Flags().GetString(ConfigFlag)
	if err != nil {
		return err
	}
	c, err := config.Load(file)
	if err != nil {
		return err
	}
	client, err := m.ClientFactory(c)
	if err != nil {
		return fmt.Errorf("cannot create %s client: %w", m.Name, err)
	}
	cmd.SetContext(context.WithValue(cmd.Context(), clientKey{}, client))
	return nil
}

// Client returns the client the module command put in ctx, as *C. It panics when there is none of that type:
// the module's ClientFactory builds another one.
func Client[C any](ctx context.Context) *C {
	client, ok := ctx.Value(clientKey{}).(*C)
	if !ok {
		panic(fmt.Sprintf("no %T client in the context", client))
	}
	return client
}
