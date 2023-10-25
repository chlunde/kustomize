// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	. "sigs.k8s.io/kustomize/kustomize/v5/commands/build"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type GenConfig struct {
	fileResources     int
	resources         int
	patches           int
	namespaced        bool
	namePrefix        string
	nameSuffix        string
	commonLabels      map[string]string
	commonAnnotations map[string]string
}

var genConfig = []GenConfig{
	{
		resources:  4,
		namePrefix: "foo-",
		nameSuffix: "-bar",
		commonLabels: map[string]string{
			"foo": "bar",
		},
		commonAnnotations: map[string]string{
			"baz": "blatti",
		},
	},
	{
		resources: 100,
	},
	{
		resources: 3,
	},
	{
		resources:     2,
		namespaced:    true,
		fileResources: 30,
		patches:       10,
	},
	{
		fileResources: 2,
	},
}

func makeKustomization(fSys filesys.FileSystem, path, id string, depth int) error {
	cfg := genConfig[depth]
	fSys.MkdirAll(path)

	var buf bytes.Buffer
	if cfg.namespaced {
		fmt.Fprintf(&buf, "namespace: %s\n", id)
	}

	if cfg.namePrefix != "" {
		fmt.Fprintf(&buf, "namePrefix: %s\n", cfg.namePrefix)
	}

	if cfg.nameSuffix != "" {
		fmt.Fprintf(&buf, "nameSuffix: %s\n", cfg.nameSuffix)
	}

	if len(cfg.commonLabels) > 0 {
		fmt.Fprintf(&buf, "commonLabels:\n")
		for k, v := range cfg.commonLabels {
			fmt.Fprintf(&buf, "  %s: %s\n", k, v)
		}
	}

	if len(cfg.commonAnnotations) > 0 {
		fmt.Fprintf(&buf, "commonAnnotations:\n")
		for k, v := range cfg.commonAnnotations {
			fmt.Fprintf(&buf, "  %s: %s\n", k, v)
		}
	}

	if cfg.fileResources > 0 || cfg.resources > 0 {
		fmt.Fprintf(&buf, "resources:\n")
		for res := 0; res < cfg.fileResources; res++ {
			fn := fmt.Sprintf("res%d.yaml", res)
			fmt.Fprintf(&buf, " - %v\n", fn)

			buf := fmt.Sprintf(`kind: ConfigMap
apiVersion: v1
metadata:
  name: %s-%d
  labels:
    foo: bar
  annotations:
    baz: blatti
data:
  k: v
`, id, res)
			fSys.WriteFile(filepath.Join(path, fn), []byte(buf))
		}

		for res := 0; res < cfg.resources; res++ {
			fn := fmt.Sprintf("res%d", res)
			fmt.Fprintf(&buf, " - %v\n", fn)
			if err := makeKustomization(fSys, path+"/"+fn, fmt.Sprintf("%s-%d", id, res), depth+1); err != nil {
				return err
			}
		}
	}

	if cfg.patches > 0 {
		fmt.Fprintf(&buf, "patches:\n")
		for res := 0; res < cfg.patches; res++ {
			// alternate between json and yaml patches to test both kinds
			if res%2 == 0 {
				fn := fmt.Sprintf("patch%d.yaml", res)
				fmt.Fprintf(&buf, " - path: %v\n", fn)
				fSys.WriteFile(filepath.Join(path, fn), []byte(fmt.Sprintf(`kind: ConfigMap
apiVersion: v1
metadata:
  name: %s-%d
data:
  k: v2
`, id, res)))
			} else {
				fn := fmt.Sprintf("patch%d.json", res)
				fmt.Fprintf(&buf, ` - path: %v
   target:
    version: v1
    kind: ConfigMap
    name: %s-%d
`, fn, id, res-1)
				fSys.WriteFile(filepath.Join(path, fn), []byte(`[{"op": "add", "path": "/data/k2", "value": "3"} ]`))
			}
		}
	}

	return fSys.WriteFile(filepath.Join(path, "kustomization.yaml"), buf.Bytes())
}

func BenchmarkBuild(b *testing.B) {
	fSys := filesys.MakeFsInMemory()
	if err := makeKustomization(fSys, "testdata", "res", 0); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	buffy := new(bytes.Buffer)
	cmd := NewCmdBuild(fSys, MakeHelp("foo", "bar"), buffy)
	if err := cmd.RunE(cmd, []string{"./testdata"}); err != nil {
		b.Fatal(err)
	}
}
