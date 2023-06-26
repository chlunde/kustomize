// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// The kustomize CLI.
package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"

	"sigs.k8s.io/kustomize/kustomize/v5/commands"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	}

	if err := commands.NewDefaultCommand().Execute(); err != nil {
		if *cpuprofile != "" {
			pprof.StopCPUProfile()
		}
		os.Exit(1)
	}
	if *cpuprofile != "" {
		pprof.StopCPUProfile()
	}
	os.Exit(0)
}
