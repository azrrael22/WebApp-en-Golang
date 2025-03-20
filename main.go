package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

//const PORT string = ":8000"

func main() {

	var puerto string
	flag.StringVar(&puerto, "p", "localhost:8000", "numero de puerto")

	flag.Parse()

	fmt.Println("Servidor corriendo en puerto", puerto)
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hola Mundo")
	})

	log.Fatal(http.ListenAndServe(puerto, mux))
}
