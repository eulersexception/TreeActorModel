package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/log"
	"github.com/AsynkronIT/protoactor-go/remote"
	"github.com/ob-vss-ws19/blatt-3-suedachse/messages"
	"github.com/ob-vss-ws19/blatt-3-suedachse/tree"
	"math/rand"
	"sync"
	"time"
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
	token := "Node_created_at_" + time.Millisecond.String()
	debug(31, fmt.Sprintf("createIdAndToken() called -> ID = %v, token = %v", id, token))

	return id, token
}

func (server Server) getTree(id int32, token string) (Tree, error) {
	debug(37, fmt.Sprintf("getTree(%v, %v) called", id, token))
	for _, v := range server.trees {
		if v.id == id {
			if v.token == token {
				debug(41, fmt.Sprintf("returning from getTree() with %v", v.root.String()))
				return v, nil
			}
			debug(44, "returning from getTree() with error")
			return Tree{}, errors.New("token mismatch")
		}
	}

	debug(49, "returning from getTree() with error")
	return Tree{}, errors.New("no tree with given ID")
}

func debug(line int, info string) {
	fmt.Printf("TreeService :: Line %v  --> %v\n", line, info)
}

func (server *Server) Receive(c actor.Context) {
	debug(57, "called Receive()")
	switch msg := c.Message().(type) {
	case *messages.CreateRequest:
		debug(61, "preparing CreateResponse")
		idx, tokenx := createIDAndToken()
		props := actor.PropsFromProducer(func() actor.Actor {
			return &tree.Node{MaxSize: msg.Code, IsLeaf: true, KeyValues: make(map[int32]string)}
		})

		pid := c.Spawn(props)
		server.trees = append(server.trees, Tree{idx, tokenx, pid})

		debug(68, "created tree and appended it to map - preparing CreateResponse")
		c.Send(c.Sender(), &messages.CreateResponse{
			Id:    idx,
			Token: tokenx,
		})
	case *messages.DeleteTreeRequest:
		debug(76, "preparing DeleteTreeResponse")
		_, err := server.getTree(msg.Id, msg.Token)

		if err != nil {
			c.Respond(&messages.DeleteTreeResponse{Code: 404, Message: err.Error()})
		} else {
			force := "Trigger \"forcetreedelete\" to delete tree permanently - take care... "
			c.Respond(&messages.DeleteTreeResponse{Code: 200, Message: force})
		}
	case *messages.ForceTreeDeleteRequest:
		debug(86, "preparing ForceTreeDeleteResponse")
		tree, err := server.getTree(msg.Id, msg.Token)

		if err != nil {
			c.Respond(&messages.ForceTreeDeleteResponse{Code: 404, Message: err.Error()})
		} else {
			for i := 0; i < len(server.trees); i++ {
				if server.trees[i] == tree {
					server.trees = append(server.trees[:i], server.trees[i+1:]...)
				}
			}
			message := "Tree delete successfully"
			c.Respond(&messages.ForceTreeDeleteResponse{Code: 200, Message: message})
		}
	case *messages.InsertRequest:
		debug(103, "preparing InsertResponse")
		tree, err := server.getTree(msg.Id, msg.Token)

		if err != nil {
			c.Respond(&messages.InsertResponse{Code: 404, Result: err.Error()})
		} else {
			c.RequestWithCustomSender(tree.root, msg, c.Sender())
		}
	case *messages.SearchRequest:
		debug(112, "preparing SearchResponse")
		tree, err := server.getTree(msg.Id, msg.Token)

		if err != nil {
			c.Respond(&messages.SearchResponse{Code: 404, Value: err.Error()})
		} else {
			c.RequestWithCustomSender(tree.root, msg, c.Sender())
		}
	case *messages.DeleteRequest:
		debug(121, "preparing DeleteResponse")
		tree, err := server.getTree(msg.Id, msg.Token)

		if err != nil {
			c.Respond(&messages.DeleteResponse{Code: 404, Result: err.Error()})
		} else {
			c.RequestWithCustomSender(tree.root, msg, c.Sender())
		}
	case *messages.TraverseRequest:
		debug(130, "preparing TraverseResponse")
		tree, err := server.getTree(msg.Id, msg.Token)

		if err != nil {
			c.Respond(&messages.TraverseResponse{Code: 404, Result: err.Error()})
		} else {
			c.RequestWithCustomSender(tree.root, msg, c.Sender())
		}
	default:
		debug(138, "check format of commands")
	}
}

func main() {
	var wg sync.WaitGroup

	wg.Add(1)

	defer wg.Wait()

	flagBind := flag.String("bind", "127.0.0.1:8091", "Bind to address")
	flag.Parse()

	remote.SetLogLevel(log.ErrorLevel)
	remote.Start(*flagBind)
	remote.Register("treeservice", actor.PropsFromProducer(func() actor.Actor {
		return &Server{[]Tree{}}
	}))

	fmt.Println(" ----> TreeService up <-----")
}
