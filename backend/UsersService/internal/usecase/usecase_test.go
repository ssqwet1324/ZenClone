package usecase

import (
	"UsersService/internal/config"
	"UsersService/internal/entity"
	"UsersService/internal/mocks"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func newTestUseCase(t *testing.T) (*Usecase, *mocks.MockRepositoryProvider, *gomock.Controller) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockRepositoryProvider(ctrl)

	uc := &Usecase{
		repo: mockRepo,
		cfg:  &config.Config{BucketName: "test-bucket"},
		log:  zap.NewNop(),
	}

	return uc, mockRepo, ctrl
}

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	return string(hash)
}

// ====== AddUser ======

func TestAddUser(t *testing.T) {
	req := entity.AddUserRequest{Username: "testuser", Login: "test@test.com", Password: "pass"}

	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockRepositoryProvider)
		wantErr   error
	}{
		{
			name: "success",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().CheckUser(gomock.Any(), req.Username).Return(false, nil)
				repo.EXPECT().AddUser(gomock.Any(), req).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "user already exists",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().CheckUser(gomock.Any(), req.Username).Return(true, nil)
			},
			wantErr: entity.ErrUserAlreadyExists,
		},
		{
			name: "check user error",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().CheckUser(gomock.Any(), req.Username).Return(false, errors.New("db error"))
			},
			wantErr: entity.ErrInternalServer,
		},
		{
			name: "add user error",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().CheckUser(gomock.Any(), req.Username).Return(false, nil)
				repo.EXPECT().AddUser(gomock.Any(), req).Return(errors.New("db error"))
			},
			wantErr: entity.ErrFailedToAddUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, mockRepo, ctrl := newTestUseCase(t)
			defer ctrl.Finish()

			tt.mockSetup(mockRepo)

			err := uc.AddUser(context.Background(), req)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ====== CompareAuthData ======

func TestCompareAuthData(t *testing.T) {
	userID := uuid.New()
	password := "Qwerty12"
	hashed := hashPassword(t, password)

	tests := []struct {
		name      string
		req       entity.AuthRequest
		mockSetup func(repo *mocks.MockRepositoryProvider)
		wantErr   error
	}{
		{
			name: "success",
			req:  entity.AuthRequest{Login: "test@test.com", Password: password},
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().
					GetUserInfoByLogin(gomock.Any(), "test@test.com").
					Return(&entity.LoginResponse{ID: userID, Password: hashed, Username: "testuser"}, nil)
			},
			wantErr: nil,
		},
		{
			name: "user not found",
			req:  entity.AuthRequest{Login: "notexist@test.com", Password: password},
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().
					GetUserInfoByLogin(gomock.Any(), "notexist@test.com").
					Return(nil, errors.New("not found"))
			},
			wantErr: entity.ErrUserNotFound,
		},
		{
			name: "incorrect password",
			req:  entity.AuthRequest{Login: "test@test.com", Password: "wrongpass"},
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().
					GetUserInfoByLogin(gomock.Any(), "test@test.com").
					Return(&entity.LoginResponse{ID: userID, Password: hashed, Username: "testuser"}, nil)
			},
			wantErr: entity.ErrIncorrectPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, mockRepo, ctrl := newTestUseCase(t)
			defer ctrl.Finish()

			tt.mockSetup(mockRepo)

			resp, err := uc.CompareAuthData(context.Background(), tt.req)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, userID, resp.ID)
			}
		})
	}
}

// ====== GetRefreshToken ======

func TestGetRefreshToken(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockRepositoryProvider)
		wantErr   error
	}{
		{
			name: "success",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().
					GetRefreshTokenByUserID(gomock.Any(), userID).
					Return(&entity.RefreshTokenResponse{RefreshToken: "some-token"}, nil)
			},
			wantErr: nil,
		},
		{
			name: "repo error",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().
					GetRefreshTokenByUserID(gomock.Any(), userID).
					Return(nil, errors.New("db error"))
			},
			wantErr: entity.ErrFailedToGetRefreshToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, mockRepo, ctrl := newTestUseCase(t)
			defer ctrl.Finish()

			tt.mockSetup(mockRepo)

			resp, err := uc.GetRefreshToken(context.Background(), userID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.RefreshToken)
			}
		})
	}
}

// ====== UpdateRefreshToken ======

