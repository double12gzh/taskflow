package workflow

import (
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
)

// Node 节点
type Node string

// Edge 直接依赖关系
type Edge struct {
	Current  Node // 当前节点
	Previous Node // 直接前驱
}

// Graph 依赖图关系
type Graph struct {
	Edges map[Edge]struct{} // 记录所有边信息
	Nodes map[Node]struct{} // 记录所有节点信息

	dependencies map[Node][]Node // 记录所有单向直接依赖信息， key 为当前节点， value 为此节点的依赖节点

	mux *sync.RWMutex
}

func NewGraph() *Graph {
	return &Graph{
		Edges:        make(map[Edge]struct{}),
		Nodes:        make(map[Node]struct{}),
		dependencies: make(map[Node][]Node, 0),
		mux:          &sync.RWMutex{},
	}
}

// HasNode 是否包含此节点
func (g *Graph) HasNode(n Node) bool {
	g.mux.RLock()
	defer g.mux.RUnlock()

	_, ok := g.Nodes[n]

	return ok
}

// AddNode 添加节点
func (g *Graph) AddNode(n Node) {
	if g.HasNode(n) {
		log.Infof("node %v exists, skipped.", n)
		return
	}

	g.mux.Lock()
	defer g.mux.Unlock()

	g.Nodes[n] = struct{}{}
}

// HasEdge 是否存在边关系
func (g *Graph) HasEdge(e Edge) bool {
	g.mux.RLock()
	defer g.mux.RUnlock()

	_, ok := g.Edges[e]

	return ok
}

// AddEdge 添加边关系
func (g *Graph) AddEdge(e Edge) error {
	if g.HasEdge(e) {
		err := fmt.Errorf("edge [%v] already exists, skipped", e)
		log.Error(err.Error())
		return err
	}

	if g.HasCycles(e) {
		err := fmt.Errorf("cycles exists for edge [%+v], skipped", e)
		log.Error(err.Error())
		return err
	}

	g.AddNode(e.Current)
	g.AddNode(e.Previous)

	g.mux.Lock()
	defer g.mux.Unlock()

	g.Edges[e] = struct{}{}
	if _, ok := g.dependencies[e.Current]; !ok {
		g.dependencies[e.Current] = make([]Node, 0)
	}

	g.dependencies[e.Current] = append(g.dependencies[e.Current], e.Previous)

	return nil
}

// GetPrevious 获取直接前驱
func (g *Graph) GetPrevious(current Node) []Node {
	return g.dependencies[current]
}

// HasCycles 查找是否存在直接依赖的环（Notice: 不考虑间接依赖)
// TODO 图算法重构
func (g *Graph) HasCycles(e Edge) bool {
	deps, ok := g.dependencies[e.Previous]
	if !ok {
		return false
	}

	for _, value := range deps {
		value := value
		if value == e.Current {
			log.Errorf("cycles exists for edge [%+v]", e)
			return true
		}
	}

	return false
}
