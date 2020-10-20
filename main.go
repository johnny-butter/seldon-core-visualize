package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"internal/drawgraph"

	"github.com/goccy/go-graphviz"
	"gopkg.in/yaml.v2"
)

type SeldonDeployment struct {
	Spec struct {
		Name       string
		Predictors []struct {
			Graph drawgraph.SeldonCoreNode
		}
	}
}

func BuildNodes(n drawgraph.SeldonCoreNode, ns []drawgraph.SeldonCoreNode) []drawgraph.SeldonCoreNode {
	ns = append(ns, n)
	if len(n.Children) > 0 {
		for _, child := range n.Children {
			ns = BuildNodes(child, ns)
		}
	}

	return ns
}

var DeploymentPath string
var OutputFilename string

func main() {
	fmt.Printf("Enter deployment path (/path/to/name.yml): ")
	fmt.Scanln(&DeploymentPath)

	fmt.Printf("Enter output filename (name.png): ")
	fmt.Scanln(&OutputFilename)
	if OutputFilename == "" {
		OutputFilename = "flowchart.png"
	}

	sd := SeldonDeployment{}

	g := graphviz.New()
	graph, err := g.Graph()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	yaml_file, err := ioutil.ReadFile(DeploymentPath)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	if err := yaml.Unmarshal([]byte(yaml_file), &sd); err != nil {
		log.Fatalf("error: %v", err)
	}

	nodes := BuildNodes(sd.Spec.Predictors[0].Graph.Build(graph), nil)

	d := &drawgraph.DrawInferenceGraph{Graph: graph, RootNode: nodes[0]}
	d.Draw()

	defer func() {
		if err := graph.Close(); err != nil {
			log.Fatal(err)
		}
		g.Close()
	}()

	g.RenderFilename(graph, "png", OutputFilename)
}
