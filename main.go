package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

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

var (
	sdepPath    string
	outFileName string
)

// Init CLI params
func init() {
	flag.StringVar(&outFileName, "o", "flowchart.png", "`output` Output flowchart name")

	flag.Usage = usage
}

// Show the usage
func usage() {
	fmt.Fprintln(os.Stderr, "Usage: draw-flowchart [OPTIONS] SDEP_PATH")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "OPTIONS:")
	flag.PrintDefaults()
}

func main() {
	flag.Parse()
	sdepPath = flag.Arg(0)
	if sdepPath == "" {
		log.Fatalf("Error: Missing Seldon Deployment path")
	}

	sd := SeldonDeployment{}

	g := graphviz.New()
	graph, err := g.Graph()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	yamlFile, err := ioutil.ReadFile(sdepPath)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	if err := yaml.Unmarshal([]byte(yamlFile), &sd); err != nil {
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

	g.RenderFilename(graph, "png", outFileName)
}
