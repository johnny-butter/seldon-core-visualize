package drawgraph

import (
	"fmt"
	"log"

	"github.com/goccy/go-graphviz/cgraph"
)

type DrawInferenceGraph struct {
	Graph     *cgraph.Graph
	RootNode  SeldonCoreNode
	Nodes     []SeldonCoreNode
	HeadNodes []SeldonCoreNode
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
	if n.TYPE == "TRANSFORMER" || n.TYPE == "MODEL" || n.TYPE == "ROUTER" {
		s = append(s, n)
	}

	var children_s [][]interface{}
	for _, cn := range n.CHILDREN {
		child_s := self.GetApiSequence(cn)

		if len(n.CHILDREN) > 1 {
			children_s = append(children_s, child_s)
		} else {
			s = append(s, child_s...)
		}
	}

	if len(children_s) > 0 {
		s = append(s, children_s)
	}

	if n.TYPE == "COMBINER" || n.TYPE == "OUTPUT_TRANSFORMER" {
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
	self.HeadNodes = append(self.HeadNodes, head)

	return edge
}

func (self *DrawInferenceGraph) CreateNode(name string, shape cgraph.Shape) SeldonCoreNode {
	var err error

	sn := SeldonCoreNode{NAME: name}
	sn.node, err = self.Graph.CreateNode(sn.NAME)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	sn.node = sn.node.SetShape(shape)

	return sn
}

func (self *DrawInferenceGraph) DrawRequestEdge(n SeldonCoreNode) {
	rn := SeldonCoreNode{NAME: "Request"}
	var err error

	rn.node, err = self.Graph.CreateNode(rn.NAME)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	rn.node = rn.node.SetShape(cgraph.InvHouseShape)

	self.DrawEdge(rn, n)
}

func (self *DrawInferenceGraph) DrawResponseEdge() {
	rn := SeldonCoreNode{NAME: "Response"}
	var err error

	rn.node, err = self.Graph.CreateNode(rn.NAME)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	rn.node = rn.node.SetShape(cgraph.InvHouseShape)

	var none_head_nodes []SeldonCoreNode
	for _, node := range self.Nodes {
		// Equal: if node not in head_node:
		//            none_head_nodes.append(node)
		found := false
		for _, head_node := range self.HeadNodes {
			if node.NAME == head_node.NAME {
				found = true
			}
		}
		if !found {
			none_head_nodes = append(none_head_nodes, node)
		}
	}

	for _, node := range none_head_nodes {
		if !node.noResp {
			self.DrawEdge(node, rn)
		}
	}
}

func (self *DrawInferenceGraph) DrawRouterEdge(n SeldonCoreNode) {
	for i, child := range n.CHILDREN {
		var edge *cgraph.Edge

		if child.TYPE == "OUTPUT_TRANSFORMER" {
			sn := self.GetRouterConnectNode(child)

			edge = self.DrawEdge(n, sn)
		} else {
			edge = self.DrawEdge(n, child)
		}

		label := fmt.Sprintf("option[%d]", i)

		edge.SetLabel(label)
	}
}

func (self *DrawInferenceGraph) GetRouterConnectNode(n SeldonCoreNode) (sn SeldonCoreNode) {
	// Handle the event that child of "ROUTER" connects to "OUTPUT_TRANSFORMER"
	for _, child := range n.CHILDREN {
		if child.TYPE == "OUTPUT_TRANSFORMER" {
			sn = self.GetRouterConnectNode(child)
		} else {
			sn = child
			break
		}
	}
	return
}

func (self *DrawInferenceGraph) DrawCombinerEdge(n SeldonCoreNode) {
	for i, child := range n.CHILDREN {
		edge := self.DrawEdge(child, n)
		label := fmt.Sprintf("element[%d]", i)

		edge.SetDir(cgraph.BothDir)
		edge.SetLabel(label)
	}

}

func (self *DrawInferenceGraph) DrawModelEdge(n SeldonCoreNode) {
	if len(n.CHILDREN) > 0 {
		self.DrawEdge(n, n.CHILDREN[0])
	}
}

func (self *DrawInferenceGraph) DrawTransformerEdge(n SeldonCoreNode) {
	if len(n.CHILDREN) > 0 {
		self.DrawEdge(n, n.CHILDREN[0])
	}
}

func (self *DrawInferenceGraph) DrawOutputTransformerEdge(n SeldonCoreNode) {
	if len(n.CHILDREN) <= 0 {
		return
	}

	if n.CHILDREN[0].TYPE == "OUTPUT_TRANSFORMER" || n.CHILDREN[0].TYPE == "COMBINER" {
		self.DrawEdge(n.CHILDREN[0], n)
	} else {
		last_child := n.CHILDREN[0]
		for len(last_child.CHILDREN) > 0 {
			last_child = last_child.CHILDREN[0]
		}

		self.DrawEdge(last_child, n)
	}
}
