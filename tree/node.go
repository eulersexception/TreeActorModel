package main

import (
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"log"
	"math"
	"sort"
	"github.com/eulersexception/vss/blatt-3-suedachse/messages"
	"time"
)

type Note struct {
	left		*actor.PID
	right		*actor.PID
	isLeaf		bool
	maxSize 	int32
	maxLeft 	int32
	keyValues 	map[int32]string
}

func (node Note) Receive(context actor.Context) {
	switch msg := context.Message().(type) {

	// Insert ----------------------------------------------------------------------------------------------------------
	case *messages.InsertRequest:
		if node.isLeaf {
			// Insert the key value pair
			node.keyValues[msg.Key] = msg.Value
			// If leaf is full, split leaf
			if int32(len(node.keyValues)) > node.maxLeft {
				// No leaf anymore
				node.isLeaf = false
				// Create two leafs
				props := actor.PropsFromProducer(func() actor.Actor {
					return &Node{maxSize: node.maxSize, isLeaf: true, key_values: make(map[int32]string)}
				})
				node.left = context.Spawn(props)
				node.right = context.Spawn(props)
				middle := int(math.Ceil(float64(len(node.keyValues)) / 2.0))
				var keys []int
				for k := range node.keyValues {
					keys = append(keys, int(k))
				}
				sort.Ints(keys)
				node.maxLeft = int32(keys[middle-1])

				for i, k := range keys {

					message := &messages.InsertRequest{
						Id:    msg.Id,
						Token: msg.Token,
						Key:   int32(k),
						Value: node.keyValues[int32(k)],
					}

					if k == int(msg.Key) {
						message.Success = true
					} else {
						message.Success = false
					}
					if i < middle {
						context.RequestWithCustomSender(node.left, message, context.Sender())
					} else {
						context.RequestWithCustomSender(node.right, message, context.Sender())
					}
				}
				// Delete map because no leaf anymore
				node.keyValues = nil

				// If not full, send response
			} else if msg.Success {
				message := fmt.Sprintf("Insertion completed: {key: %d, value: %s}", msg.Key, msg.Value)
				log.Println(message)
				context.Respond(&messages.InsertResponse{
					Code:    200,
					Result: message,
				})
			}
			// If node, send request to the proper leaf
		} else {
			if msg.Key > node.maxLeft {
				context.RequestWithCustomSender(node.right, msg, context.Sender())
			} else {
				context.RequestWithCustomSender(node.left, msg, context.Sender())
			}
		}

	// Search ----------------------------------------------------------------------------------------------------------
	case *messages.SearchRequest:
		if node.isLeaf {
			value := node.keyValues[msg.Key]
			var message string
			if value != "" {
				message = fmt.Sprintf("Value found: {key: %d, value: %s}", msg.Key, value)
			} else {
				message = fmt.Sprintf("Key %d does not exist", msg.Key)
			}
			log.Println(message)
			context.Respond(&messages.InsertResponse{
				Code:    200,
				Result: message,
			})
		} else {
			if msg.Key > node.maxLeft {
				context.RequestWithCustomSender(node.right, msg, context.Sender())
			} else {
				context.RequestWithCustomSender(node.left, msg, context.Sender())
			}
		}

	// Delete ----------------------------------------------------------------------------------------------------------
	case *messages.DeleteRequest:
		if node.isLeaf {
			value := node.keyValues[msg.Key]
			var message string
			if value != "" {
				delete(node.keyValues, msg.Key)
				message = fmt.Sprintf("Pair {key: %d, value: %s} deleted", msg.Key, value)
			} else {
				message = fmt.Sprintf("Key %d does not exist", msg.Key)
			}
			log.Println(message)
			context.Respond(&messages.InsertResponse{
				Code:    200,
				Result: message,
			})
		} else {
			if msg.Key > node.maxLeft {
				context.RequestWithCustomSender(node.right, msg, context.Sender())
			} else {
				context.RequestWithCustomSender(node.left, msg, context.Sender())
			}
		}

	// Traverse	--------------------------------------------------------------------------------------------------------
	case *messages.TraverseRequest:
		// If it's a Node send the traverse request to the leafs and wait for their responses
		if !node.isLeaf {
			// Send messages to leafs and set the future timeout to 5 seconds
			leftFuture := context.RequestFuture(node.left, msg, 5*time.Second)
			rightFuture := context.RequestFuture(node.right, msg, 5*time.Second)

			// Take the results and break if any of the leafs timed out
			leftResult, errLeft := leftFuture.Result()

			if errLeft != nil {
				context.Respond(&messages.TraverseResponse{
					Code:    500,
					Result: "Left leaf timed out",
					Pairs:    nil,
				})
				log.Println("Left leaf timed out")
				break
			}
			rightResult, errRight := rightFuture.Result()
			if errRight != nil {
				context.Respond(&messages.TraverseResponse{
					Code:    500,
					Result: "Right leaf timed out",
					Pairs:	nil,
				})
				log.Println("Right leaf timed out")
				break
			}

			pairs := make([]*messages.Pair, 0)

			// Send error if it's something different than a slice of pairs
			switch res := leftResult.(type) {
			case *messages.TraverseResponse:
				for _, el := range res.Pairs {
					pairs = append(pairs, &messages.Pair{Key: el.Key, Value: el.Value})
				}
			default:
				context.Respond(&messages.TraverseResponse{
					Code:    500,
					Result: "invalid type",
					Pairs:    nil,
				})
				log.Println("invalid type")
			}

			switch res := rightResult.(type) {
			case *messages.TraverseResponse:
				for _, el := range res.Pairs {
					pairs = append(pairs, &messages.Pair{Key:el.Key, Value: el.Value})
				}
			default:
				context.Respond(&messages.TraverseResponse{
					Code:    500,
					Result: "invalid type",
					Pairs:    nil,
				})
				log.Println("invalid type")
			}

			context.Respond(&messages.TraverseResponse{
				Code:    200,
				Result: "OK",
				Pairs:	pairs,
			})
		} else {
			// sorting pairs by keys if isLeaf

			var keysInt []int
			pairs := make([]*messages.Pair, 0)

			for k := range node.keyValues {
				keysInt = append(keysInt, int(k))
			}

			sort.Ints(keysInt)

			for _, k := range keysInt {
				pairs = append(pairs, &messages.Pair{Key: int32(k), Value: node.keyValues[int32(k)]})
			}

			context.Respond(&messages.TraverseResponse{
				Code:    200,
				Result: "OK",
				Pairs:   pairs,
			})
		}
	default:
	}
}
