// Command llm-wiki is the deterministic engine CLI for portable knowledge
// bundles. It is callable directly, without Claude Code: every command runs the
// same engine and, under --json, emits the same versioned contract envelope
// that skills, hooks, and CI consume (ADR-003).
package main

import "os"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
