package main

import (
	"fmt"
	"log"
	"net/http"
)

const PORT string = ":8000"

func main() {
	fmt.Println("Servidor corriendo en puerto", PORT)
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hola Mundo")
	})

	log.Fatal(http.ListenAndServe(PORT, mux))
}
