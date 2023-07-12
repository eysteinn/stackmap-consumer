package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"

	"test/api"
	"test/prog"
)

func main() {
	fmt.Println("Hey")
	if len(os.Args) > 1 {
		err := prog.Run()
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	api.Run()

}
