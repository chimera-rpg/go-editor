package data

import (
	"sort"
	"strings"
)

// ArchetypeTreeNode is a tree structure representing archetypes with their proper folder hierarchy.
type ArchetypeTreeNode struct {
	Name     string
	IsTree   bool
	Children []ArchetypeTreeNode
}

// ParseArchetypesIntoTree parses a full string slice of archetype paths to a tree.
func ParseArchetypesIntoTree(paths []string) ArchetypeTreeNode {
	tree := ArchetypeTreeNode{
		IsTree: true,
	}

	// Collect our paths into a map of [top][children]
	children := make(map[string][]string)
	for i := 0; i < len(paths); i++ {
		parts := strings.SplitN(paths[i], "/", 2)
		var key string
		if len(parts) == 0 {
			key = paths[i]
		} else {
			key = parts[0]
		}
		_, ok := children[key]
		if !ok {
			children[key] = make([]string, 0)
		}
		if len(parts) > 1 {
			children[key] = append(children[key], parts[1])
		}
	}

	keys := make([]string, 0, len(children))
	for k := range children {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Now paths should be parsed into hierarchical buckets.
	for _, k := range keys {
		c := children[k]
		var node ArchetypeTreeNode
		if len(c) > 0 {
			node = ParseArchetypesIntoTree(c)
		} else {
			node = ArchetypeTreeNode{
				IsTree: false,
			}
		}
		node.Name = k
		tree.Children = append(tree.Children, node)
	}
	return tree
}
