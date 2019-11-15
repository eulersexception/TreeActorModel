package tree

import (
	"errors"
	"fmt"
	console "github.com/AsynkronIT/goconsole"
	"github.com/AsynkronIT/protoactor-go/actor"
	"sort"

	"math/rand"
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
	cur := root

	if cur.isLeaf {
		for k,v := range cur.key_values {
			if k == key {
				result = v
			}
		}
		if result == "" {
			err = errors.New("The key doesn't exist in this tree.")
		}
	} else {
		for ; cur.isLeaf; {
			if cur.maxLeft <= key {
				cur = *cur.right
			} else {
				cur = *cur.left
			}
		}

		for k,v := range cur.key_values {
			if k == key {
				result = v
			}
		}
		if result == "" {
			err = errors.New("The key doesn't exist in this tree.")
		}
	}

	return result, nil
}

func (root Node) insert(key int32, value string) bool {
	cur := root

	if cur.isLeaf {
		if int32(len(cur.key_values)) < cur.maxSize {
			cur.key_values[key] = value
			return true
		} else {
			rand.Seed(time.Now().UnixNano())
			cur.left = &Node {
					cur.id,
					cur.token,
					0,
					cur.maxSize,
					nil,
					nil,
					true,
					make(map[int32]string)}

			cur.right = &Node {
				cur.id,
				cur.token,
				0,
				cur.maxSize,
				nil,
				nil,
				true,
				make(map[int32]string)}

			cur.maxLeft, cur.left.key_values, cur.right.key_values = splitMapAndMarkFlag(cur.key_values)
			cur.isLeaf = false

			if cur.maxLeft < key {
				cur = *cur.right
			} else {
				cur = *cur.left
			}
			cur.insert(key, value)
		}
	} else {
		for ; cur.isLeaf; {
			if cur.maxLeft < key {
				cur = *cur.right
			} else {
				cur = *cur.left
			}
		}
		cur.insert(key, value)
	}
	return true
}

// helper method
func splitMapAndMarkFlag(src map[int32]string) (int32, map[int32]string, map[int32]string) {
	var flag int
	left := make(map[int32]string)
	right := make(map[int32]string)

	keys := make([]int, len(src))

	for k := range src {
		keys = append(keys, int(k))
	}

	sort.Ints(keys)
	length := len(keys)
	flag = length/2

	for i := 0; i < length; i++ {
		key := int32(keys[i])
		if i <= flag {
			left[key] = src[key]
		} else {
			right[key] = src[key]
		}
	}

	return int32(flag), left, right
}

func (root Node) delete(key int32) bool {
	return true
}

func (root Node) traverse(id int32) []string {
	return nil
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
	//elements := context.Send(pid, &Traverse{})
	//fmt.Print(elements)
	console.ReadLine()
}


