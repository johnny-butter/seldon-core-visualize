module visualize

go 1.14

require (
	github.com/goccy/go-graphviz v0.0.5
	gopkg.in/yaml.v2 v2.3.0
	internal/drawgraph v0.0.1
)

replace internal/drawgraph => ./internal/drawgraph
