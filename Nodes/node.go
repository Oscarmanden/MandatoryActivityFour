package main

import (
	proto "MandatoryActivityFour/grpc"
	"bufio"
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var state = "released"
var MyId int64
var ts int64 = 1
var myPort string = ":50051"
var queue []int64

var id = int64(1)
var peers = map[int64]string{
	2: "localhost:50052",
	3: "localhost:50053",
}
var clients = map[int64]proto.MafClient{} // map med clients hvor der er established connection
// hvor id er key, og value er forbindelsen

// the queue logic for Nodes from https://www.geeksforgeeks.org/go-language/queue-in-go-language/
func enqueue(queue []int64, element int64) []int64 {
	queue = append(queue, element)
	fmt.Println("Enqueued:", element)
	return queue
}

func dequeue(queue []int64) (int64, []int64) {
	element := queue[0]
	if len(queue) == 1 {
		var tmp = []int64{}
		return element, tmp
	}
	return element, queue[1:] // Slice off the element once it is dequeued.
}

type Server struct {
	proto.UnimplementedMafServer
}

func main() {
	// bruger flag så vi ikke behøves flere node filer, men bare kan skrive: go run .\Nodes\node1.go --id=x --port=:xxxxx i tre terminaler
	id := flag.Int("id", 1, "Node ID")
	port := flag.String("port", ":50051", "Port to listen on")
	flag.Parse()
	MyId = int64(*id)
	myPort = *port

	// server logikken:
	listener, err := net.Listen("tcp", myPort)
	if err != nil {
		fmt.Println("failed to listen:", err)
		return
	}
	grpcServer := grpc.NewServer()
	proto.RegisterMafServer(grpcServer, &Server{}) // Det her NodeRequest bliver kaldt fra, hvis vi modtager requests.
	go grpcServer.Serve(listener)

	//Client logikken:
	// Her establisher vi connection til de andre noder og smider dem i mappet clients.
	for peerID, addr := range peers {
		conn, _ := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		clients[peerID] = proto.NewMafClient(conn)
	}

	//skriv wanted i terminal for at starte nodens process for at tilgå cs
	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "wanted" {
			RequestCS()
		}
	}

}

func RequestCS() {
	state = "wanted"
	ts = ts + 1
	SendAndWaitForReplies()
}

func SendAndWaitForReplies() {
	var missingReplies = len(clients)
	for _, cli := range clients {
		resp, _ := cli.NodeRequest(context.Background(), &proto.Request{LamportTime: ts, Nid: MyId})
		if resp.Grant == true {
			missingReplies--
		}
	}
	if missingReplies == 0 {
		csAccess()
	}
}

func csAccess() {
	state = "held"
	fmt.Println(MyId, "has Accessed the Critical Section")
	releaseCs()
}

func releaseCs() {
	state = "released"
	replyQueue()
}

func replyQueue() {
	if len(queue) == 0 {
		fmt.Println("The queue is empty")
		return
	}

	for len(queue) > 0 {
		var peerID int64
		peerID, queue = dequeue(queue)

		replyClient, ok := clients[int64(peerID)]
		if !ok {
			fmt.Println("No reply Nodes found for peer:", peerID)
			continue
		}

		response := &proto.Response{
			Grant: true,
		}

		_, err := replyClient.Reply(context.Background(), response)
		if err != nil {
			fmt.Println("Failed to send reply to:", peerID, ":", err)
		} else {
			fmt.Println("Sent reply to:", peerID)
		}
	}
}

// NodeRequest håndtere når noden modtager requests
func (s *Server) NodeRequest(ctx context.Context, req *proto.Request) (*proto.Response, error) {
	fmt.Println("Received request from node: ", req.Nid)

	ts = max(ts, req.LamportTime) + 1

	if state == "held" || (state == "wanted" && (ts < req.LamportTime || (ts == req.LamportTime && id < req.Nid))) {
		enqueue(queue, req.Nid)
		return &proto.Response{Grant: false}, nil
	}

	fmt.Println("Sending grant access to: ", req.Nid)
	return &proto.Response{Grant: true}, nil
}
