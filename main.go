package main

import (
	"context"
	"fmt"
)

// main just for local usage
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("start")
	res, err := Handler(ctx, struct{}{})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(res.Body)
}
