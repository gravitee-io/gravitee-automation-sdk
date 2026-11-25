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
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Confirm asks a yes/no question and returns true only on "y" or "yes".
// Anything else, including an empty answer or a read error, means no.
func Confirm(in io.Reader, out io.Writer, question string) bool {
	_, _ = fmt.Fprintf(out, "%s [y/N] ", question)
	answer, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && answer == "" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}
