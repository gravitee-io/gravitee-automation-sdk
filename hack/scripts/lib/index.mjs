/**
 * Copyright (C) 2015 The Gravitee team (http://gravitee.io)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

const { log } = console;

const newLoggerFn =
  (color) =>
  (...args) =>
    log(color(...args));

// Color loggers
const green = newLoggerFn(chalk.green);
const blue = newLoggerFn(chalk.blue);
const magenta = newLoggerFn(chalk.magenta);
const yellow = newLoggerFn(chalk.yellow);
const red = newLoggerFn(chalk.red);

export const LOG = Object.seal({ green, blue, magenta, yellow, red, log });

// Toggle zx verbosity (print commands before executing them)
export function toggleVerbosity(verbose = false) {
  $.verbose = verbose;
}

export function isEmptyString(str) {
  return !(String(str) === str && str.trim().length > 0);
}