func TestUpdateRefreshToken(t *testing.T) {
	req := entity.UpdateRefreshTokenRequest{ID: uuid.New(), RefreshToken: "new-token"}

	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockRepositoryProvider)
		wantErr   error
	}{
		{
			name: "success",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().UpdateRefreshToken(gomock.Any(), req.ID, req.RefreshToken).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "repo error",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().UpdateRefreshToken(gomock.Any(), req.ID, req.RefreshToken).Return(errors.New("db error"))
			},
			wantErr: entity.ErrFailedToUpdateRefreshToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, mockRepo, ctrl := newTestUseCase(t)
			defer ctrl.Finish()

			tt.mockSetup(mockRepo)

			err := uc.UpdateRefreshToken(context.Background(), req)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ====== GetUserProfileByID ======

func TestGetUserProfileByID(t *testing.T) {
	yourID := uuid.New()
	otherID := uuid.New()

	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockRepositoryProvider)
		wantErr   error
	}{
		{
			name: "success",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().
					GetUserProfileByID(gomock.Any(), otherID).
					Return(&entity.ProfileUserInfoResponse{FirstName: "John", LastName: "Doe"}, nil)
				repo.EXPECT().
					GetAvatarURL(gomock.Any(), "test-bucket", otherID).
					Return("http://avatar.url", nil)
				repo.EXPECT().
					CheckSubscription(gomock.Any(), yourID.String(), otherID.String()).
					Return(false, nil)
			},
			wantErr: nil,
		},
		{
			name: "user not found",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().
					GetUserProfileByID(gomock.Any(), otherID).
					Return(nil, errors.New("not found"))
			},
			wantErr: entity.ErrFailedToGetUserInfo,
		},
		{
			name: "avatar url error",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().
					GetUserProfileByID(gomock.Any(), otherID).
					Return(&entity.ProfileUserInfoResponse{FirstName: "John"}, nil)
				repo.EXPECT().
					GetAvatarURL(gomock.Any(), "test-bucket", otherID).
					Return("", errors.New("storage error"))
			},
			wantErr: entity.ErrFailedToGetAvatarURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, mockRepo, ctrl := newTestUseCase(t)
			defer ctrl.Finish()

			tt.mockSetup(mockRepo)

			resp, err := uc.GetUserProfileByID(context.Background(), yourID, otherID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

// ====== SubscribeToUser ======

func TestSubscribeToUser(t *testing.T) {
	followerID := uuid.New()
	followingID := uuid.New()

	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockRepositoryProvider)
		wantErr   error
	}{
		{
			name: "success",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().SubscribeFromUser(gomock.Any(), followerID, followingID).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "repo error",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().SubscribeFromUser(gomock.Any(), followerID, followingID).Return(errors.New("db error"))
			},
			wantErr: entity.ErrFailedToSubscribe,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, mockRepo, ctrl := newTestUseCase(t)
			defer ctrl.Finish()

			tt.mockSetup(mockRepo)

			err := uc.SubscribeToUser(context.Background(), followerID, followingID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ====== UnsubscribeFromUser ======

func TestUnsubscribeFromUser(t *testing.T) {
	followerID := uuid.New()
	followingID := uuid.New()

	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockRepositoryProvider)
		wantErr   error
	}{
		{
			name: "success",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().UnsubscribeFromUser(gomock.Any(), followerID, followingID).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "repo error",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().UnsubscribeFromUser(gomock.Any(), followerID, followingID).Return(errors.New("db error"))
			},
			wantErr: entity.ErrFailedToUnsubscribe,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, mockRepo, ctrl := newTestUseCase(t)
			defer ctrl.Finish()

			tt.mockSetup(mockRepo)

			err := uc.UnsubscribeFromUser(context.Background(), followerID, followingID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ====== GetSubsUser ======

func TestGetSubsUser(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockRepositoryProvider)
		wantErr   error
	}{
		{
			name: "success",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().
					GetSubsUser(gomock.Any(), userID, "test-bucket").
					Return(&entity.SubsList{Subs: []entity.SubUserInfo{{Username: "user1"}}}, nil)
			},
			wantErr: nil,
		},
		{
			name: "repo error",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().
					GetSubsUser(gomock.Any(), userID, "test-bucket").
					Return(nil, errors.New("db error"))
			},
			wantErr: entity.ErrFailedToGetUserInfo,
		},
		{
			name: "nil result - no subscriptions",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().
					GetSubsUser(gomock.Any(), userID, "test-bucket").
					Return(nil, nil)
			},
			wantErr: entity.ErrNoSubscriptions,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, mockRepo, ctrl := newTestUseCase(t)
			defer ctrl.Finish()

			tt.mockSetup(mockRepo)

			resp, err := uc.GetSubsUser(context.Background(), userID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

// ====== GlobalSearchPeople ======

func TestGlobalSearchPeople(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockRepositoryProvider)
		wantErr   error
	}{
		{
			name: "success",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().
					GlobalSearchPeople(gomock.Any(), "John", "Doe", "test-bucket").
					Return(&entity.PersonDateList{Persons: []entity.PersonDate{{Name: "John", LastName: "Doe"}}}, nil)
			},
			wantErr: nil,
		},
		{
			name: "no results",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().
					GlobalSearchPeople(gomock.Any(), "John", "Doe", "test-bucket").
					Return(&entity.PersonDateList{Persons: []entity.PersonDate{}}, nil)
			},
			wantErr: entity.ErrUserNotFound,
		},
		{
			name: "repo error",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().
					GlobalSearchPeople(gomock.Any(), "John", "Doe", "test-bucket").
					Return(nil, errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, mockRepo, ctrl := newTestUseCase(t)
			defer ctrl.Finish()

			tt.mockSetup(mockRepo)

			resp, err := uc.GlobalSearchPeople(context.Background(), "John", "Doe")

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}
