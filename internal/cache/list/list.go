package list

type (
	// Node is the linked list's node.
	Node struct {
		key   string
		value any
		next  *Node
		prev  *Node
	}

	// DoublyLinkedList is a doubly linked list of nodes.
	DoublyLinkedList struct {
		size uint64
		head *Node
		tail *Node
	}
)

// NewDoublyLinkedList returns a new doubly linked list.
func NewDoublyLinkedList() *DoublyLinkedList {
	return &DoublyLinkedList{}
}

// GetVal returns the node's value.
func (n *Node) GetVal() any {
	return n.value
}

// SetVal sets node's value.
func (n *Node) SetVal(val any) {
	n.value = val
}

// GetKey returns the node's key.
func (n *Node) GetKey() string {
	return n.key
}

// Size returns the size of the linked list.
func (l *DoublyLinkedList) Size() uint64 {
	return l.size
}

// Head returns the linked list's head.
func (l *DoublyLinkedList) Head() *Node {
	return l.head
}

// Tail returns the linked list's tail.
func (l *DoublyLinkedList) Tail() *Node {
	return l.tail
}

// AddToBack adds a new key-value pair to the back of the linked list and returns the added node.
func (l *DoublyLinkedList) AddToBack(key string, val any) *Node {
	if l.size == 0 {
		node := Node{key: key, value: val, next: nil, prev: nil}
		l.head = &node
		l.tail = &node
		l.size++

		return &node
	}

	node := Node{key: key, value: val, next: nil, prev: l.tail}
	l.tail.next = &node
	l.tail = &node
	l.size++

	return &node
}

// MoveToBack moves a node to the back of the linked list.
func (l *DoublyLinkedList) MoveToBack(node *Node) {
	if l.size == 0 {
		panic("list is empty")
	}

	switch node {
	case l.tail:
	case l.head:
		l.head = node.next
		node.next.prev = nil
		node.next = nil
		node.prev = l.tail
		l.tail.next = node
		l.tail = node
	default:
		prev := node.prev
		next := node.next
		prev.next = next
		next.prev = prev
		node.next = nil
		node.prev = l.tail
		l.tail = node
	}
}

// RemoveHead removes the head node.
func (l *DoublyLinkedList) RemoveHead() string {
	if l.size == 0 {
		panic("list is empty")
	}

	head := l.head

	next := head.next
	l.head = next
	l.size--
	if next != nil {
		next.prev = nil
	}

	key := head.key
	head = nil

	return key
}
