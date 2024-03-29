package main

import (
	"flag"
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/remote"
	"github.com/ob-vss-ws19/blatt-3-suedachse/messages"
	"strconv"
	"sync"
	"time"
)

type Client struct {
	count int
	wg    *sync.WaitGroup
}

func (state *Client) Receive(context actor.Context) {
	debug(20, "called Receive()")
	switch msg := context.Message().(type) {
	case *messages.CreateResponse:
		fmt.Printf("Tree created! Id =  %v, token = %v\n", msg.GetId(), msg.GetToken())
		defer state.wg.Done()
	case *messages.DeleteTreeResponse:
		fmt.Printf("Response code %v - tree deletion alert. %v\n", msg.GetCode(), msg.GetMessage())
		defer state.wg.Done()
	case *messages.ForceTreeDeleteResponse:
		fmt.Printf("Response code %v - tree has been deleted. %v\n", msg.GetCode(), msg.GetMessage())
		defer state.wg.Done()
	case *messages.InsertResponse:
		fmt.Printf("Response code for insertion %v - %v\n", msg.GetCode(), msg.GetResult())
		defer state.wg.Done()
	case *messages.SearchResponse:
		fmt.Printf("Response code for search %v - value is %v\n", msg.GetCode(), msg.GetValue())
		defer state.wg.Done()
	case *messages.DeleteResponse:
		fmt.Printf("Response code for deletion %v - %v\n", msg.GetCode(), msg.GetResult())
		defer state.wg.Done()
	case *messages.TraverseResponse:
		fmt.Printf("Response code for traversion %v\n - %v\n", msg.GetCode(), msg.GetResult())

		for k, v := range msg.GetPairs() {
			fmt.Printf("{keys: %v, values: %v}\n", k, v)
		}
		defer state.wg.Done()

	default:
	}
}

const (
	newTree         = "newtree"
	deleteTree      = "deletetree"
	forceTreeDelete = "forcetreedelete"
	insert          = "insert"
	search          = "search"
	delete          = "delete"
	traverse        = "traverse"
)

func main() {
	debug(70, "Defining flags")
	flagBind := flag.String("bind", "127.0.0.1:8090", "Bind to address")
	flagRemote := flag.String("remote", "127.0.0.1:8091", "remote host:port")
	flagID := flag.Int("id", -1, "Tree id")
	flagToken := flag.String("token", "", "Tree token")
	debug(75, "Flags defined -- now parsing")
	flag.Parse()
	debug(77, "flags parsed")

	flagArgs := flag.Args()
	debug(80, fmt.Sprintf("Args = %v", flagArgs))
	message := getMessage(int32(*flagID), *flagToken, flagArgs)

	if message == nil {
		printHelp()
		return
	}

	debug(88, "starting Remote")
	remote.Start(*flagBind)

	var wg sync.WaitGroup
	wg.Add(1)

	props := actor.PropsFromProducer(func() actor.Actor {
		return &Client{0, &wg}
	})
	rootContext := actor.EmptyRootContext
	pid, _ := rootContext.SpawnNamed(props, "treecli")
	debug(99, fmt.Sprintf("created props, spawned them, got PID = %s", pid))

	remote.Register("treecli", props)
	debug(102, "registered Remote")

	pidResp, err := remote.SpawnNamed(*flagRemote, "remote", "treeservice", 5*time.Second)

	if err != nil {
		fmt.Printf("Couldn't connect to %s\n", *flagRemote)
		return
	}

	remotePid := pidResp.Pid
	debug(112, fmt.Sprintf("got Remote PID = %s", remotePid))

	rootContext.RequestWithCustomSender(remotePid, message, pid)
	//rootContext.RequestFuture(remotePid, message, 5*time.Second)

	debug(117, fmt.Sprintf("Send message from treecli PID %s to treeservice PID %s: \"%s\"", pid, remotePid, message))

	wg.Wait()
}

func printHelp() {
	help := "\n-----------------------------------------------------\n\n" +
		"  This is a demonstration of distributed software systems by \n" +
		"  an implementation of the \"Remote Actor Model\".\n" +
		"  By using listed commands you can create a tree to store key-value pairs. \n\n" +
		"  Keys are of type integer and values of type string. \n\n" +
		"  Create new tree:\n" +
		"    treecli newtree <max number of key-value-pairs>\n\n" +
		"  Commands on existing trees:\n" +
		"    treecli --id=<treeID> --token=<token> <command> <key> <value>\n\n" +
		"  Possible commands and parameters:\n" +
		"    " + insert + " <key> <value>\n" +
		"    " + search + " <key>\n" +
		"    " + delete + " <key>\n" +
		"    " + deleteTree + "\n" +
		"    " + forceTreeDelete + "\n" +
		"    " + traverse + "\n"
	fmt.Print(help)
}

