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

// Package config loads the giops configuration from a file and env vars, and decodes resource files.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Load reads the configuration: file when it exists, then a .env file in the current directory, then the env
// vars, which override both. It fails when the result has no module section or an invalid one; the error
// then also says why file was not loaded, if it was not.
func Load(file string) (CLIConfig, error) {

	c := &CLIConfig{}

	// loads the file first, ignore errs for now
	fileErr := fromFile(file, c)

	// override with .env, ignore errors
	_ = godotenv.Load()

	// then env vars, which can also define a section the file does not have
	c.addMissingSections()
	if err := env.Parse(c); err != nil {
		return CLIConfig{}, fmt.Errorf("cannot read the configuration from env vars: %w", err)
	}
	c.dropEmptySections()

	err := validator.New(validator.WithRequiredStructEnabled()).Struct(c)
	if err != nil {
		if fileErr == nil {
			return CLIConfig{}, err
		}
		return CLIConfig{}, fmt.Errorf("%w\nfile '%s' was not loaded: %s", err, file, fileErr.Error())
	}

	return *c, nil
}

// fromFile decodes file into c, and fails when file is missing or is a directory.
func fromFile(file string, c *CLIConfig) error {
	stat, err := os.Stat(file)
	if err != nil {
		return err
	}

	if stat.IsDir() {
		return fmt.Errorf("%s is not a file", file)
	}

	return FileToStruct(file, c)
}

// FileToStruct decodes a YAML (.yaml, .yml) or JSON (.json) file into s, a pointer, using the JSON field
// names of s. Any other extension is refused.
func FileToStruct(file string, s any) error {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	isYaml := strings.HasSuffix(file, ".yaml") || strings.HasSuffix(file, ".yml")
	isJson := strings.HasSuffix(file, ".json")

	if isYaml {
		bytes, err = yamlToJSON(bytes)
		if err != nil {
			return err
		}
	} else if !isJson {
		return fmt.Errorf("%s: unsupported file type, use .yaml, .yml or .json", file)
	}
	return json.Unmarshal(bytes, s)
}

// yamlToJSON lets YAML files use the JSON field names of the SDK types, and inline embedded structs as JSON does.
func yamlToJSON(in []byte) ([]byte, error) {
	var tree any
	if err := yaml.Unmarshal(in, &tree); err != nil {
		return nil, err
	}
	return json.Marshal(tree)
}
