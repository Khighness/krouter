package krouter

import (
	"net/http"
	"strings"
)

// @Author KHighness
// @Update 2022-11-08

// Tree is a trie tree.
type Tree struct {
	root       *Node
	parameters Parameters
	routes     map[string]*Node
}

// Node records any URL params, and executes an end handler.
type Node struct {
	// key records Node's key
	key string

	// path records a request uri
	path string

	// handle is a function to process current path's request
	handle http.HandlerFunc

	// depths records Node's depth
	depth int

	// children records Node's children node
	children map[string]*Node

	// isEnd judges if Node is leaf
	isEnd bool

	// middleware records middleware stack
	middleware []Middleware
}

// NewNode creates a newly initialized Node.
func NewNode(key string, depth int) *Node {
	return &Node{
		key:      key,
		depth:    depth,
		children: make(map[string]*Node),
	}
}

// NewTree creates a newly initialized Tree.
func NewTree() *Tree {
	return &Tree{
		root:   NewNode("/", 1),
		routes: make(map[string]*Node),
	}
}

// Register adds a node to Tree.
func (t *Tree) Register(pattern string, handle http.HandlerFunc, middleware ...Middleware) {
	var currNode = t.root

	if pattern != currNode.key {
		pattern = trimPathPrefix(pattern)
		keyList := splitPattern(pattern)
		for _, key := range keyList {
			node, ok := currNode.children[key]
			if !ok {
				node = NewNode(key, currNode.depth+1)
				if len(middleware) > 0 {
					node.middleware = append(node.middleware, middleware...)
				}
				currNode.children[key] = node
			}
			currNode = node
		}
	}

	if len(middleware) > 0 && currNode.depth == 1 {
		currNode.middleware = append(currNode.middleware, middleware...)
	}

	currNode.handle = handle
	currNode.isEnd = true
	currNode.path = pattern

	if routeName := t.parameters.routeName; routeName != "" {
		t.routes[routeName] = currNode
	}
}

// Search returns nodes that the request matches the route pattern.
func (t *Tree) Search(pattern string, isRegex bool) (nodes []*Node) {
	var (
		node  = t.root
		queue []*Node
	)

	if pattern == node.path {
		nodes = append(nodes, node)
		return
	}

	if !isRegex {
		pattern = trimPathPrefix(pattern)
	}
	keyList := splitPattern(pattern)

	for _, key := range keyList {
		child, ok := node.children[key]
		if !ok {
			if isRegex {
				break
			} else {
				return
			}
		}
		if pattern == child.path && !isRegex {
			nodes = append(nodes, child)
			return
		}
		node = child
	}

	queue = append(queue, node)

	for len(queue) > 0 {
		var queueTemp []*Node
		for _, n := range queue {
			if n.isEnd {
				nodes = append(nodes, n)
			}
			for _, childNode := range n.children {
				queueTemp = append(queueTemp, childNode)
			}
		}
		queue = queueTemp
	}

	return
}

// trimPathPrefix removes the prefix symbol '/' of pattern string.
func trimPathPrefix(pattern string) string {
	return strings.TrimPrefix(pattern, "/")
}

// splitPattern splits pattern string into string array.
func splitPattern(pattern string) []string {
	return strings.Split(pattern, "/")
}
