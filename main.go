package main

import "fmt"

func main() {
	fmt.Println("start")
	res, err := Handler()

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(res.Body)
}
