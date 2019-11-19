package tree

import (
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/remote"
	"github.com/ob-vss-ws19/blatt-3-suedachse/messages"
	"log"
	"math"
	"sort"
	"time"
)

type Node struct {
	left      *actor.PID
	right     *actor.PID
	IsLeaf    bool
	MaxSize   int32
	MaxLeft   int32
	KeyValues map[int32]string
}

func (node Node) Receive(context actor.Context) {
	remote.Start("127.0.0.1:8092")
	remoteActor := context.Sender()
	switch msg := context.Message().(type) {
	case *messages.InsertRequest:
		if node.IsLeaf {
			node.KeyValues[msg.Key] = msg.Value

			if int32(len(node.KeyValues)) > node.MaxSize {
				node.IsLeaf = false

				props := actor.PropsFromProducer(func() actor.Actor {
					return &Node{MaxSize: node.MaxSize, IsLeaf: true, KeyValues: make(map[int32]string)}
				})

				node.left = context.Spawn(props)
				node.right = context.Spawn(props)

				middle := int(math.Ceil(float64(len(node.KeyValues)) / 2.0))

				var keys []int

				for k := range node.KeyValues {
					keys = append(keys, int(k))
				}

				sort.Ints(keys)

				node.MaxLeft = int32(keys[middle-1])

				for i, k := range keys {
					message := &messages.InsertRequest{
						Id:    msg.Id,
						Token: msg.Token,
						Key:   int32(k),
						Value: node.KeyValues[int32(k)],
					}

					if k == int(msg.Key) {
						message.Success = true
					} else {
						message.Success = false
					}

					if i < middle {
						context.RequestWithCustomSender(node.left, message, remoteActor)
					} else {
						context.RequestWithCustomSender(node.right, message, remoteActor)
					}
				}
				// Delete map because no leaf anymore
				node.KeyValues = nil
			} else if int32(len(node.KeyValues)) <= node.MaxSize { // If not full, send response
				message := fmt.Sprintf("Insertion completed: {key: %d, value: %s}", msg.Key, msg.Value)

				log.Println(message)

				response := &messages.InsertResponse{
					Code:   200,
					Result: message,
				}

				context.Send(remoteActor, response)
			}
		} else { // If node, send request to the proper leaf
			if msg.Key > node.MaxLeft {
				context.RequestWithCustomSender(node.right, msg, remoteActor)
			} else {
				context.RequestWithCustomSender(node.left, msg, remoteActor)
			}
		}
	case *messages.SearchRequest:
		if node.IsLeaf {
			value := node.KeyValues[msg.Key]

			var message string

			if value == "" {
				context.Send(remoteActor, &messages.SearchResponse{
					Code:500,
					Value: fmt.Sprintf("Key %d does not exist", msg.Key),
				})
			}

			log.Println(message)

			context.Send(remoteActor, &messages.SearchResponse{
				Code:   200,
				Value: value,
			})
		} else {
			if msg.Key > node.MaxLeft {
				context.RequestWithCustomSender(node.right, msg, context.Sender())
			} else {
				context.RequestWithCustomSender(node.left, msg, context.Sender())
			}
		}
	case *messages.DeleteRequest:
		if node.IsLeaf {
			value := node.KeyValues[msg.Key]

			var message string

			if value != "" {
				delete(node.KeyValues, msg.Key)
				message = fmt.Sprintf("Pair {key: %d, value: %s} deleted", msg.Key, value)
			} else {
				message = fmt.Sprintf("Key %d does not exist", msg.Key)
			}

			log.Println(message)

			context.Send(remoteActor, &messages.DeleteResponse{
				Code:   200,
				Result: message,
			})
		} else {
			if msg.Key > node.MaxLeft {
				context.RequestWithCustomSender(node.right, msg, context.Sender())
			} else {
				context.RequestWithCustomSender(node.left, msg, context.Sender())
			}
		}
	case *messages.TraverseRequest:
		// If it's a Node send the traverse request to the leafs and wait for their responses
		if !node.IsLeaf {
			// Send messages to leafs and set the future timeout to 5 seconds
			leftFuture := context.RequestFuture(node.left, msg, 5*time.Second)
			rightFuture := context.RequestFuture(node.right, msg, 5*time.Second)

			// Take the results and break if any of the leafs timed out
			leftResult, errLeft := leftFuture.Result()

			if errLeft != nil {
				context.Send(remoteActor, &messages.TraverseResponse{
					Code:   500,
					Result: "Left leaf timed out",
					Pairs:  nil,
				})
				log.Println("Left leaf timed out")

				break
			}

			rightResult, errRight := rightFuture.Result()

			if errRight != nil {
				context.Send(remoteActor, &messages.TraverseResponse{
					Code:   500,
					Result: "Right leaf timed out",
					Pairs:  nil,
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
				context.Send(remoteActor, &messages.TraverseResponse{
					Code:   500,
					Result: "invalid type",
					Pairs:  nil,
				})
				log.Println("invalid type")
			}

			switch res := rightResult.(type) {
			case *messages.TraverseResponse:
				for _, el := range res.Pairs {
					pairs = append(pairs, &messages.Pair{Key: el.Key, Value: el.Value})
				}
			default:
				context.Send(remoteActor, &messages.TraverseResponse{
					Code:   500,
					Result: "invalid type",
					Pairs:  nil,
				})
				log.Println("invalid type")
			}

			context.Send(remoteActor, &messages.TraverseResponse{
				Code:   200,
				Result: "OK",
				Pairs:  pairs,
			})
		} else {
			// sorting pairs by keys if IsLeaf
			var keysInt []int

			pairs := make([]*messages.Pair, 0)

			for k := range node.KeyValues {
				keysInt = append(keysInt, int(k))
			}

			sort.Ints(keysInt)

			for _, k := range keysInt {
				pairs = append(pairs, &messages.Pair{Key: int32(k), Value: node.KeyValues[int32(k)]})
			}

			context.Send(remoteActor, &messages.TraverseResponse{
				Code:   200,
				Result: "OK",
				Pairs:  pairs,
			})
		}
	default:
	}
}
