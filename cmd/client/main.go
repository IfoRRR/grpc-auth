package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	demov1 "portfolio/grpc-auth/gen/demo/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	mode := "both"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	switch mode {
	case "rest":
		demoREST()
	case "grpc":
		demoGRPC()
	default:
		demoREST()
		fmt.Println("---")
		demoGRPC()
	}
}

func demoREST() {
	body, _ := json.Marshal(map[string]string{"title": "learn gRPC via REST"})
	resp, err := http.Post("http://127.0.0.1:8088/tasks", "application/json", bytes.NewReader(body))
	must(err)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	fmt.Printf("REST create [%d]: %s", resp.StatusCode, raw)

	resp2, err := http.Get("http://127.0.0.1:8088/tasks")
	must(err)
	defer resp2.Body.Close()
	raw2, _ := io.ReadAll(resp2.Body)
	fmt.Printf("REST list   [%d]: %s", resp2.StatusCode, raw2)
}

func demoGRPC() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := grpc.NewClient("127.0.0.1:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	must(err)
	defer conn.Close()

	c := demov1.NewTaskServiceClient(conn)
	created, err := c.CreateTask(ctx, &demov1.CreateTaskRequest{Title: "learn gRPC via gRPC"})
	must(err)
	fmt.Printf("gRPC create: id=%s title=%q\n", created.Task.Id, created.Task.Title)

	listed, err := c.ListTasks(ctx, &demov1.ListTasksRequest{})
	must(err)
	fmt.Printf("gRPC list:   %d tasks\n", len(listed.Tasks))
	for _, t := range listed.Tasks {
		fmt.Printf("  - %s %q\n", t.Id, t.Title)
	}
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
