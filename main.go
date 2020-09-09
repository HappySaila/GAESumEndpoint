package main

import (
	"bytes"
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
	Msg string `json:"msg"`
}

func main() {
	fmt.Println("NO SUCH THING AS ONE HTTP CALLING ANOTHER ON SAME ENDPOINT LOCALHOST, but in cloud it works :)")
	http.HandleFunc("/Sum", Sum)
	http.HandleFunc("/Sum2", Sum2)
	http.HandleFunc("/Hello", Hello)
	http.HandleFunc("/IsDev", IsDev)
	http.HandleFunc("/TestJson", TestJson)
	http.HandleFunc("/PostSum", PostSum)
	http.HandleFunc("/RecursivePost", RecursivePost)

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
	var body ChunkData
	err := json.NewDecoder(r.Body).Decode(&body)
	handleErr(&w, err)

	encodeStruct(&w, ChunkData{
		End: 7,
		Start: 7,
		Total: 7,
		Msg: body.Msg,
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

func PostSum(w http.ResponseWriter, r *http.Request) {
	dts, err := json.Marshal(ChunkData{Msg: "I am initialized"})
	handleErr(&w, err)

	res, err := http.Post(url + "/Sum2", "application/json", bytes.NewBuffer(dts))
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

func RecursivePost(w http.ResponseWriter, r *http.Request) {
	var bodyIn ChunkData
	err := json.NewDecoder(r.Body).Decode(&bodyIn)
	if err != nil {
		bodyIn = ChunkData{Start: 9, Msg: "Opening:"}
	}

	bodyIn.Start--

	if bodyIn.Start <=1 || bodyIn.Start >= 10 {
		bodyIn.Msg = bodyIn.Msg + "+:Closing"
		encodeStruct(&w, bodyIn)
		return
	} else {
		bodyIn.Msg += "+"
	}

	dts, err := json.Marshal(bodyIn)
	handleErr(&w, err)

	res, err := http.Post(url + "/RecursivePost", "application/json", bytes.NewBuffer(dts))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer res.Body.Close()

	//read data into json
	var bodyOut ChunkData
	err = json.NewDecoder(res.Body).Decode(&bodyOut)
	handleErr(&w, err)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(bodyOut)
	handleErr(&w, err)
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