func debug(line int, info string) {
	fmt.Printf("TreeCli :: Line %v  --> %v\n", line, info)
}

func getMessage(id int32, token string, args []string) (message interface{}) {
	argsLength := len(args)
	message = &messages.ErrorResponse{Message: "too few arguments - check your command"}
	wrongCredentials := fmt.Sprintf("Id = %v or token = %v invalid", id, token)

	debug(151, fmt.Sprintf("getMessage(%v, %v) with default message \"to few arguments\"", id, token))

	if argsLength == 0 {
		return message
	}

	switch args[0] {
	case newTree:
		debug(159, "switched to case newTree")
		if argsLength == 2 {
			maxLeafSize, error := strconv.Atoi(args[1])
			if error == nil {
				debug(163, "preparing CreateRequest")
				message = &messages.CreateRequest{Code: int32(maxLeafSize)}
			}
		}
	case deleteTree:
		if argsLength == 1 {
			if id != -1 && token != "" {
				debug(170, "preparing DeleteRequest")
				message = &messages.DeleteTreeRequest{Id: id, Token: token}
			} else {
				debug(173, "preparing ErrorResponse")
				message = &messages.ErrorResponse{Message: wrongCredentials}
			}
		}
	case forceTreeDelete:
		if argsLength == 1 {
			if id != -1 && token != "" {
				debug(180, "preparing ForceTreeDeleteRequest")
				message = &messages.ForceTreeDeleteRequest{Id: id, Token: token}
			} else {
				debug(183, "preparing ErrorResponse")
				message = &messages.ErrorResponse{Message: wrongCredentials}
			}
		}
	case insert:
		if argsLength == 3 {
			key, error := strconv.Atoi(args[1])
			if error != nil {
				debug(191, "preparing ErrorResponse")
				response := fmt.Sprintf("invalid input for <key>: %s", args[1])
				message = &messages.ErrorResponse{Message: response}

				break
			}

			value := args[2]

			if id != -1 && token != "" {
				debug(201, "preparing InsertRequest")
				message = &messages.InsertRequest{Id: id, Token: token, Key: int32(key), Value: value, Success: true, Ip: "127.0.0.1", Port: 8090}
			} else {
				debug(204, "preparing ErrorResponse")
				message = messages.ErrorResponse{Message: wrongCredentials}
			}
		}
	case search:
		if argsLength == 2 {
			key, error := strconv.Atoi(args[1])
			if error != nil {
				debug(212, "preparing ErrorResponse")
				response := fmt.Sprintf("invalid input for <key>: %s", args[1])
				message = &messages.ErrorResponse{Message: response}

				break
			}

			if id != -1 && token != "" {
				debug(220, "preparing SearchRequest")
				message = &messages.SearchRequest{Id: id, Token: token, Key: int32(key)}
			} else {
				debug(223, "preparing ErrorResponse")
				message = messages.ErrorResponse{Message: wrongCredentials}
			}
		}
	case delete:
		if argsLength == 2 {
			key, error := strconv.Atoi(args[1])

			if error != nil {
				debug(232, "preparing ErrorResponse")
				response := fmt.Sprintf("invalid input for <key>: %s", args[1])
				message = &messages.ErrorResponse{Message: response}

				break
			}

			if id != -1 && token != "" {
				debug(240, "preparing DeleteRequest")
				message = &messages.DeleteRequest{Id: id, Token: token, Key: int32(key)}
			} else {
				debug(243, "preparing ErrorResponse")
				message = messages.ErrorResponse{Message: wrongCredentials}
			}
		}
	case traverse:
		if argsLength == 1 {
			if id != -1 && token != "" {
				debug(250, "preparing TraverseRequest")
				message = &messages.TraverseRequest{Id: id, Token: token}
			} else {
				debug(253, "preparing ErrorResponse")
				message = messages.ErrorResponse{Message: wrongCredentials}
			}
		}
	default:
	}

	return message
}
