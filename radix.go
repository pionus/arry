package arry

import (
	"path"
	"sort"
	"strings"
)

// NodeType categorizes nodes for priority-based matching
type NodeType uint8

const (
	ntRoot NodeType = iota // Root node (/)
	ntStatic               // Exact match segment (/users)
	ntParam                // Parameter segment (/:id)
	ntCatchAll             // Wildcard segment (/*)
)

// radixNode represents a node in the radix tree
type radixNode struct {
	// prefix is the compressed path prefix for this node
	prefix string

	// label is the first character of prefix for quick lookups
	label byte

	// nodeType indicates the type of this node
	nodeType NodeType

	// children stores static child nodes, sorted by label
	children []*radixNode

	// paramChild is the single parameter child node
	paramChild *radixNode

	// catchAllChild is the single wildcard child node
	catchAllChild *radixNode

	// methods maps HTTP method to handler
	methods map[string]Handler

	// paramName stores the parameter name for Param nodes
	paramName string
}

// radixTree is the radix tree structure for routing
type radixTree struct {
	root *radixNode
}

// newRadixTree creates a new radix tree
func newRadixTree() *radixTree {
	return &radixTree{
		root: &radixNode{
			prefix:   "/",
			label:    '/',
			nodeType: ntRoot,
			methods:  make(map[string]Handler),
		},
	}
}

// insert adds a route to the radix tree
func (t *radixTree) insert(method, pattern string, handler Handler) {
	segments := strings.Split(pattern, "/")[1:] // Skip empty first element
	t.root.insertRecursive(method, segments, handler, 0)
}

// search finds the handler for the given path
func (t *radixTree) search(path string, ctx Context) *radixNode {
	segments := strings.Split(path, "/")[1:]
	return t.root.searchRecursive(segments, ctx)
}

// insertRecursive recursively inserts route segments into the tree
func (n *radixNode) insertRecursive(method string, segments []string, handler Handler, depth int) {
	// Base case: reached the end of the pattern
	if depth >= len(segments) {
		if n.methods == nil {
			n.methods = make(map[string]Handler)
		}
		n.methods[strings.ToUpper(method)] = handler
		return
	}

	segment := segments[depth]

	// Skip empty segments
	if len(segment) == 0 {
		n.insertRecursive(method, segments, handler, depth+1)
		return
	}

	// Determine segment type
	segmentType := ntStatic
	paramName := segment
	if segment[0] == ':' {
		segmentType = ntParam
		paramName = segment[1:] // Remove ':' prefix
	} else if segment[0] == '*' {
		segmentType = ntCatchAll
		paramName = "*"
	}

	// Route to appropriate child based on type
	switch segmentType {
	case ntParam:
		// Parameter routes: stored in paramChild
		if n.paramChild == nil {
			n.paramChild = &radixNode{
				prefix:    paramName,
				label:     ':',
				nodeType:  ntParam,
				paramName: paramName,
				methods:   make(map[string]Handler),
			}
		}
		n.paramChild.insertRecursive(method, segments, handler, depth+1)

	case ntCatchAll:
		// Wildcard routes: stored in catchAllChild
		if n.catchAllChild == nil {
			n.catchAllChild = &radixNode{
				prefix:    "*",
				label:     '*',
				nodeType:  ntCatchAll,
				paramName: "*",
				methods:   make(map[string]Handler),
			}
		}
		// Catch-all must be the last segment
		if n.catchAllChild.methods == nil {
			n.catchAllChild.methods = make(map[string]Handler)
		}
		n.catchAllChild.methods[strings.ToUpper(method)] = handler

	case ntStatic:
		// Static routes: use LCP logic for prefix compression
		n.insertStatic(method, segment, segments, handler, depth)
	}
}

