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
var PORT = "8085"
var url = "https://gae-by-endpoint.uc.r.appspot.com"

type TestData struct {
	UserId  int `json:"userId"`
	Id  int `json:"id"`
	Title  string `json:"title"`
	Completed  bool `json:"completed"`
}

type ChunkData struct {
	Start int64 `json:"start"`
	End int64 `json:"end"`
	Total int64 `json:"total"`
}

func main() {
	if os.Getenv("IsDev") == "" {
		fmt.Println("Running on Dev")
		url = "https://localhost:" + PORT
	}
	http.HandleFunc("/Sum", Sum)
	http.HandleFunc("/Sum2", Sum2)
	http.HandleFunc("/Hello", Hello)
	http.HandleFunc("/IsDev", IsDev)
	http.HandleFunc("/TestJson", TestJson)

	port := os.Getenv("PORT")
	if port == "" {
		port = PORT
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Nice")
}

func Hello(w http.ResponseWriter, r *http.Request) {
	encodeStruct(&w, ChunkData{Start: 777})
}

func encodeStruct(w *http.ResponseWriter, obj interface{}) {
	(*w).Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(*w).Encode(obj)
}

func IsDev(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, os.Getenv("IsDev") != "")
}

// Sum will add the total of n integers
func Sum(w http.ResponseWriter, r *http.Request) {
	res, err := http.Get(url + "/Sum2")
	fmt.Println(url)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer res.Body.Close()

	//read data into json
	var body ChunkData
	err = json.NewDecoder(res.Body).Decode(&body)
	handleErr(&w, err)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(body)
	handleErr(&w, err)
}

func Sum2(w http.ResponseWriter, r *http.Request) {
	encodeStruct(&w, ChunkData{
		End: 7,
		Start: 7,
		Total: 7,
	})
}

func TestJson(w http.ResponseWriter, r *http.Request) {
	res, err := http.Get("https://jsonplaceholder.typicode.com/todos/1")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer res.Body.Close()

	//decode HTTP body into golang struct
	var body TestData
	err = json.NewDecoder(res.Body).Decode(&body)
	handleErr(&w, err)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(body)
	handleErr(&w, err)
	fmt.Fprint(w, "Done")
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

func handleErr(w *http.ResponseWriter, err error) {
	if err != nil {
		fmt.Fprint(*w, "Err")
		fmt.Fprint(*w, err.Error())
	}
}
