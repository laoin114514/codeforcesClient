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
	resp, err := client.WithContext(ctx).UserInfo(&UserInfoParams{Handles: "tourist"})
	if err != nil {
		t.Fatalf("UserInfo failed: %v", err)
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
	resp, err := client.WithContext(ctx).UserRating(&UserRatingParams{Handle: "tourist"})
	if err != nil {
		t.Fatalf("UserRating failed: %v", err)
	}
}

func TestIntegrationContestList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := NewClient()
	resp, err := client.WithContext(ctx).ContestList(&ContestListParams{Gym: false})
	if err != nil {
		t.Fatalf("ContestList failed: %v", err)
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
	resp, err := client.WithContext(ctx).UserFriends(&UserFriendsParams{})
	if err != nil {
		t.Fatalf("UserFriends failed: %v", err)
	}
}
