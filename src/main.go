package main

import (
	"log"

	_ "github.com/lib/pq"

	"test/api"
)

func main() {
	log.Println("Stackmap-Consumer")
	/*if len(os.Args) > 1 {
		err := prog.Run()
		if err != nil {
			log.Fatal(err)
		}
		return
	}*/
	api.Run()

}
