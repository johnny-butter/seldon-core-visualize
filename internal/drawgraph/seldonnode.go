package drawgraph

import (
	"log"

	"github.com/goccy/go-graphviz/cgraph"
)

type SeldonCoreNode struct {
	Name     string
	Type     string
	Children []SeldonCoreNode
	node     *cgraph.Node
}

func (self *SeldonCoreNode) Build(g *cgraph.Graph) SeldonCoreNode {
	var err error
	self.node, err = g.CreateNode(self.Name)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	self.Dress()

	if len(self.Children) > 0 {
		for index, child := range self.Children {
			self.Children[index] = child.Build(g)
		}
	}
	return *self
}

func (self *SeldonCoreNode) Dress() {
	switch self.Type {
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
