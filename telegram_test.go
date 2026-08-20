package main

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestTelegramClientSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{\"ok\":true}"))
	}))
	defer server.Close()

	client := NewTelegramClient("dummy_token", "12345")
	client.apiBaseURL = server.URL
	client.httpClient = server.Client()

	err := client.SendMessage("Hello world")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestTelegramClientRetrySuccess(t *testing.T) {
	var count int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if atomic.AddInt32(&count, 1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{\"ok\":true}"))
	}))
	defer server.Close()

	client := NewTelegramClient("dummy_token", "12345")
	client.apiBaseURL = server.URL
	client.httpClient = server.Client()

	err := client.SendMessage("Hello retry")
	if err != nil {
		t.Fatalf("expected success on second attempt, got %v", err)
	}
}

func TestTelegramClientErrors(t *testing.T) {
	// 1. Status not OK
	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer s1.Close()

	client := NewTelegramClient("dummy_token", "12345")
	client.httpClient = s1.Client()

	err := client.doSend(s1.URL, sendMessageRequest{ChatID: "1", Text: "hi"})
	if err == nil {
		t.Error("expected error for non-200 status")
	}

	// 2. Invalid JSON response
	s2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{not-json}"))
	}))
	defer s2.Close()
	client.httpClient = s2.Client()

	err = client.doSend(s2.URL, sendMessageRequest{ChatID: "1", Text: "hi"})
	if err == nil {
		t.Error("expected error for invalid json response")
	}

	// 3. result.OK == false
	s3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{\"ok\":false,\"error_code\":400,\"description\":\"chat not found\"}"))
	}))
	defer s3.Close()
	client.httpClient = s3.Client()

	err = client.doSend(s3.URL, sendMessageRequest{ChatID: "1", Text: "hi"})
	if err == nil {
		t.Error("expected error when ok is false")
	}

	// 4. Invalid URL
	err = client.doSend("::invalid-url", sendMessageRequest{ChatID: "1", Text: "hi"})
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestSendMessageFailures(t *testing.T) {
	client := NewTelegramClient("dummy", "123")
	client.apiBaseURL = "http://127.0.0.1:0" // unreachable
	err := client.SendMessage("test")
	if err == nil {
		t.Error("expected error sending to invalid endpoint")
	}
}
