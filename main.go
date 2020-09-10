package main

import (
	"bytes"
	"cloud.google.com/go/logging"
	"context"
	"encoding/json"
	"fmt"
	ltype "google.golang.org/genproto/googleapis/logging/type"
	"log"
	"net/http"
	"os"
	"strconv"
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
	http.HandleFunc("/TestGlog", TestGlog)
	http.HandleFunc("/ParallelSumStart", ParallelSumStart)
	http.HandleFunc("/ParallelSum", ParallelSum)

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

type ParallelData struct {
	Bucket string `json:"bucket"`
	Start int `json:"start"`
	End int `json:"end"`
	Total int64 `json:"total"`
	Thresh int `json:"thresh"`
}

func ParallelSumStart(w http.ResponseWriter, r *http.Request) {
	gLog("Program Start!", nil)
	// Initialize the Parallel Sum
	data := ParallelData{
		Bucket: "Bucket1000",
		Start: 0,
		End: 1000,
		Total: 0,
		Thresh: 70,
	}

	dts, err := json.Marshal(data)
	handleErr(&w, err)
	res, err := http.Post(url + "/ParallelSum", "application/json", bytes.NewBuffer(dts))
	var bodyOut ParallelData
	err = json.NewDecoder(res.Body).Decode(&bodyOut)
	handleErr(&w, err)
	encodeStruct(&w, bodyOut)
	gLog("Program End!", nil)
}

func SequencialSum(data ParallelData) ParallelData {
	// Read file from datastore

	// Read data from file into transaction[] and process
	gLog(fmt.Sprintf("Processing in: %s, out: %s", data.Start, data.End), nil)
	data.Total = 1
	return data
}

func ParallelSum(w http.ResponseWriter, r *http.Request) {
	gLog("Start thread PS\n", nil)
	// Unpack data sent from parent
	var bodyIn ParallelData
	err := json.NewDecoder(r.Body).Decode(&bodyIn)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	gLog("Check Thresh\n", nil)
	gLog(fmt.Sprintf("Current in: %s, out: %s", bodyIn.Start, bodyIn.End), nil)
	// Get data from storage bucket
	var bodyOut ParallelData
	if bodyIn.End - bodyIn.Start <= bodyIn.Thresh {
		bodyOut = SequencialSum(bodyIn)
		encodeStruct(&w, bodyOut)
		return
	}

	subResult := make(chan int64)
	subResult2 := make(chan int64)
	boundary := (bodyIn.Start + bodyIn.End)/2
	gLog(fmt.Sprintf("Boundary: %s Thresh\n", boundary), nil)

	go func(data ParallelData) {
		data.End = boundary
		gLog(fmt.Sprintf("1start: %s, end: %s", data.Start, data.End), nil)
		dts, err := json.Marshal(data)
		handleErr(&w, err)
		res, err := http.Post(url + "/ParallelSum", "application/json", bytes.NewBuffer(dts))
		handleErr(&w, err)

		// Bubble up child call
		var childBodyIn ParallelData
		err = json.NewDecoder(res.Body).Decode(&childBodyIn)
		handleErr(&w, err)

		subResult <- childBodyIn.Total
	}(bodyIn)

	go func(data ParallelData) {
		data.Start = boundary
		gLog(fmt.Sprintf("2start: %s, end: %s", data.Start, data.End), nil)
		dts, err := json.Marshal(data)
		handleErr(&w, err)
		res, err := http.Post(url + "/ParallelSum", "application/json", bytes.NewBuffer(dts))
		handleErr(&w, err)

		// Bubble up child call
		var childBodyIn ParallelData
		err = json.NewDecoder(res.Body).Decode(&childBodyIn)
		handleErr(&w, err)

		subResult2 <- childBodyIn.Total
	}(bodyIn)

	gLog("Adding Values\n", nil)
	val := <-subResult + <-subResult2
	bodyOut.Total = val
	bodyOut.Bucket += "+"
	encodeStruct(&w, bodyOut)
}

func handleErr(w *http.ResponseWriter, err error) {
	if err != nil {
		fmt.Fprint(*w, "Err")
		fmt.Fprint(*w, err.Error())
	}
}

func gLog(text string, severity *ltype.LogSeverity) {
	ctx := context.Background()

	// Creates a client.
	client, err := logging.NewClient(ctx, "gae-by-endpoint")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Sets log name to unix nano second
	logger := client.Logger(strconv.Itoa(int(time.Now().UnixNano())))

	// Set severity based on params. Default Severity: DEBUG
	var logSeverity logging.Severity
	if severity == nil {
		logSeverity = logging.Severity(ltype.LogSeverity_DEBUG)
	} else {
		logSeverity = logging.Severity(*severity)
	}

	// Adds an entry to the log buffer.
	logger.Log(logging.Entry{
		Payload: text,
		Severity: logSeverity,
	})

	// Closes the client and flushes the buffer to the Stackdriver Logging
	// service.
	if err := client.Close(); err != nil {
		log.Fatalf("Failed to close client: %v", err)
	}
}

func TestGlog(w http.ResponseWriter, r *http.Request) {
	gLog("HELLO GLOG TEST", nil)
}
