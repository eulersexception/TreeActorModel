package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/log"
	"github.com/AsynkronIT/protoactor-go/remote"
	"github.com/ob-vss-ws19/blatt-3-suedachse/messages"
	"github.com/ob-vss-ws19/blatt-3-suedachse/tree"
)

type Tree struct {
	id    int32
	token string
	root  *actor.PID
}

type Server struct {
	trees []Tree
}

func createIDAndToken() (int32, string) {
	rand.Seed(time.Now().UnixNano())
	id := rand.Int31n(10000000)
	token := "Node_created_at_" + time.Now().String()

	return id, token
}

func (server Server) getTree(id int32, token string) (Tree, error) {
	for _, v := range server.trees {
		if v.id == id {
			if v.token == token {
				return v, nil
			}

			return Tree{}, errors.New("token mismatch")
		}
	}

	return Tree{}, errors.New("no tree with given ID")
}

func (server *Server) Receive(c actor.Context) {
	switch msg := c.Message().(type) {
	case *messages.CreateRequest:
		idx, tokenx := createIDAndToken()
		props := actor.PropsFromProducer(func() actor.Actor {

			return &tree.Node{MaxSize: msg.Size_, IsLeaf: true, KeyValues: make(map[int32]string)}
		})

		pid := c.Spawn(props)
		server.trees = append(server.trees, Tree{idx, tokenx, pid})

		c.Respond(&messages.CreateResponse{
			Id:    idx,
			Token: tokenx,
		})
	case *messages.DeleteTreeRequest:
		_, err := server.getTree(msg.Id, msg.Token)

		if err != nil {
			c.Respond(&messages.DeleteTreeResponse{Code: 404, Message: err.Error()})
		} else {
			force := "Trigger \"ForceTreeDeleteRequest\" to delete tree ultimately - take care... "
			c.Respond(&messages.DeleteTreeResponse{Code: 200, Message: force})
		}
	case *messages.ForceTreeDeleteRequest:
		tree, err := server.getTree(msg.Id, msg.Token)

		if err != nil {
			c.Respond(&messages.ForceTreeDeleteResponse{Code: 404, Message: err.Error()})
		} else {
			for i := 0; i < len(server.trees); i++ {
				if server.trees[i] == tree {
					server.trees = append(server.trees[:i], server.trees[i+1:]...)

					break
				}
			}
			message := "Tree delete successfully"
			c.Respond(&messages.ForceTreeDeleteResponse{Code: 200, Message: message})
		}
	case *messages.InsertRequest:
		tree, err := server.getTree(msg.Id, msg.Token)

		if err != nil {
			c.Respond(&messages.InsertResponse{Code: 404, Result: err.Error()})
		} else {
			c.RequestWithCustomSender(tree.root, msg, c.Sender())
		}
	case *messages.SearchRequest:
		tree, err := server.getTree(msg.Id, msg.Token)

		if err != nil {
			c.Respond(&messages.SearchResponse{Code: 404, Value: err.Error()})
		} else {
			c.RequestWithCustomSender(tree.root, msg, c.Sender())
		}
	case *messages.DeleteRequest:
		tree, err := server.getTree(msg.Id, msg.Token)

		if err != nil {
			c.Respond(&messages.DeleteResponse{Code: 404, Result: err.Error()})
		} else {
			c.RequestWithCustomSender(tree.root, msg, c.Sender())
		}
	case *messages.TraverseRequest:
		tree, err := server.getTree(msg.Id, msg.Token)

		if err != nil {
			c.Respond(&messages.TraverseResponse{Code: 404, Result: err.Error()})
		} else {
			c.RequestWithCustomSender(tree.root, msg, c.Sender())
		}
	default:
	}
}

func main() {
	var wg sync.WaitGroup

	wg.Add(1)

	defer wg.Wait()

	flagBind := flag.String("bind", "localhost:8091", "Bind to address")
	flag.Parse()

	remote.SetLogLevel(log.ErrorLevel)
	remote.Start(*flagBind)
	remote.Register("treeservice", actor.PropsFromProducer(func() actor.Actor {
		return &Server{[]Tree{}}
	}))

	fmt.Println(" ----> TreeService up <-----")
}
