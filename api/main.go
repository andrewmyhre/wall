package main

import (
  "encoding/json"
  "log"
  "net/http"
  "github.com/gorilla/mux"
)

var bricks []Brick

func main() {
  bricks = append(bricks, Brick{ID: "1"})

  router := mux.NewRouter()
  router.HandleFunc("/hello", GetHello).Methods("GET")
  log.Fatal(http.ListenAndServe(":8000", router))
}

func GetHello(w http.ResponseWriter, r *http.Request) {
  json.NewEncoder(w).Encode(bricks)
}

type Brick struct {
  ID string `json:"id,omitempty"`
}
