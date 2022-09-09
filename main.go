// PowerStatus project main.go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Listening \nhttp://localhost:10042")

	fs := http.FileServer(http.Dir("static"))

	http.Handle("/powerstatus/static/", http.StripPrefix("/powerstatus/static/", fs))
	http.HandleFunc("/powerstatus/setstatus", setStatus)
	http.HandleFunc("/powerstatus/status", displayStatus)
	http.HandleFunc("/powerstatus", displayStatus)
	http.HandleFunc("/powerstatus/", displayStatus)
	http.HandleFunc("/", displayStatus)

	err := http.ListenAndServe(":10042", nil)
	if err != nil {
		fmt.Println("Error while listening: " + err.Error())
	}
}
