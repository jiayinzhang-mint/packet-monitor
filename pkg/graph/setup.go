package graph

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/types"
)

// Graph graph def
type Graph struct {
	Nodes map[string]*Node   `json:"nodes"` // use ip as the key
	Edges map[string][]*Edge `json:"edges"` // use sourceip as the key

	lock sync.RWMutex
}

// NodeType types of graph nodes
type NodeType int

const (
	// TypeService node type service
	TypeService NodeType = iota
	// TypePod node type pod
	TypePod
)

// Node node def, denotes single microservice instance
type Node struct {
	Type    NodeType  `json:"type"`
	PodIP   string    `json:"podIP"`
	PodName string    `json:"podName"`
	PodUID  types.UID `json:"podUID"`
}

// Edge edge between nodes with direction, denotes dependency relationship
type Edge struct {
	// node key
	SourceIP string `json:"sourceIP"`
	DestIP   string `json:"destIP"`

	// additional info
	SourcePort uint16 `json:"sourcePort"`
	DestPort   uint16 `json:"destPort"`
	Protocol   string `json:"protocol"`
	Length     int    `json:"length"`
}

// Reset reset graph
func (g *Graph) Reset() {
	defer g.lock.Unlock()
	g.lock.Lock()

	g.Nodes = make(map[string]*Node)
	g.Edges = make(map[string][]*Edge)
}

// AddNode add a node to the graph
func (g *Graph) AddNode(n *Node) {
	defer g.lock.Unlock()
	g.lock.Lock()

	if g.Nodes == nil {
		g.Nodes = make(map[string]*Node)
	}

	g.Nodes[n.PodIP] = Merge(g.Nodes[n.PodIP], n)
}

// AddEdge add an edge to the graph
func (g *Graph) AddEdge(e *Edge) {
	defer g.lock.Unlock()
	g.lock.Lock()

	if g.Edges == nil {
		g.Edges = make(map[string][]*Edge)
	}

	g.Edges[e.SourceIP] = append(g.Edges[e.SourceIP], e)
}

// Merge merge node b into node a
func Merge(a, b *Node) *Node {
	if a == nil {
		return b
	}

	ra := reflect.ValueOf(a).Elem()
	rb := reflect.ValueOf(b).Elem()

	numFields := ra.NumField()

	for i := 0; i < numFields; i++ {
		fieldA := ra.Field(i)
		fieldB := rb.Field(i)

		switch fieldA.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
			if fieldA.IsNil() && fieldA.CanSet() && !fieldB.IsNil() {
				fieldA.Set(fieldB)
			}
		case reflect.String:
			if fieldA.IsZero() && fieldA.CanSet() && !fieldB.IsZero() {
				fieldA.Set(fieldB)
			}
		}
	}

	return (ra.Addr().Interface()).(*Node)
}

// String get graph structure with string
func (g *Graph) String() (s string) {
	defer g.lock.RUnlock()
	g.lock.RLock()

	for k, v := range g.Nodes {
		near := g.Edges[k]

		if len(near) == 0 {
			s += v.PodIP + "\n"
			continue
		}

		for j := 0; j < len(near); j++ {
			s += v.PodIP + " -> " +
				strconv.Itoa(int(near[j].SourcePort)) + " -> " +
				strconv.Itoa(int(near[j].DestPort)) + " -> " +
				near[j].DestIP + "\n"
		}
	}

	return
}

// Export export & save graph structure
func (g *Graph) Export() (err error) {
	var (
		exportURL = "export"
		res       []byte
	)

	type exportGraph struct {
		Nodes []*Node `json:"nodes"` // use ip as the key
		Edges []*Edge `json:"edges"` // use sourceip as the key
	}

	ge := &exportGraph{
		Nodes: make([]*Node, 0),
		Edges: make([]*Edge, 0),
	}

	for _, n := range g.Nodes {
		ge.Nodes = append(ge.Nodes, n)
	}
	for _, e := range g.Edges {
		ge.Edges = append(ge.Edges, e...)
	}

	envExportURL := os.Getenv("exportURL")
	if envExportURL != "" {
		exportURL = envExportURL
	}

	if res, err = json.Marshal(ge); err != nil {
		return
	}

	if err = os.WriteFile(filepath.Join(exportURL, strconv.Itoa(int(time.Now().Unix()))+".json"), res, os.ModePerm); err != nil {
		return
	}

	return
}
