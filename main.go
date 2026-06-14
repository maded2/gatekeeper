// Gatekeeper — Automated Quality Score Tool
//
// Evaluates git commits, pull requests, or full codebases against
// an organizational Quality Score Standard using static analysis
// and LLM-driven reasoning.
package main

import "gatekeeper/cmd"

func main() {
	cmd.Execute()
}
