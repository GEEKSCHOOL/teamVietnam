package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func init() {
	currentSignal = &Signal{ActionType: 1, Time: time.Now().Unix() + 10000}
}

func TestGetSignal(t *testing.T) {
	req, _ := http.NewRequest("GET", "/signal", nil)

	w := httptest.NewRecorder()

	MiddleWare(signalHandler)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code: %v but got: %v", http.StatusOK, w.Code)
		return
	}

	var res Response
	json.Unmarshal(w.Body.Bytes(), &res)

	if res.Code != http.StatusOK {
		t.Errorf("expected Code %v but got %v", http.StatusOK, res.Code)
		return
	}
}

func TestPostSignal(t *testing.T) {
	// reset currentSignal
	currentSignal = &Signal{}

	data, _ := json.Marshal(Signal{ActionType: 1, Time: 10})
	req, _ := http.NewRequest("POST", "/signal", bytes.NewReader(data))

	w := httptest.NewRecorder()
	MiddleWare(signalHandler)(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status code: %v but got %v", http.StatusCreated, w.Code)
		return
	}

	var res Response

	json.Unmarshal(w.Body.Bytes(), &res)

	if res.Code != http.StatusCreated {
		t.Errorf("expected code %v but got %v", http.StatusCreated, res.Code)
		return
	}
}
