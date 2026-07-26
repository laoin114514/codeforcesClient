package codeforcesClient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoRequestSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.String(), "user.info") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(apiResponse{
			Status: "OK",
			Result: json.RawMessage(`[{"handle":"tourist","rating":3800}]`),
		})
	}))
	defer srv.Close()

	client := NewClient(WithBaseURL(srv.URL+"/"), WithRateLimit(0))
	var resp UserInfoResponse
	err := client.doRequest(context.Background(), "user.info",
		&UserInfoParams{Handles: "tourist"}, nil, &resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Result) != 1 || resp.Result[0].Handle != "tourist" || resp.Result[0].Rating != 3800 {
		t.Errorf("unexpected result: %+v", resp)
	}
}

func TestDoRequestAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(apiResponse{
			Status:  "FAILED",
			Comment: "handles: Field should not be empty",
		})
	}))
	defer srv.Close()

	client := NewClient(WithBaseURL(srv.URL+"/"), WithRateLimit(0))
	var resp UserInfoResponse
	err := client.doRequest(context.Background(), "user.info",
		&UserInfoParams{}, nil, &resp)

	var cfErr *CFError
	if !errors.As(err, &cfErr) {
		t.Fatalf("expected *CFError, got %T: %v", err, err)
	}
	if cfErr.Code != ErrAPI {
		t.Errorf("code = %d, want ErrAPI(%d)", cfErr.Code, ErrAPI)
	}
	if cfErr.Message != "handles: Field should not be empty" {
		t.Errorf("message = %q", cfErr.Message)
	}
}

func TestDoRequestHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	client := NewClient(WithBaseURL(srv.URL+"/"), WithRateLimit(0))
	var resp UserInfoResponse
	err := client.doRequest(context.Background(), "user.info",
		&UserInfoParams{Handles: "tourist"}, nil, &resp)

	var cfErr *CFError
	if !errors.As(err, &cfErr) {
		t.Fatalf("expected *CFError, got %v", err)
	}
	if cfErr.Code != ErrNetwork {
		t.Errorf("code = %d, want ErrNetwork(%d)", cfErr.Code, ErrNetwork)
	}
}

func TestDoRequestRateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := NewClient(WithBaseURL(srv.URL+"/"), WithRateLimit(0))
	var resp UserInfoResponse
	err := client.doRequest(context.Background(), "user.info",
		&UserInfoParams{Handles: "tourist"}, nil, &resp)

	var cfErr *CFError
	if !errors.As(err, &cfErr) {
		t.Fatalf("expected *CFError, got %v", err)
	}
	if cfErr.Code != ErrRateLimit {
		t.Errorf("code = %d, want ErrRateLimit(%d)", cfErr.Code, ErrRateLimit)
	}
}

func TestDoRequestInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	client := NewClient(WithBaseURL(srv.URL+"/"), WithRateLimit(0))
	var resp UserInfoResponse
	err := client.doRequest(context.Background(), "user.info",
		&UserInfoParams{Handles: "tourist"}, nil, &resp)

	var cfErr *CFError
	if !errors.As(err, &cfErr) {
		t.Fatalf("expected *CFError, got %v", err)
	}
	if cfErr.Code != ErrNetwork {
		t.Errorf("code = %d, want ErrNetwork(%d)", cfErr.Code, ErrNetwork)
	}
}

func TestRawRequest(t *testing.T) {
	expected := `{"status":"OK","result":[{"handle":"tourist","rating":3800}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(expected))
	}))
	defer srv.Close()

	client := NewClient(WithBaseURL(srv.URL+"/"), WithRateLimit(0))
	body, err := client.RawRequest(context.Background(), "user.info", &UserInfoParams{Handles: "tourist"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != expected {
		t.Errorf("body = %s, want %s", string(body), expected)
	}
}

func TestRawRequestAuthError(t *testing.T) {
	client := NewClient(WithSigner(NewStaticSigner("", "")))
	_, err := client.RawRequest(context.Background(), "user.friends", &UserFriendsParams{})

	var cfErr *CFError
	if !errors.As(err, &cfErr) {
		t.Fatalf("expected *CFError, got %v", err)
	}
	if cfErr.Code != ErrAuth {
		t.Errorf("code = %d, want ErrAuth(%d)", cfErr.Code, ErrAuth)
	}
}

func TestDoRequestContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := NewClient()
	var resp UserInfoResponse
	err := client.doRequest(ctx, "user.info",
		&UserInfoParams{Handles: "tourist"}, nil, &resp)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParamsEncoding(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		qs := r.URL.Query()
		if qs.Get("handle") != "tourist" {
			t.Errorf("handle = %s, want tourist", qs.Get("handle"))
		}
		if qs.Get("from") != "1" {
			t.Errorf("from = %s, want 1", qs.Get("from"))
		}
		json.NewEncoder(w).Encode(apiResponse{
			Status: "OK",
			Result: json.RawMessage(`[]`),
		})
	}))
	defer srv.Close()

	client := NewClient(WithBaseURL(srv.URL+"/"), WithRateLimit(0))
	var resp UserStatusResponse
	err := client.doRequest(context.Background(), "user.status",
		&UserStatusParams{Handle: "tourist", From: 1, Count: 10}, nil, &resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
