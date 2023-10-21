package main

import "net/http"

func main() {

}

func setupAPI() {
	http.Handle("/", http.FileServer(http.Dir("./frontend")))
}
