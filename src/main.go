package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/model", modelHandler)
	http.HandleFunc("/content", contentHandler)

	log.Println("Lisenting on 5000")
	err := http.ListenAndServe(":5000", nil)
	if err != nil {
		log.Fatalln("ListenAndServe: ", err)
	}
}
