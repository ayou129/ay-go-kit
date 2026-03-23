package treex

// TreeNode 树形节点接口
type TreeNode interface {
	GetID() int64
	GetParentID() int64
	SetChildren(children []TreeNode)
	GetChildren() []TreeNode
}

// BuildTree 通用树形结构构建函数
// nodes: 所有节点列表（必须实现 TreeNode 接口）
// 返回: 根节点列表（parent_id = 0 的节点）
func BuildTree[T TreeNode](nodes []T) []T {
	if len(nodes) == 0 {
		return []T{}
	}

	// 构建 ID -> Node 映射
	nodeMap := make(map[int64]T)
	for _, node := range nodes {
		nodeMap[node.GetID()] = node
		// 初始化 children（避免 nil）
		node.SetChildren([]TreeNode{})
	}

	// 组装树形结构
	var tree []T
	for _, node := range nodes {
		if node.GetParentID() == 0 {
			tree = append(tree, node)
		} else {
			if parent, exists := nodeMap[node.GetParentID()]; exists {
				children := parent.GetChildren()
				children = append(children, node)
				parent.SetChildren(children)
			}
		}
	}

	return tree
}
