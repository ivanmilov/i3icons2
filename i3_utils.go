package main

// stolen from github.com/mdirkse/i3ipc-go/blob/master/tree_utils.go
import (
	"go.i3wm.org/i3/v4"
)

// Returns a slice of all descendent nodes.
func Descendents(tree i3.Tree, node_id int64) []*i3.Node {

	node := tree.Root.FindChild(func(n *i3.Node) bool { return n.ID == i3.NodeID(node_id) })

	if node == nil {
		return nil
	}

	var collectDescendents func(*i3.Node, []*i3.Node) []*i3.Node

	// Collects descendent nodes recursively
	collectDescendents = func(n *i3.Node, collected []*i3.Node) []*i3.Node {
		for i := range n.Nodes {
			if n.Type != "dockarea" {
				collected = append(collected, n.Nodes[i])
				collected = collectDescendents(n.Nodes[i], collected)
			}
		}
		for i := range n.FloatingNodes {
			if n.Type != "dockarea" {
				collected = append(collected, n.FloatingNodes[i])
				collected = collectDescendents(n.FloatingNodes[i], collected)
			}
		}
		return collected
	}

	return collectDescendents(node, nil)
}

// Returns nodes that has no children nodes (leaves).
func Leaves(tree i3.Tree, node_id int64) (leaves []*i3.Node) {

	nodes := Descendents(tree, node_id)

	for i := range nodes {
		node := nodes[i]

		if len(node.Nodes) == 0 && node.Type == "con" {
			leaves = append(leaves, node)
		}
	}
	return
}

// i3-msg -t get_workspaces doesn't fill IDs for the workspaces
// This function as a workaround
func GetWorkspaces(tree i3.Tree) []i3.Node {

	predicate := func(node *i3.Node) bool {
		return node.Type == "workspace" && node.Name != "__i3_scratch"
	}

	wss := []i3.Node{}
	wss = AppendChild(wss, tree.Root, predicate)

	return wss
}

func AppendChild(wss []i3.Node, n *i3.Node, predicate func(*i3.Node) bool) []i3.Node {
	if predicate(n) {
		wss = append(wss, *n)
	}
	for _, c := range n.Nodes {
		wss = AppendChild(wss, c, predicate)
	}
	return wss
}
