package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var p = fmt.Println

func main() {
	http.HandleFunc("/Sum", Sum)
	http.HandleFunc("/Sum2", Sum2)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

type test_struct struct {
	Test string
}

// Sum will add the total of n integers
func Sum(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/Sum" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprint(w, "Hello, World! This is sum")
	//resp, err := http.Get("https://gae-by-endpoint.uc.r.appspot.com/Sum2")
	//if err != nil {
	//	fmt.Fprint(w, err.Error())
	//}

	r.ParseForm()
	log.Println(r.Form)
	//LOG: map[{"test": "that"}:[]]
	var t test_struct
	for key, _ := range r.Form {
		log.Println(key)
		//LOG: {"test": "that"}
		err := json.Unmarshal([]byte(key), &t)
		if err != nil {
			log.Println(err.Error())
		}
	}
	log.Println(t.Test)
}

func Sum2(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/Sum2" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprint(w, "Hello, World! This is sum2")
	ans := 10
	json.NewEncoder(w).Encode(ans)
}

func SequencialSum(arr []int64) int64 {
	total := int64(0)
	for _, v := range arr {
		total = total + int64(v)
		time.Sleep(time.Millisecond)
	}
	return total
}

func ParrellelSum(arr []int64, thresh int, start int, end int, result chan int64) {
	if end - start <= thresh {
		result <- SequencialSum(arr[start:end])
		return
	}

	subResult := make(chan int64)
	subResult2 := make(chan int64)

	go func(arr []int64, thresh int, start int, end int) {
		p("Goroutine spawned")
		ParrellelSum(arr, thresh, start, (start+end)/2, subResult)
	}(arr, thresh, start, end)

	go func(arr []int64, thresh int, start int, end int) {
		p("Goroutine spawned")
		ParrellelSum(arr, thresh, (start+end)/2, end, subResult2)
	}(arr, thresh, start, end)
	val := <-subResult + <-subResult2
	result <- val
}