package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Server starting on :8080...")
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, RunSync Pro API!")
	})

	// 8080ポートで待機
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}