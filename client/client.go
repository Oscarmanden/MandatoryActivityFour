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
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var state = "released"
var MyId int64 = 1
var ts int64 = 1
var myPort string = ":50051"
var myReqTs int64

var peers = map[int64]string{
	1: "localhost:50051",
	2: "localhost:50052",
	3: "localhost:50053",
}
var clients = map[int64]proto.MafClient{}

func hasPriority(myTs, myId, otherTs, otherId int64) bool {
	if myTs < otherTs {
		return true
	}
	if myTs == otherTs && myId < otherId {
		return true
	}
	return false
}

type Server struct {
	proto.UnimplementedMafServer
}

func main() {
	idFlag := flag.Int64("id", 1, "my node id (1..3)")
	flag.Parse()
	MyId = *idFlag
	myPort = peers[MyId]

	listener, err := net.Listen("tcp", myPort)
	if err != nil {
		fmt.Println("failed to listen:", err)
		return
	}
	grpcServer := grpc.NewServer()
	proto.RegisterMafServer(grpcServer, &Server{})
	go grpcServer.Serve(listener)

	// connect to peers
	for pid, addr := range peers {
		if pid == MyId {
			continue
		}
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Println("dial failed:", pid, addr, err)
			continue
		}
		clients[pid] = proto.NewMafClient(conn)
	}
	time.Sleep(500 * time.Millisecond)

	fmt.Println("Node", MyId, "listening on", myPort)
	fmt.Println("Type 'request' to try the CS, or 'quit' to exit.")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		cmd := strings.TrimSpace(strings.ToLower(scanner.Text()))
		switch cmd {
		case "request", "r":
			fmt.Println(MyId, "is requesting access to Critical Section")
			time.Sleep(5000 * time.Millisecond)
			RequestCS()
		case "quit", "exit", "q":
			fmt.Println("bye")
			return
		default:
			fmt.Println("commands: request | quit")
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("stdin error:", err)
	}
}

func RequestCS() {
	state = "wanted"
	ts++
	myReqTs = ts

	for pid, cli := range clients {
		if pid == MyId || cli == nil {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		_, err := cli.NodeRequest(ctx, &proto.Request{LamportTime: myReqTs, Nid: MyId})
		cancel()
		if err != nil {
			fmt.Println("request to", pid, "failed:", err)
		}
	}

	state = "held"
	csAccess()
}

func csAccess() {
	fmt.Println(MyId, "has Accessed the Critical Section")
	releaseCs()
}

func releaseCs() {
	state = "released"
}
func (s *Server) NodeRequest(ctx context.Context, req *proto.Request) (*proto.Response, error) {
	fmt.Println("Received request from node", req.Nid)

	if req.LamportTime > ts {
		ts = req.LamportTime
	}
	ts++
	for state == "held" || (state == "wanted" && hasPriority(myReqTs, MyId, req.LamportTime, req.Nid)) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}

	// Grant
	return &proto.Response{Grant: true}, nil
}
