package drawgraph

import (
	"fmt"
	"log"
	"reflect"

	"github.com/goccy/go-graphviz/cgraph"

	"github.com/iancoleman/strcase"
)

type DrawInferenceGraph struct {
	Graph     *cgraph.Graph
	Nodes     []SeldonCoreNode
	HeadNodes []SeldonCoreNode
}

func (self *DrawInferenceGraph) Draw() {
	var topNode SeldonCoreNode
	for _, n := range self.Nodes {
		if n.TYPE != "OUTPUT_TRANSFORMER" || len(self.Nodes) == 1 {
			topNode = n
			break
		}
	}

	self.DrawRequestEdge(topNode)

	for _, node := range self.Nodes {
		args := []reflect.Value{reflect.ValueOf(node)}
		method_name := fmt.Sprintf("Draw%sEdge", strcase.ToCamel(strcase.ToSnake(node.TYPE)))

		reflect.ValueOf(self).MethodByName(method_name).Call(args)
	}

	self.DrawResponseEdge()
}

func (self *DrawInferenceGraph) DrawEdge(head SeldonCoreNode, tail SeldonCoreNode) *cgraph.Edge {
	edge, _ := self.Graph.CreateEdge("", head.node, tail.node)
	self.HeadNodes = append(self.HeadNodes, head)

	return edge
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
