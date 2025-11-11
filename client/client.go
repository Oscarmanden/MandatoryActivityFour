package main

import (
	proto "MandatoryActivityFour/grpc"
	"context"
	"fmt"
	"net"

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
var clients = map[int64]proto.MafClient{}

// the queue logic for client from https://www.geeksforgeeks.org/go-language/queue-in-go-language/
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
	listener, err := net.Listen("tcp", myPort)
	if err != nil {
		fmt.Println("failed to listen:", err)
		return
	}
	grpcServer := grpc.NewServer()

	proto.RegisterMafServer(grpcServer, &Server{})
	go grpcServer.Serve(listener)

	//connect to peers
	for peerID, addr := range peers {
		conn, _ := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		clients[peerID] = proto.NewMafClient(conn)
	}

	RequestCS()
}

func RequestCS() {
	state = "wanted"
	ts = ts + 1
	SendAndWaitForReplies()
}

func SendAndWaitForReplies() {
	var replies = 2
	for pid, cli := range clients {
		if pid == MyId {
			continue
		}
		resp, _ := cli.NodeRequest(context.Background(), &proto.Request{LamportTime: ts, Nid: MyId})
		if resp.Grant == true {
			replies--
		}
	}
	csAccess()
}

func csAccess() {
	fmt.Println(MyId, "has Accessed the Critical Section")
	releaseCs()
}

func releaseCs() {
	state = "released"
	replyQueue()
}

func replyQueue() {
	for len(queue) > 0 {
		var peerID int64
		peerID, queue = dequeue(queue)

		replyClient, ok := clients[int64(peerID)]
		if !ok {
			fmt.Println("No reply client found for peer", peerID)
			continue
		}

		response := &proto.Response{
			Grant: true,
		}

		_, err := replyClient.Reply(context.Background(), response)
		if err != nil {
			fmt.Println("Failed to send reply to", peerID, ":", err)
		} else {
			fmt.Println("Sent reply to", peerID)
		}
	}
}

func replyToRequest() {

}

func onReceiveRequest(req *proto.Request) {
	reqClient := req.Nid
	reqTimestamp := req.LamportTime
	if state == "held" || (state == "wanted" && ((ts < reqTimestamp) && (id < req.Nid))) {
		enqueue(queue, reqClient)
	} else {
		//reply() // to reqClient
	}
}

func (s *Server) NodeRequest(ctx context.Context, req *proto.Request) (*proto.Response, error) {
	fmt.Println("Received request from node ", req.Nid)

	// Update Lamport clock
	if req.LamportTime > ts {
		ts = req.LamportTime
	}
	ts++
	onReceiveRequest(req)

	if len(queue) > 0 {
		replyQueue()
		return &proto.Response{}, nil
	}
	return &proto.Response{Grant: true}, nil
}

func Reply(ctx context.Context, req *proto.RecievedResponseButEmpty) {

}
