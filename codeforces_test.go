//go:build integration

package codeforcesClient

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestIntegrationUserInfo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := NewClient()
	resp, err := client.UserInfo(ctx, &UserInfoParams{Handles: "tourist"})
	if err != nil {
		t.Fatalf("UserInfo failed: %v", err)
	}
	if resp.Status != "OK" {
		t.Fatalf("status = %s, want OK (comment: %s)", resp.Status, resp.Comment)
	}
	if len(resp.Result) == 0 {
		t.Fatal("no users returned")
	}
	if resp.Result[0].Handle != "tourist" {
		t.Errorf("handle = %s, want tourist", resp.Result[0].Handle)
	}
}

func TestIntegrationUserRating(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := NewClient()
	resp, err := client.UserRating(ctx, &UserRatingParams{Handle: "tourist"})
	if err != nil {
		t.Fatalf("UserRating failed: %v", err)
	}
	if resp.Status != "OK" {
		t.Fatalf("status = %s, want OK", resp.Status)
	}
}

func TestIntegrationContestList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := NewClient()
	resp, err := client.ContestList(ctx, &ContestListParams{Gym: false})
	if err != nil {
		t.Fatalf("ContestList failed: %v", err)
	}
	if resp.Status != "OK" {
		t.Fatalf("status = %s, want OK", resp.Status)
	}
	if len(resp.Result) == 0 {
		t.Fatal("no contests returned")
	}
}

func TestIntegrationWithAuth(t *testing.T) {
	apiKey := os.Getenv("CF_API_KEY")
	secret := os.Getenv("CF_SECRET")
	if apiKey == "" || secret == "" {
		t.Skip("CF_API_KEY and CF_SECRET not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := NewClient(WithSigner(NewStaticSigner(apiKey, secret)))
	resp, err := client.UserFriends(ctx, &UserFriendsParams{})
	if err != nil {
		t.Fatalf("UserFriends failed: %v", err)
	}
	if resp.Status != "OK" {
		t.Fatalf("status = %s, want OK", resp.Status)
	}
}
