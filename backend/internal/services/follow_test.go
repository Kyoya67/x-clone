package services

import (
	"context"
	"errors"
	"testing"
)

type fakeFollowRepository struct {
	followCalled   bool
	unfollowCalled bool
	followerID     string
	followeeID     string
	err            error
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
