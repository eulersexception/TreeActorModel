package main

import (
	"flag"
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/log"
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
	switch msg := context.Message().(type) {
	case *messages.CreateResponse:
		fmt.Printf("Tree created! Id =  %v, token = %v\n", msg.GetId(), msg.GetToken())
		state.wg.Done()
	case *messages.DeleteTreeResponse:
		fmt.Printf("Response code %v - tree deletion alert. %v\n", msg.GetCode(), msg.GetMessage())
		state.wg.Done()
	case *messages.ForceTreeDeleteResponse:
		fmt.Printf("Response code %v - tree has been deleted. %v\n", msg.GetCode(), msg.GetMessage())
		state.wg.Done()
	case *messages.InsertResponse:
		fmt.Printf("Response code for insertion %v - %v\n", msg.GetCode(), msg.GetResult())
		state.wg.Done()
	case *messages.SearchResponse:
		fmt.Printf("Response code for search %v - value is %v\n", msg.GetCode(), msg.GetValue())
		state.wg.Done()
	case *messages.DeleteResponse:
		fmt.Printf("Response code for deletion %v - %v\n", msg.GetCode(), msg.GetResult())
		state.wg.Done()
	case *messages.TraverseResponse:
		fmt.Printf("Response code for traversion %v\n - %v\n", msg.GetCode(), msg.GetResult())

		for k, v := range msg.GetPairs() {
			fmt.Printf("{keys: %v, values: %v}\n", k, v)
		}

		state.wg.Done()
	default:
	}
}

const (
	newTree         = "newtree"
	deleteTree      = "deletetree"
	forceTreeDelete = "forTreeDelete"
	insert          = "insert"
	search          = "search"
	delete          = "delete"
	traverse        = "traverse"
)

func main() {
	flagBind := flag.String("bind", "localhost:8090", "Bind to address")
	flagRemote := flag.String("remote", "localhost:8091", "remote host:port")
	flagID := flag.Int("id", -1, "Tree id")
	flagToken := flag.String("token", "", "Tree token")

	flag.Parse()

	flagArgs := flag.Args()

	message, err := getMessage(int32(*flagID), *flagToken, flagArgs)

	if err != nil {
		logError(err)
		return
	}

	if message == nil {
		printHelp()
		return
	}

	remote.SetLogLevel(log.ErrorLevel)

	remote.Start(*flagBind)

	var wg sync.WaitGroup

	props := actor.PropsFromProducer(func() actor.Actor {
		wg.Add(1)
		return &Client{0, &wg}
	})
	rootContext := actor.EmptyRootContext
	pid := rootContext.Spawn(props)

	pidResp, err := remote.SpawnNamed(*flagRemote, "remote", "treeservice", 5*time.Second)

	if err != nil {
		fmt.Printf("Couldn't connect to %s\n", *flagRemote)
		return
	}

	remotePid := pidResp.Pid

	rootContext.RequestWithCustomSender(remotePid, message, pid)

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

func logError(err error) {
	fmt.Printf("An error ocured - %s", err.Error())
}

func getMessage(id int32, token string, args []string) (message interface{}, err error) {
	argsLength := len(args)
	message = &messages.ErrorResponse{"too few arguments - check your command"}
	wrongCredentials := fmt.Sprintf("Id = %v or token = %v invalid", id, token)

	if argsLength == 0 {
		return message, err
	}

	switch args[0] {
	case newTree:
		if argsLength == 2 {
			maxLeafSize, error := strconv.Atoi(args[1])
			if error == nil {
				message = &messages.CreateRequest{Size_: int32(maxLeafSize)}
			}
		}
	case deleteTree:
		if argsLength == 1 {
			if id != -1 && token != "" {
				message = &messages.DeleteTreeRequest{Id: id, Token: token}
			} else {
				message = &messages.ErrorResponse{wrongCredentials}
			}
		}
	case forceTreeDelete:
		if argsLength == 1 {
			if id != -1 && token != "" {
				message = &messages.DeleteTreeRequest{Id: id, Token: token}
			} else {
				message = &messages.ErrorResponse{wrongCredentials}
			}
		}
	case insert:
		if argsLength == 3 {
			key, error := strconv.Atoi(args[1])
			if error != nil {
				response := fmt.Sprintf("invalid input for <key>: %s", args[1])
				message = &messages.ErrorResponse{response}
				break
			}

			value := args[2]

			if id != -1 && token != "" {
				message = &messages.InsertRequest{Id: id, Token: token, Key: int32(key), Value: value, Success: true}
			} else {
				message = messages.ErrorResponse{wrongCredentials}
			}
		}
	case search:
		if argsLength == 2 {
			key, error := strconv.Atoi(args[1])
			if error != nil {
				response := fmt.Sprintf("invalid input for <key>: %s", args[1])
				message = &messages.ErrorResponse{response}
				break
			}

			if id != -1 && token != "" {
				message = &messages.SearchRequest{Id: id, Token: token, Key: int32(key)}
			} else {
				message = messages.ErrorResponse{wrongCredentials}			}
		}
	case delete:
		if argsLength == 2 {
			key, error := strconv.Atoi(args[1])

			if error != nil {
				response := fmt.Sprintf("invalid input for <key>: %s", args[1])
				message = &messages.ErrorResponse{response}
				break
			}

			if id != -1 && token != "" {
				message = &messages.DeleteRequest{Id: id, Token: token, Key: int32(key)}
			} else {
				message = messages.ErrorResponse{wrongCredentials}
			}
		}
	case traverse:
		if argsLength == 1 {
			if id != -1 && token != "" {
				message = &messages.TraverseRequest{Id: id, Token: token}
			} else {
				message = messages.ErrorResponse{wrongCredentials}
			}
		}
	default:
	}

	return message, err
}
