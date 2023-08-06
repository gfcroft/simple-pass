/*
Copyright © 2023 George Wheatcroft

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"fmt"
	"os"

	"github.com/georgewheatcroft/simple-pass/cmd"
	easyLogFmt "github.com/t-tomalak/logrus-easy-formatter"

	log "github.com/sirupsen/logrus"
)

func getLogLevel() log.Level {
	level, exists := os.LookupEnv("LOG_LEVEL")
	if !exists {
		return log.InfoLevel
	}

	switch level {
	case "DEBUG", "debug", "Debug":
		return log.DebugLevel
	case "INFO", "Info", "info":
		return log.InfoLevel
	case "TRACE", "Trace", "trace":
		return log.TraceLevel
	case "WARN", "Warn", "warn":
		return log.TraceLevel
	case "ERROR", "Error", "error":
		return log.TraceLevel
	default:
		panic(fmt.Sprintf("unrecognised log level given in environment: %v", level))
	}
}

func main() {
	log.SetLevel(getLogLevel())
	log.SetFormatter(&easyLogFmt.Formatter{
		LogFormat: "%msg%\n",
	})
	cmd.Execute(os.Stdout, os.Stderr)
}
