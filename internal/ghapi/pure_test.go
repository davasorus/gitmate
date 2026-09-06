package ghapi

import (
	"testing"
	"time"
)

func TestParseDispatchInputs(t *testing.T) {
	yml := `
on:
  workflow_dispatch:
    inputs:
      version:
        description: "the version"
        required: true
        default: "1.0"
        type: string
      env:
        type: choice
        options: [dev, prod]
jobs:
  build:
    runs-on: ubuntu-latest
`
	inputs, has := parseDispatchInputs(yml)
	if !has {
		t.Fatal("expected hasDispatch true")
	}
	if len(inputs) != 2 {
		t.Fatalf("expected 2 inputs, got %d: %+v", len(inputs), inputs)
	}
	if inputs[0].Name != "version" || !inputs[0].Required || inputs[0].Default != "1.0" {
		t.Errorf("version input wrong: %+v", inputs[0])
	}
	if inputs[1].Type != "choice" || len(inputs[1].Options) != 2 {
		t.Errorf("env choice wrong: %+v", inputs[1])
	}
}

func TestParseDispatchInputs_NoDispatch(t *testing.T) {
	_, has := parseDispatchInputs("on:\n  push:\njobs: {}\n")
	if has {
		t.Fatal("expected hasDispatch false for push-only")
	}
}

func TestParseDispatchInputs_ScalarOn(t *testing.T) {
	_, has := parseDispatchInputs("on: workflow_dispatch\njobs: {}\n")
	if !has {
		t.Fatal("expected hasDispatch true for scalar on: workflow_dispatch")
	}
}

func TestSplitStepLogs(t *testing.T) {
	raw := "##[group]Run checkout\nchecked out\n##[endgroup]\n##[group]Build\ncompiling\ndone\n##[endgroup]\n"
	steps := splitStepLogs(raw)
	if len(steps) != 2 {
		t.Fatalf("expected 2 steps, got %d: %+v", len(steps), steps)
	}
	if steps[0].Name != "Run checkout" {
		t.Errorf("step0 name %q", steps[0].Name)
	}
	if steps[1].Name != "Build" {
		t.Errorf("step1 name %q", steps[1].Name)
	}
}

func TestParseJobGraph(t *testing.T) {
	yml := `
jobs:
  build:
    runs-on: ubuntu-latest
  test:
    needs: build
  deploy:
    needs: [build, test]
`
	nodes := parseJobGraph(yml)
	if len(nodes) != 3 {
		t.Fatalf("expected 3 jobs, got %d", len(nodes))
	}
	byName := map[string][]string{}
	for _, n := range nodes {
		byName[n.Name] = n.Needs
	}
	if len(byName["build"]) != 0 {
		t.Errorf("build should have no needs")
	}
	if len(byName["test"]) != 1 || byName["test"][0] != "build" {
		t.Errorf("test needs wrong: %v", byName["test"])
	}
	if len(byName["deploy"]) != 2 {
		t.Errorf("deploy needs wrong: %v", byName["deploy"])
	}
}

func TestHumanDur(t *testing.T) {
	// via fmtInt indirectly; humanDur takes a time.Duration
	cases := map[int]string{5: "5s", 65: "1m5s", 3661: "61m1s"}
	for secs, want := range cases {
		got := humanDur(durSeconds(secs))
		if got != want {
			t.Errorf("humanDur(%ds) = %q, want %q", secs, got, want)
		}
	}
}

func TestParseRepo(t *testing.T) {
	cases := []struct {
		url, owner, repo string
		wantErr          bool
	}{
		{"https://github.com/davasorus/gitmate.git", "davasorus", "gitmate", false},
		{"https://github.com/o/r", "o", "r", false},
		{"git@github.com:o/r.git", "o", "r", false},
		{"https://gitlab.com/o/r", "", "", true},
	}
	for _, tc := range cases {
		o, r, err := ParseRepo(tc.url)
		if tc.wantErr {
			if err == nil {
				t.Errorf("%s: expected error", tc.url)
			}
			continue
		}
		if err != nil || o != tc.owner || r != tc.repo {
			t.Errorf("%s: got (%q,%q,%v)", tc.url, o, r, err)
		}
	}
}

func durSeconds(s int) time.Duration { return time.Duration(s) * time.Second }
