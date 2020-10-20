package drawgraph

import (
	"fmt"
	"log"

	"github.com/goccy/go-graphviz/cgraph"
)

type DrawInferenceGraph struct {
	Graph    *cgraph.Graph
	RootNode SeldonCoreNode
}

func (self *DrawInferenceGraph) Draw() {
	// Create REQUEST, RESPONSE node
	req := self.CreateNode("Request", cgraph.InvHouseShape)
	rsp := self.CreateNode("Response", cgraph.InvHouseShape)

	// Get the sequence of api call
	s := self.GetApiSequence(self.RootNode)

	// Draw the flowchart of api call
	self.ConcatNodes(req, s[0])

	for i := 0; i < len(s)-1; i++ {
		self.ConcatNodes(s[i], s[i+1])
	}

	self.ConcatNodes(s[len(s)-1], rsp)
}

// Follow the logic of Seldon Core. Reference:
// https://github.com/SeldonIO/seldon-core/blob/master/engine/src/main/java/io/seldon/engine/predictors/PredictiveUnitBean.java
// -> public Future<SeldonMessage> getOutputAsync
func (self *DrawInferenceGraph) GetApiSequence(n SeldonCoreNode) (s []interface{}) {
	if n.Type == "TRANSFORMER" || n.Type == "MODEL" || n.Type == "ROUTER" {
		s = append(s, n)
	}

	var children_s [][]interface{}
	for _, cn := range n.Children {
		child_s := self.GetApiSequence(cn)

		if len(n.Children) > 1 {
			children_s = append(children_s, child_s)
		} else {
			s = append(s, child_s...)
		}
	}

	if len(children_s) > 0 {
		s = append(s, children_s)
	}

	if n.Type == "COMBINER" || n.Type == "OUTPUT_TRANSFORMER" {
		s = append(s, n)
	}

	return
}

func (self *DrawInferenceGraph) ConcatNodes(h interface{}, e interface{}) {
	switch h := h.(type) {
	case SeldonCoreNode:
		switch e := e.(type) {
		case SeldonCoreNode:
			self.DrawEdge(h, e)
		case [][]interface{}:
			for _, ee := range e {
				for i, eee := range ee {
					// Nested concat
					if i > 0 {
						self.ConcatNodes(ee[i-1], eee)
						continue
					}
					// Only connect to first one
					switch eee := eee.(type) {
					case SeldonCoreNode:
						self.DrawEdge(h, eee)
					case interface{}:
						self.ConcatNodes(h, eee)
					default:
						fmt.Println("Unexpected type of eee")
						fmt.Println(eee)
					}
				}
			}
		}
	case [][]interface{}:
		for _, hh := range h {
			// Only use last one to connect
			switch hhh := hh[len(hh)-1].(type) {
			case SeldonCoreNode:
				self.ConcatNodes(hhh, e)
			default:
				fmt.Println("Unexpected type of hhh")
				fmt.Println(hhh)
			}
		}
	default:
		fmt.Println("Unexpected type of h")
		fmt.Println(h)
	}
}

func (self *DrawInferenceGraph) DrawEdge(head SeldonCoreNode, tail SeldonCoreNode) *cgraph.Edge {
	edge, _ := self.Graph.CreateEdge("", head.node, tail.node)

	return edge
}

func (self *DrawInferenceGraph) CreateNode(name string, shape cgraph.Shape) SeldonCoreNode {
	var err error

	sn := SeldonCoreNode{Name: name}
	sn.node, err = self.Graph.CreateNode(sn.Name)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	sn.node = sn.node.SetShape(shape)

	return sn
}
