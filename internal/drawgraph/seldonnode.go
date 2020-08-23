package drawgraph

import (
	"log"

	"github.com/goccy/go-graphviz/cgraph"
)

type SeldonCoreNode struct {
	NAME     string
	TYPE     string
	CHILDREN []SeldonCoreNode
	node     *cgraph.Node
	noResp   bool
}

func (self *SeldonCoreNode) Build(g *cgraph.Graph) SeldonCoreNode {
	var err error
	self.node, err = g.CreateNode(self.NAME)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	self.Dress()

	if self.TYPE == "COMBINER" {
		// Prevent COMBINER children connect to Response node,
		// e.g. [ combiner [ transformer[ model-a ], model-b ] ],
		//      model-a should not connect to Response node
		self.PreventChildrenConnectResponseNode()
	}

	if len(self.CHILDREN) > 0 {
		for index, child := range self.CHILDREN {
			self.CHILDREN[index] = child.Build(g)
		}
	}
	return *self
}

func (self *SeldonCoreNode) Dress() {
	switch self.TYPE {
	case "MODEL":
		self.node = self.node.SetColor("chocolate1")
		self.node = self.node.SetShape(cgraph.OctagonShape)
		self.node = self.node.SetStyle(cgraph.FilledNodeStyle)
	case "TRANSFORMER":
		self.node = self.node.SetColor("burlywood")
		self.node = self.node.SetShape(cgraph.OvalShape)
		self.node = self.node.SetStyle(cgraph.FilledNodeStyle)
	case "OUTPUT_TRANSFORMER":
		self.node = self.node.SetColor("burlywood")
		self.node = self.node.SetShape(cgraph.OvalShape)
		self.node = self.node.SetStyle(cgraph.FilledNodeStyle)
	case "ROUTER":
		self.node = self.node.SetColor("blue")
		self.node = self.node.SetShape(cgraph.DoubleCircleShape)
	case "COMBINER":
		self.node = self.node.SetColor("turquoise")
		self.node = self.node.SetShape(cgraph.DoubleOctagonShape)
	}
}

func (self *SeldonCoreNode) PreventChildrenConnectResponseNode() {
	for i, child := range self.CHILDREN {
		child.noResp = true
		child.PreventChildrenConnectResponseNode()

		self.CHILDREN[i] = child
	}
}
