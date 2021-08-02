package trie

// RuneTrie is a trie of runes with string keys and interface{} values.
// Note that internal nodes have nil values so a stored nil Value will not
// be distinguishable and will not be included in Walks.
// Changed the value and children to exported values in order to reach them from code.
type RuneTrie struct {
	Value    interface{}
	Children map[rune]*RuneTrie
}

// NewRuneTrie allocates and returns a new *RuneTrie.
func NewRuneTrie() *RuneTrie {
	return new(RuneTrie)
}

// Get returns the Value stored at the given key. Returns nil for internal
// nodes or for nodes with a Value of nil.
func (trie *RuneTrie) Get(key string) interface{} {
	node := trie
	for _, r := range key {
		node = node.Children[r]
		if node == nil {
			return nil
		}
	}
	return node.Value
}

// Put inserts the Value into the trie at the given key, replacing any
// existing items. It returns true if the put adds a new Value, false
// if it replaces an existing Value.
// Note that internal nodes have nil values so a stored nil Value will not
// be distinguishable and will not be included in Walks.
func (trie *RuneTrie) Put(key string, value interface{}) bool {
	node := trie
	for _, r := range key {
		child, _ := node.Children[r]
		if child == nil {
			if node.Children == nil {
				node.Children = map[rune]*RuneTrie{}
			}
			child = new(RuneTrie)
			node.Children[r] = child
		}
		node = child
	}
	// does node have an existing Value?
	isNewVal := node.Value == nil
	node.Value = value
	return isNewVal
}

// Delete removes the Value associated with the given key. Returns true if a
// node was found for the given key. If the node or any of its ancestors
// becomes childless as a result, it is removed from the trie.
func (trie *RuneTrie) Delete(key string) bool {
	path := make([]nodeRune, len(key)) // record ancestors to check later
	node := trie
	for i, r := range key {
		path[i] = nodeRune{r: r, node: node}
		node = node.Children[r]
		if node == nil {
			// node does not exist
			return false
		}
	}
	// delete the node Value
	node.Value = nil
	// if leaf, remove it from its parent's Children map. Repeat for ancestor
	// path.
	if node.isLeaf() {
		// iterate backwards over path
		for i := len(key) - 1; i >= 0; i-- {
			parent := path[i].node
			r := path[i].r
			delete(parent.Children, r)
			if !parent.isLeaf() {
				// parent has other Children, stop
				break
			}
			parent.Children = nil
			if parent.Value != nil {
				// parent has a Value, stop
				break
			}
		}
	}
	return true // node (internal or not) existed and its Value was nil'd
}

// Walk iterates over each key/Value stored in the trie and calls the given
// walker function with the key and Value. If the walker function returns
// an error, the walk is aborted.
// The traversal is depth first with no guaranteed order.
func (trie *RuneTrie) Walk(walker WalkFunc) error {
	return trie.walk("", walker)
}

// WalkPath iterates over each key/Value in the path in trie from the root to
// the node at the given key, calling the given walker function for each
// key/Value. If the walker function returns an error, the walk is aborted.
func (trie *RuneTrie) WalkPath(key string, walker WalkFunc) error {
	// Get root Value if one exists.
	if trie.Value != nil {
		if err := walker("", trie.Value); err != nil {
			return err
		}
	}

	for i, r := range key {
		if trie = trie.Children[r]; trie == nil {
			return nil
		}
		if trie.Value != nil {
			if err := walker(string(key[0:i+1]), trie.Value); err != nil {
				return err
			}
		}
	}
	return nil
}

// RuneTrie node and the rune key of the child the path descends into.
type nodeRune struct {
	node *RuneTrie
	r    rune
}

func (trie *RuneTrie) walk(key string, walker WalkFunc) error {
	if trie.Value != nil {
		if err := walker(key, trie.Value); err != nil {
			return err
		}
	}
	for r, child := range trie.Children {
		if err := child.walk(key+string(r), walker); err != nil {
			return err
		}
	}
	return nil
}

func (trie *RuneTrie) isLeaf() bool {
	return len(trie.Children) == 0
}
