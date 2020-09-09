package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
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
	//go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Nice")
	//}()
}

func Hello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(ChunkData{
		End: 9,
		Start: 10,
		Total: 9,
	})
}

func IsDev(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, os.Getenv("IsDev") != "")
}

// Sum will add the total of n integers
func Sum(w http.ResponseWriter, r *http.Request) {
	res, err := http.Get(url + "/Sum2")
	if err != nil {
		fmt.Println("errr")
		fmt.Println(err)
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
	fmt.Fprint(w, "Donne")
}

func Sum2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	data := ChunkData{
		End: 7,
		Start: 7,
		Total: 7,
	}
	json.NewEncoder(w).Encode(data)
}

func TestJson(w http.ResponseWriter, r *http.Request) {
	res, err := http.Get("https://jsonplaceholder.typicode.com/todos/1")
	if err != nil {
		fmt.Println("err")
		fmt.Println(err.Error())
		return
	}
	defer res.Body.Close()

	//read data into json
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

//helpers to decode
type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if r.Header.Get("Content-Type") != "" {
		value := r.Header.Get("Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			return &malformedRequest{status: http.StatusUnsupportedMediaType, msg: msg}
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			return &malformedRequest{status: http.StatusRequestEntityTooLarge, msg: msg}

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		msg := "Request body must only contain a single JSON object"
		return &malformedRequest{status: http.StatusBadRequest, msg: msg}
	}

	return nil
}