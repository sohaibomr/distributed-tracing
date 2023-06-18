package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Order struct {
	ID       string `json:"id,omitempty"`
	Quantity int    `json:"quantity,omitempty"`
	Name     string `json:"name,omitempty"`
}

func payment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Processing payment...")
	item := Order{}
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	trans := newTransaction(item.ID)
	err = trans.processTransaction(r.Context())
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)

}
