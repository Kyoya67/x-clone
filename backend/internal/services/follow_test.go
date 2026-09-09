package services

import (
	"context"
	"errors"
	"testing"
)

type fakeFollowRepository struct {
	followCalled   bool
	unfollowCalled bool
	listCalled     bool
	followerID     string
	followeeID     string
	followeeIDs    []string
	err            error
}

func (f *fakeFollowRepository) ListFolloweeIDs(_ context.Context, followerID string) ([]string, error) {
	f.listCalled = true
	f.followerID = followerID
	return f.followeeIDs, f.err
}

func (f *fakeFollowRepository) Follow(_ context.Context, followerID, followeeID string) error {
	f.followCalled = true
	f.followerID = followerID
	f.followeeID = followeeID
	return f.err
}

func (f *fakeFollowRepository) Unfollow(_ context.Context, followerID, followeeID string) error {
	f.unfollowCalled = true
	f.followerID = followerID
	f.followeeID = followeeID
	return f.err
}

func TestFollowServiceFollow(t *testing.T) {
	repository := &fakeFollowRepository{}
	service := NewFollowService(repository)

	if err := service.Follow(context.Background(), "follower-1", "followee-1"); err != nil {
		t.Fatal(err)
	}
	if !repository.followCalled || repository.followerID != "follower-1" || repository.followeeID != "followee-1" {
		t.Fatalf("unexpected repository call: %+v", repository)
	}
}

func TestFollowServiceFollowRejectsSelfFollow(t *testing.T) {
	repository := &fakeFollowRepository{}
	service := NewFollowService(repository)

	err := service.Follow(context.Background(), "user-1", "user-1")
	if !errors.Is(err, ErrCannotFollowSelf) {
		t.Fatalf("expected self-follow error, got %v", err)
	}
	if repository.followCalled {
		t.Fatal("repository should not be called")
	}
}

func TestFollowServiceUnfollow(t *testing.T) {
	repository := &fakeFollowRepository{}
	service := NewFollowService(repository)

	if err := service.Unfollow(context.Background(), "follower-1", "followee-1"); err != nil {
		t.Fatal(err)
	}
	if !repository.unfollowCalled || repository.followerID != "follower-1" || repository.followeeID != "followee-1" {
		t.Fatalf("unexpected repository call: %+v", repository)
	}
}

func TestFollowServiceUnfollowRejectsSelfFollow(t *testing.T) {
	repository := &fakeFollowRepository{}
	service := NewFollowService(repository)

	err := service.Unfollow(context.Background(), "user-1", "user-1")
	if !errors.Is(err, ErrCannotFollowSelf) {
		t.Fatalf("expected self-follow error, got %v", err)
	}
	if repository.unfollowCalled {
		t.Fatal("repository should not be called")
	}
}

func TestFollowServiceListFolloweeIDs(t *testing.T) {
	repository := &fakeFollowRepository{followeeIDs: []string{"followee-1", "followee-2"}}
	service := NewFollowService(repository)

	followeeIDs, err := service.ListFolloweeIDs(context.Background(), "follower-1")
	if err != nil {
		t.Fatal(err)
	}
	if !repository.listCalled || repository.followerID != "follower-1" {
		t.Fatalf("unexpected repository call: %+v", repository)
	}
	if len(followeeIDs) != 2 || followeeIDs[0] != "followee-1" || followeeIDs[1] != "followee-2" {
		t.Fatalf("unexpected followee IDs: %v", followeeIDs)
	}
}
