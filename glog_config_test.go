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
	"bytes"
	"strings"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestGetSetVGlobal(t *testing.T) {
	logging.toWriter = true
	buf := bytes.NewBuffer([]byte{})

	defer SetVGlobal(VGlobal())
	defer SetOutput(SetOutput(buf))

	log := "too verbose for me!"

	SetVGlobal(1)
	V(2).Info(log)
	if got := string(buf.Bytes()); got != "" {
		t.Fatalf("unexpected log written to buffer: %#v", got)
	}

	prev := SetVGlobal(10)
	V(2).Info(log)
	if got := string(buf.Bytes()); !strings.Contains(got, log) {
		t.Fatalf("unexpected log written to buffer: %#v", got)
	}
	if prev != 1 {
		t.Fatalf("unexpected previous value: %#v", prev)
	}
}

func TestGetSetVModule(t *testing.T) {
	logging.toWriter = true
	buf := bytes.NewBuffer([]byte{})

	defer SetVModule(VModule())
	defer SetOutput(SetOutput(buf))

	log := "too verbose for me!"

	_, err := SetVModule("glog_config_test=0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if VModule() != "glog_config_test=0" {
		t.Fatalf("unexpected error: %v", err)
	}

	V(2).Info(log)
	if got := string(buf.Bytes()); got != "" {
		t.Fatalf("unexpected log written to buffer: %#v", got)
	}

	prev, err := SetVModule("glog_config_test=10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	V(2).Info(log)
	if got := string(buf.Bytes()); !strings.Contains(got, log) {
		t.Fatalf("unexpected log written to buffer: %#v", got)
	}
	if prev != "glog_config_test=0" {
		t.Fatalf("unexpected previous vmodule: %#v", prev)
	}

}

func TestLimitToDuration(t *testing.T) {
	m := map[rate.Limit]time.Duration{
		rate.Every(time.Second):      time.Second,
		rate.Every(0):                0,
		rate.Every(time.Hour):        time.Hour,
		rate.Every(time.Millisecond): time.Millisecond,
	}
	for l, d := range m {
		newD := limitToDuration(l)
		if d != newD {
			t.Errorf("limitToDuration(%v) = %d, should be %v", l, newD, d)
		}
	}
}

func TestSetOutput(t *testing.T) {
	logging.toWriter = true
	buf := bytes.NewBuffer([]byte{})
	defer SetOutput(SetOutput(buf))
	log := "log to buffer"
	Info(log)
	if got := string(buf.Bytes()); !strings.Contains(got, log) {
		t.Fatalf("unexpected log written to buffer: %s", got)
	}
}
