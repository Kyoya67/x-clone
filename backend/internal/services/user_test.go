package services

import (
	"context"
	"errors"
	"testing"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

type fakeUserRepository struct {
	updateCalled bool
	findCalled   bool
	createCalled bool
	id           string
	request      models.UpdateUserProfileRequest
	user         models.User
	err          error
}

func (f *fakeUserRepository) FindOrCreateByOIDC(context.Context, string, string, string) (models.User, error) {
	f.createCalled = true
	return f.user, f.err
}

func (f *fakeUserRepository) FindByID(context.Context, string) (models.User, error) {
	f.findCalled = true
	return f.user, f.err
}

func (f *fakeUserRepository) UpdateProfile(_ context.Context, id string, request models.UpdateUserProfileRequest) (models.User, error) {
	f.updateCalled = true
	f.id = id
	f.request = request
	return f.user, f.err
}

func TestUserServiceUpdateProfile(t *testing.T) {
	repository := &fakeUserRepository{user: models.User{ID: "user-1", Handle: "kyoya_dev"}}
	service := NewUserService(repository)

	user, err := service.UpdateProfile(context.Background(), "user-1", models.UpdateUserProfileRequest{
		Handle:      " kyoya_dev ",
		DisplayName: " 京谷 ",
		Bio:         " hello ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if user.Handle != "kyoya_dev" {
		t.Fatalf("unexpected user: %+v", user)
	}
	if !repository.updateCalled || repository.id != "user-1" {
		t.Fatalf("unexpected repository call: %+v", repository)
	}
	if repository.request.Handle != "kyoya_dev" || repository.request.DisplayName != "京谷" || repository.request.Bio != "hello" {
		t.Fatalf("request should be trimmed: %+v", repository.request)
	}
}

func TestUserServiceFindOrCreateByOIDC(t *testing.T) {
	repository := &fakeUserRepository{user: models.User{ID: "user-1"}}
	service := NewUserService(repository)

	user, err := service.FindOrCreateByOIDC(context.Background(), "subject", "user@example.com", "User")
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != "user-1" || !repository.createCalled {
		t.Fatalf("unexpected result: user=%+v repository=%+v", user, repository)
	}
}

func TestUserServiceFindByID(t *testing.T) {
	repository := &fakeUserRepository{user: models.User{ID: "user-1"}}
	service := NewUserService(repository)

	user, err := service.FindByID(context.Background(), "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != "user-1" || !repository.findCalled {
		t.Fatalf("unexpected result: user=%+v repository=%+v", user, repository)
	}
}

func TestUserServiceUpdateProfileRejectsInvalidHandle(t *testing.T) {
	tests := []string{"", "日本語", "has-hyphen", "toolong_handle_14"}
	for _, handle := range tests {
		t.Run(handle, func(t *testing.T) {
			repository := &fakeUserRepository{}
			service := NewUserService(repository)

			_, err := service.UpdateProfile(context.Background(), "user-1", models.UpdateUserProfileRequest{
				Handle:      handle,
				DisplayName: "京谷",
			})

			var appErr *apperrors.Error
			if !errors.As(err, &appErr) || appErr.ErrCode != string(apperrors.BadParam) {
				t.Fatalf("expected bad param error, got %v", err)
			}
			if repository.updateCalled {
				t.Fatal("repository should not be called")
			}
		})
	}
}

func TestUserServiceUpdateProfileRejectsInvalidDisplayName(t *testing.T) {
	service := NewUserService(&fakeUserRepository{})

	_, err := service.UpdateProfile(context.Background(), "user-1", models.UpdateUserProfileRequest{
		Handle:      "kyoya_dev",
		DisplayName: " ",
	})

	var appErr *apperrors.Error
	if !errors.As(err, &appErr) || appErr.ErrCode != string(apperrors.BadParam) {
		t.Fatalf("expected bad param error, got %v", err)
	}
}

func TestUserServiceUpdateProfileRejectsTooLongBio(t *testing.T) {
	service := NewUserService(&fakeUserRepository{})
	bio := ""
	for range 161 {
		bio += "あ"
	}

	_, err := service.UpdateProfile(context.Background(), "user-1", models.UpdateUserProfileRequest{
		Handle:      "kyoya_dev",
		DisplayName: "京谷",
		Bio:         bio,
	})

	var appErr *apperrors.Error
	if !errors.As(err, &appErr) || appErr.ErrCode != string(apperrors.BadParam) {
		t.Fatalf("expected bad param error, got %v", err)
	}
}
