// Go support for leveled logs, analogous to https://code.google.com/p/google-glog/
//
// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package glog

import (
	"errors"
	"flag"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

func init() {
	flag.Var(&logging.verbosity, "v", "log level for V logs")
	flag.Var(&logging.vmodule, "vmodule",
		"comma-separated list of pattern=N settings for file-filtered logging")
	flag.Var(&logging.traceLocation, "log_backtrace_at",
		"when logging hits line file:N, emit a stack trace")
	logging.discard = flag.Bool("discard", false, "discard all logs")

	logging.toWriter = true
	logging.writer = os.Stderr

	logging.setVState(0, nil, false)
}

// VGlobal returns the current global verbosity level.
func VGlobal() Level {
	// we can't rely on LoadInt32 here because that returns 0 for some time between
	// a config set transition, so Locck logging.mu
	logging.mu.Lock()
	v := logging.verbosity.get()
	logging.mu.Unlock()
	return v
}

// VModule gets the per-module verosity level.
func VModule() string {
	return logging.vmodule.String() // holds logging.mu
}

// SetVGlobal sets the global verbosity level.
func SetVGlobal(v Level) Level {
	logging.mu.Lock()
	prev, _ := logging.setVState(v, logging.vmodule.filter, false)
	logging.mu.Unlock()
	return prev
}

var errVmoduleSyntax = errors.New("syntax error: expect comma-separated list of filename=N")

// SetVModule sets the per-module verbosity level.
// Syntax: message=2,routing*=1
func SetVModule(value string) (string, error) {
	var filter []modulePat
	for _, pat := range strings.Split(value, ",") {
		if len(pat) == 0 {
			// Empty strings such as from a trailing comma can be ignored.
			continue
		}
		patLev := strings.Split(pat, "=")
		if len(patLev) != 2 || len(patLev[0]) == 0 || len(patLev[1]) == 0 {
			return "", errVmoduleSyntax
		}
		pattern := patLev[0]
		v, err := strconv.Atoi(patLev[1])
		if err != nil {
			return "", errors.New("syntax error: expect comma-separated list of filename=N")
		}
		if v < 0 {
			return "", errors.New("negative value for vmodule level")
		}
		if v == 0 {
			// hack: change 0 from flag to -1 here; this will mean to ignore this file
			v = -1
		}
		// TODO: check syntax of filter?
		filter = append(filter, modulePat{pattern, isLiteral(pattern), Level(v)})
	}
	logging.mu.Lock()
	_, prev := logging.setVState(logging.verbosity, filter, true)
	logging.mu.Unlock()
	return specToString(prev), nil
}

// SetOutput sets the writer for log output. By default this is os.StdErr.
// It returns the writer that was previously set.
func SetOutput(w io.Writer) io.Writer {
	prev := logging.writer
	logging.writer = w
	return prev
}

// SetOnFatalFunc sets a function to be called on glog.Fatal() invocation.
// It allows to run some logic to report or track the error before os.Exit.
// It's expected that the passed function doesn't hang (i.e. doing network request)
// as it will halt the termination.
func SetOnFatalFunc(f func([]byte)) {
	logging.onFatalFunc = f
}

// limitToDuration is inverse of rate.Every() func.
func limitToDuration(in rate.Limit) time.Duration {
	if in == rate.Inf {
		return 0
	}
	seconds := float64(time.Second) * (1 / float64(in))
	return time.Duration(seconds)
}

// getRateLimit gets the rate limit config and converts it.
// it does not take logging.mu.
func getRateLimit() (time.Duration, int) {
	if logging.rateLimiter == nil {
		return 0, 0
	}
	limit := limitToDuration(logging.rateLimiter.Limit())
	burst := logging.rateLimiter.Burst()
	return limit, burst
}

// GetRateLimit returns the seconds and burst size for the current rate limiter.
//
// If no rate limit it setup, it will return zero values.
func GetRateLimit() (time.Duration, int) {
	logging.mu.Lock()
	defer logging.mu.Unlock()
	return getRateLimit()
}

// SetRateLimit sets the rate limit in seconds and burst size.
func SetRateLimit(limit time.Duration, burst int) (time.Duration, int) {
	logging.mu.Lock()
	defer logging.mu.Unlock()
	prevL, prevB := getRateLimit()
	logging.rateLimiter = rate.NewLimiter(rate.Every(limit), burst)
	return prevL, prevB
}
