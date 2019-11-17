package main

import (
	"errors"
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"math/rand"
	"sort"
	"time"
)

// this is the actor struct
type Node struct {
	id			int32
	token 		string
	maxLeft		int32
	maxSize 	int32
	left		*Node
	right		*Node
	isLeaf		bool
	key_values	map[int32]string
}

type Key_Value struct {
	Key		int32
	Value   string
}

// next structs are messages for Node actors
type NewTree struct {
	size 	int32
}

type Search struct {
	id 		int32
	token	string
	key		int32
}

type Insert struct {
	id 		int32
	token	string
	value	string
}

type Delete struct {
	id 		int32
	token	string
	key		int32
}

type Traverse struct {
	id 		int32
	token	string
}

type DelTree struct{
	id 		int32
	token	string
}

func newTree(size int32) Node {
	rand.Seed(time.Now().UnixNano())

	result := Node {
		rand.Int31n(10000),
		"Node_"+time.Now().String(),
		0,
		size,
		nil,
		nil,
		true,
		make(map[int32]string)}

	return result
}

func (root Node) search(key int32) (result string, err error) {
	result = ""
	cur := root.findLeaf(key)

		for k,v := range cur.key_values {
			if k == key {
				result = v
			}
		}
		if result == "" {
			err = errors.New("The key doesn't exist in this tree.\n")
		}

	return result, err
}

func (root Node) insert(key int32, value string) bool {
	cur := root

	if cur.isLeaf {
		cur.key_values[key] = value
		if int32(len(cur.key_values)) > cur.maxSize {
			cur.isLeaf = false
			cur.left = cur.createChild()
			cur.right = cur.createChild()
			cur.maxLeft, cur.left.key_values, cur.right.key_values = splitMapAndMarkFlag(cur.key_values)

			for k := range cur.key_values {
				delete(cur.key_values, k)
			}

			if (cur.maxLeft < key) {
				cur.right.insert(key, value)
			} else {
				cur.left.insert(key, value)
			}
		}
	}
	return true
}

func (cur Node) createChild() *Node {
	return &Node {
		cur.id,
		cur.token,
		0,
		cur.maxSize,
		nil,
		nil,
		true,
		make(map[int32]string)}
}

// helper method
func splitMapAndMarkFlag(src map[int32]string) (int32, map[int32]string, map[int32]string) {
	var flag int
	left := make(map[int32]string)
	right := make(map[int32]string)

	keys := make([]int, 0)

	for k := range src {
		keys = append(keys, int(k))
	}

	sort.Ints(keys)
	length := len(keys)
	flag = length/2

	for i := 0; i < length; i++ {
		key := int32(keys[i])
		if i < flag {
			left[key] = src[key]
		} else {
			right[key] = src[key]
		}
	}

	return int32(keys[flag]), left, right
}

func (root Node) delete(key int32) {
	cur := root.findLeaf(key)
	delete(cur.key_values, key)
}

func (node Node) findLeaf(key int32) Node {
	cur := node
	for ; !cur.isLeaf; {
		if cur.maxLeft < key {
			cur = *cur.right
		} else {
			cur = *cur.left
		}
	}
	return cur
}

func (root Node) traverse(id int32) []Key_Value {
	cur := root

	if cur.left == nil && cur.right == nil {
		return cur.getPairsForLeaf()
	} else {
		k_v_left := cur.left.traverse(id)
		k_v_right := cur.right.traverse(id)
		pairs := make([]Key_Value, 0)

		for _, el := range k_v_left {
			pairs = append(pairs, el)
		}

		for _, el := range k_v_right {
			pairs = append(pairs, el)
		}
		return pairs
	}
}

func(node Node) getPairsForLeaf() []Key_Value {
	cur := node
	data := cur.key_values
	keys := make([]int, 0)

	for k, _ := range data {
		keys = append(keys, int(k))
	}

	sort.Ints(keys)
	pairs := make([]Key_Value, 0)

	for _, k := range keys {
		pairs = append(pairs, Key_Value{int32(k), cur.key_values[int32(k)]})
	}
	return pairs
}

func (state *Node) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case NewTree:
		fmt.Print("%v", msg)
	case Search:
		fmt.Print("%v", msg)
	case Insert:
		fmt.Print("%v", msg)
	case Delete:
		fmt.Print("%v", msg)
	case Traverse:
		fmt.Print("%v", msg)
	case DelTree:
		fmt.Print("%v", msg)
	}
}


func main() {

	root := newTree(2)

	id := root.id
	token := root.token

	fmt.Printf("Stats: root.id = %v, root.token = %v, isLeaf = %v, map length = %v\n", id, token, root.isLeaf, len(root.key_values))

	root.insert(1, "a")
	root.insert(2, "b")

	fmt.Printf("Stats: root.id = %v, root.token = %v, isLeaf = %v, map length = %v\n", id, token, root.isLeaf, len(root.key_values))

	root.insert(3, "c")
	root.insert(4, "d")
	root.insert(5, "e")

	fmt.Printf("Stats: root.id = %v, root.token = %v, isLeaf = %v, map length = %v\n", id, token, root.isLeaf, len(root.key_values))

	root.delete(5)

	fmt.Printf("Stats: root.id = %v, root.token = %v, isLeaf = %v, map length = %v\n", id, token, root.isLeaf, len(root.key_values))


	k_v := root.traverse(id)
	fmt.Print("Traversed tree: \n")

	for _,v := range k_v{
		fmt.Printf("key = %v, v = %v\n", v.Key, v.Value)
	}

	/*
		context := actor.EmptyRootContext
		props := actor.PropsFromProducer(func() actor.Actor {
			return &Node{}
		})
		pid := context.Spawn(props)
		context.Send(pid, &NewTree{2})
		context.Send(pid, &Insert{})
		context.Send(pid, &Insert{})
		context.Send(pid, &Insert{})
		context.Send(pid, &Delete{})
		elements := context.Send(pid, &Traverse{})
		fmt.Print(elements)
		console.ReadLine()
	*/

	fmt.Print("------------------------------------------\n")

	//arr := { }
}
