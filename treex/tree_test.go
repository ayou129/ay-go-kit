package treex

import (
	"testing"
)

// testNode 测试用节点
type testNode struct {
	ID       int64
	ParentID int64
	Name     string
	Children []TreeNode
}

func (n *testNode) GetID() int64              { return n.ID }
func (n *testNode) GetParentID() int64        { return n.ParentID }
func (n *testNode) SetChildren(c []TreeNode)  { n.Children = c }
func (n *testNode) GetChildren() []TreeNode   { return n.Children }

func TestBuildTree_Empty(t *testing.T) {
	result := BuildTree([]*testNode{})
	if len(result) != 0 {
		t.Errorf("empty input should return empty slice, got %d", len(result))
	}
}

func TestBuildTree_AllRoots(t *testing.T) {
	nodes := []*testNode{
		{ID: 1, ParentID: 0, Name: "a"},
		{ID: 2, ParentID: 0, Name: "b"},
		{ID: 3, ParentID: 0, Name: "c"},
	}
	result := BuildTree(nodes)
	if len(result) != 3 {
		t.Errorf("all roots: expected 3, got %d", len(result))
	}
}

func TestBuildTree_TwoLevels(t *testing.T) {
	nodes := []*testNode{
		{ID: 1, ParentID: 0, Name: "root"},
		{ID: 2, ParentID: 1, Name: "child1"},
		{ID: 3, ParentID: 1, Name: "child2"},
	}
	result := BuildTree(nodes)

	if len(result) != 1 {
		t.Fatalf("expected 1 root, got %d", len(result))
	}
	root := result[0]
	if len(root.GetChildren()) != 2 {
		t.Errorf("root should have 2 children, got %d", len(root.GetChildren()))
	}
}

func TestBuildTree_ThreeLevels(t *testing.T) {
	nodes := []*testNode{
		{ID: 1, ParentID: 0, Name: "root"},
		{ID: 2, ParentID: 1, Name: "child"},
		{ID: 3, ParentID: 2, Name: "grandchild"},
	}
	result := BuildTree(nodes)

	if len(result) != 1 {
		t.Fatalf("expected 1 root, got %d", len(result))
	}
	child := result[0].GetChildren()[0]
	if len(child.GetChildren()) != 1 {
		t.Errorf("child should have 1 grandchild, got %d", len(child.GetChildren()))
	}
}

func TestBuildTree_OrphanNodes(t *testing.T) {
	nodes := []*testNode{
		{ID: 1, ParentID: 0, Name: "root"},
		{ID: 2, ParentID: 999, Name: "orphan"}, // parent 不存在
	}
	result := BuildTree(nodes)

	if len(result) != 1 {
		t.Errorf("orphan should be excluded from roots, got %d roots", len(result))
	}
	if len(result[0].GetChildren()) != 0 {
		t.Errorf("root should have 0 children (orphan not attached), got %d", len(result[0].GetChildren()))
	}
}

func TestBuildTree_PreservesOrder(t *testing.T) {
	nodes := []*testNode{
		{ID: 3, ParentID: 0, Name: "c"},
		{ID: 1, ParentID: 0, Name: "a"},
		{ID: 2, ParentID: 0, Name: "b"},
	}
	result := BuildTree(nodes)

	if result[0].Name != "c" || result[1].Name != "a" || result[2].Name != "b" {
		t.Errorf("order not preserved: %s %s %s", result[0].Name, result[1].Name, result[2].Name)
	}
}