// insertStatic handles insertion of static route segments with LCP-based prefix compression
func (n *radixNode) insertStatic(method, segment string, segments []string, handler Handler, depth int) {
	// Search for existing child with common prefix
	for i, child := range n.children {
		lcp := longestCommonPrefix(segment, child.prefix)

		if lcp > 0 {
			// Case 1: Exact match - traverse into child
			if lcp == len(child.prefix) && lcp == len(segment) {
				child.insertRecursive(method, segments, handler, depth+1)
				return
			}

			// Case 2: Partial match - need to split the child node
			if lcp < len(child.prefix) {
				child.splitNode(lcp)
			}

			// Case 3: Segment continues beyond LCP - create new sibling
			if lcp < len(segment) {
				newChild := &radixNode{
					prefix:   segment[lcp:],
					label:    segment[lcp],
					nodeType: ntStatic,
					methods:  make(map[string]Handler),
				}
				// Insert new child after the split point
				n.children = append(n.children[:i+1], append([]*radixNode{newChild}, n.children[i+1:]...)...)
				n.sortChildren()
				newChild.insertRecursive(method, segments, handler, depth+1)
			} else {
				// Segment matches exactly up to child prefix
				child.insertRecursive(method, segments, handler, depth+1)
			}
			return
		}
	}

	// No matching child found - create new static child
	newChild := &radixNode{
		prefix:   segment,
		label:    segment[0],
		nodeType: ntStatic,
		methods:  make(map[string]Handler),
	}
	n.children = append(n.children, newChild)
	n.sortChildren()
	newChild.insertRecursive(method, segments, handler, depth+1)
}

// searchRecursive finds the matching node with specificity priority
func (n *radixNode) searchRecursive(segments []string, ctx Context) *radixNode {
	if len(segments) == 0 {
		return n
	}

	segment := segments[0]

	// Skip empty segments
	if len(segment) == 0 {
		return n.searchRecursive(segments[1:], ctx)
	}

	next := segments[1:]

	// PRIORITY 1: Static routes (exact match) - HIGHEST PRIORITY
	for _, child := range n.children {
		if child.prefix == segment {
			result := child.searchRecursive(next, ctx)
			if result != nil {
				return result
			}
		}
	}

	// PRIORITY 2: Parameter routes - MEDIUM PRIORITY
	if n.paramChild != nil {
		setParams(n.paramChild.paramName, segment, ctx)
		result := n.paramChild.searchRecursive(next, ctx)
		if result != nil {
			return result
		}
	}

	// PRIORITY 3: Catch-all routes - LOWEST PRIORITY
	if n.catchAllChild != nil {
		remaining := path.Join(segments...)
		setParams("*", remaining, ctx)
		return n.catchAllChild
	}

	// No match found
	return nil
}

// splitNode splits a node at the given position
func (n *radixNode) splitNode(pos int) {
	// Create child node with the remainder of the current prefix
	child := &radixNode{
		prefix:        n.prefix[pos:],
		label:         n.prefix[pos],
		nodeType:      n.nodeType,
		children:      n.children,
		paramChild:    n.paramChild,
		catchAllChild: n.catchAllChild,
		methods:       n.methods,
		paramName:     n.paramName,
	}

	// Update current node to be the split point
	n.prefix = n.prefix[:pos]
	if len(n.prefix) > 0 {
		n.label = n.prefix[0]
	}
	n.nodeType = ntStatic
	n.children = []*radixNode{child}
	n.paramChild = nil
	n.catchAllChild = nil
	n.methods = make(map[string]Handler)
	n.paramName = ""
}

// sortChildren sorts the static children by their label
func (n *radixNode) sortChildren() {
	if len(n.children) > 1 {
		sort.Slice(n.children, func(i, j int) bool {
			return n.children[i].label < n.children[j].label
		})
	}
}

// longestCommonPrefix calculates the longest common prefix between two strings
func longestCommonPrefix(a, b string) int {
	max := len(a)
	if len(b) < max {
		max = len(b)
	}

	var i int
	for i = 0; i < max; i++ {
		if a[i] != b[i] {
			break
		}
	}
	return i
}
