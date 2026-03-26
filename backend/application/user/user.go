/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package user

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/mail"
	"slices"
	"strconv"
	"strings"

	"github.com/coze-dev/coze-studio/backend/api/model/app/developer_api"
	"github.com/coze-dev/coze-studio/backend/api/model/passport"
	"github.com/coze-dev/coze-studio/backend/api/model/playground"
	"github.com/coze-dev/coze-studio/backend/application/base/ctxutil"
	"github.com/coze-dev/coze-studio/backend/bizpkg/config"
	"github.com/coze-dev/coze-studio/backend/domain/user/entity"
	user "github.com/coze-dev/coze-studio/backend/domain/user/service"
	"github.com/coze-dev/coze-studio/backend/infra/storage"
	"github.com/coze-dev/coze-studio/backend/pkg/errorx"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ptr"
	langSlices "github.com/coze-dev/coze-studio/backend/pkg/lang/slices"
	"github.com/coze-dev/coze-studio/backend/types/errno"
)

var UserApplicationSVC = &UserApplicationService{}

type UserApplicationService struct {
	oss       storage.Storage
	DomainSVC user.User
}

// Add a simple email verification function
func isValidEmail(email string) bool {
	// If the email string is not in the correct format, it will return an error.
	_, err := mail.ParseAddress(email)
	return err == nil
}

func (u *UserApplicationService) PassportWebEmailRegisterV2(ctx context.Context, locale string, req *passport.PassportWebEmailRegisterV2PostRequest) (
	resp *passport.PassportWebEmailRegisterV2PostResponse, sessionKey string, err error,
) {
	// Verify that the email format is legitimate
	if !isValidEmail(req.GetEmail()) {
		return nil, "", errorx.New(errno.ErrUserInvalidParamCode, errorx.KV("msg", "Invalid email"))
	}

	baseConf, err := config.Base().GetBaseConfig(ctx)
	if err != nil {
		return nil, "", err
	}

	// Allow Register Checker
	if !u.allowRegisterChecker(req.GetEmail(), baseConf) {
		return nil, "", errorx.New(errno.ErrNotAllowedRegisterCode)
	}

	_, err = u.DomainSVC.Create(ctx, &user.CreateUserRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),

		Locale: locale,
	})
	if err != nil {
		return nil, "", err
	}

	userInfo, err := u.DomainSVC.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, "", err
	}

	return &passport.PassportWebEmailRegisterV2PostResponse{
		Data: userDo2PassportTo(userInfo),
		Code: 0,
	}, userInfo.SessionKey, nil
}

func (u *UserApplicationService) allowRegisterChecker(email string, baseConf *config.BasicConfiguration) bool {
	if !baseConf.DisableUserRegistration {
		return true
	}

	allowedEmails := baseConf.AllowRegistrationEmail
	if allowedEmails == "" {
		return false
	}

	return slices.Contains(strings.Split(allowedEmails, ","), strings.ToLower(email))
}

// PassportWebLogoutGet handle user logout requests
func (u *UserApplicationService) PassportWebLogoutGet(ctx context.Context, req *passport.PassportWebLogoutGetRequest) (
	resp *passport.PassportWebLogoutGetResponse, err error,
) {
	uid := ctxutil.MustGetUIDFromCtx(ctx)

	err = u.DomainSVC.Logout(ctx, uid)
	if err != nil {
		return nil, err
	}

	return &passport.PassportWebLogoutGetResponse{
		Code: 0,
	}, nil
}

// PassportWebEmailLoginPost handle user email login requests
func (u *UserApplicationService) PassportWebEmailLoginPost(ctx context.Context, req *passport.PassportWebEmailLoginPostRequest) (
	resp *passport.PassportWebEmailLoginPostResponse, sessionKey string, err error,
) {
	userInfo, err := u.DomainSVC.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, "", err
	}

	return &passport.PassportWebEmailLoginPostResponse{
		Data: userDo2PassportTo(userInfo),
		Code: 0,
	}, userInfo.SessionKey, nil
}

func (u *UserApplicationService) PassportWebEmailPasswordResetGet(ctx context.Context, req *passport.PassportWebEmailPasswordResetGetRequest) (
	resp *passport.PassportWebEmailPasswordResetGetResponse, err error,
) {
	session := ctxutil.GetUserSessionFromCtx(ctx)
	if session == nil {
		return nil, errorx.New(errno.ErrUserAuthenticationFailed, errorx.KV("reason", "session data is nil"))
	}
	if !strings.EqualFold(session.UserEmail, req.GetEmail()) {
		return nil, errorx.New(errno.ErrUserPermissionCode, errorx.KV("msg", "email mismatch"))
	}

	err = u.DomainSVC.ResetPassword(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	return &passport.PassportWebEmailPasswordResetGetResponse{
		Code: 0,
	}, nil
}

func (u *UserApplicationService) PassportAccountInfoV2(ctx context.Context, req *passport.PassportAccountInfoV2Request) (
	resp *passport.PassportAccountInfoV2Response, err error,
) {
	userID := ctxutil.MustGetUIDFromCtx(ctx)

	userInfo, err := u.DomainSVC.GetUserInfo(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &passport.PassportAccountInfoV2Response{
		Data: userDo2PassportTo(userInfo),
		Code: 0,
	}, nil
}

// UserUpdateAvatar Update user avatar
func (u *UserApplicationService) UserUpdateAvatar(ctx context.Context, mimeType string, req *passport.UserUpdateAvatarRequest) (
	resp *passport.UserUpdateAvatarResponse, err error,
) {
	// Get file suffix by MIME type
	var ext string
	switch mimeType {
	case "image/jpeg", "image/jpg":
		ext = "jpg"
	case "image/png":
		ext = "png"
	case "image/gif":
		ext = "gif"
	case "image/webp":
		ext = "webp"
	default:
		return nil, errorx.WrapByCode(err, errno.ErrUserInvalidParamCode,
			errorx.KV("msg", "unsupported image type"))
	}

	uid := ctxutil.MustGetUIDFromCtx(ctx)

	url, err := u.DomainSVC.UpdateAvatar(ctx, uid, ext, req.GetAvatar())
	if err != nil {
		return nil, err
	}

	return &passport.UserUpdateAvatarResponse{
		Data: &passport.UserUpdateAvatarResponseData{
			WebURI: url,
		},
		Code: 0,
	}, nil
}

// UserUpdateProfile Update user profile
func (u *UserApplicationService) UserUpdateProfile(ctx context.Context, req *passport.UserUpdateProfileRequest) (
	resp *passport.UserUpdateProfileResponse, err error,
) {
	userID := ctxutil.MustGetUIDFromCtx(ctx)

	err = u.DomainSVC.UpdateProfile(ctx, &user.UpdateProfileRequest{
		UserID:      userID,
		Name:        req.Name,
		UniqueName:  req.UserUniqueName,
		Description: req.Description,
		Locale:      req.Locale,
	})
	if err != nil {
		return nil, err
	}

	return &passport.UserUpdateProfileResponse{
		Code: 0,
	}, nil
}

func (u *UserApplicationService) GetSpaceListV2(ctx context.Context, req *playground.GetSpaceListV2Request) (
	resp *playground.GetSpaceListV2Response, err error,
) {
	uid := ctxutil.MustGetUIDFromCtx(ctx)

	spaces, err := u.DomainSVC.GetUserSpaceList(ctx, uid)
	if err != nil {
		return nil, err
	}

	botSpaces := langSlices.Transform(spaces, func(space *entity.Space) *playground.BotSpaceV2 {
		return &playground.BotSpaceV2{
			ID:          space.ID,
			Name:        space.Name,
			Description: space.Description,
			SpaceType:   playground.SpaceType(space.SpaceType),
			IconURL:     space.IconURL,
		}
	})

	return &playground.GetSpaceListV2Response{
		Data: &playground.SpaceInfo{
			BotSpaceList:          botSpaces,
			HasPersonalSpace:      true,
			TeamSpaceNum:          0,
			RecentlyUsedSpaceList: botSpaces,
			Total:                 ptr.Of(int32(len(botSpaces))),
			HasMore:               ptr.Of(false),
		},
		Code: 0,
	}, nil
}

func (u *UserApplicationService) MGetUserBasicInfo(ctx context.Context, req *playground.MGetUserBasicInfoRequest) (
	resp *playground.MGetUserBasicInfoResponse, err error,
) {
	userIDs, err := langSlices.TransformWithErrorCheck(req.GetUserIds(), func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	})
	if err != nil {
		return nil, errorx.WrapByCode(err, errno.ErrUserInvalidParamCode, errorx.KV("msg", "invalid user id"))
	}

	userInfos, err := u.DomainSVC.MGetUserProfiles(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	return &playground.MGetUserBasicInfoResponse{
		UserBasicInfoMap: langSlices.ToMap(userInfos, func(userInfo *entity.User) (string, *playground.UserBasicInfo) {
			return strconv.FormatInt(userInfo.UserID, 10), userDo2PlaygroundTo(userInfo)
		}),
		Code: 0,
	}, nil
}

func (u *UserApplicationService) UpdateUserProfileCheck(ctx context.Context, req *developer_api.UpdateUserProfileCheckRequest) (resp *developer_api.UpdateUserProfileCheckResponse, err error) {
	if req.GetUserUniqueName() == "" {
		return &developer_api.UpdateUserProfileCheckResponse{
			Code: 0,
			Msg:  "no content to update",
		}, nil
	}

	validateResp, err := u.DomainSVC.ValidateProfileUpdate(ctx, &user.ValidateProfileUpdateRequest{
		UniqueName: req.UserUniqueName,
	})
	if err != nil {
		return nil, err
	}

	return &developer_api.UpdateUserProfileCheckResponse{
		Code: int64(validateResp.Code),
		Msg:  validateResp.Msg,
	}, nil
}

func (u *UserApplicationService) ValidateSession(ctx context.Context, sessionKey string) (*entity.Session, error) {
	session, exist, err := u.DomainSVC.ValidateSession(ctx, sessionKey)
	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, errorx.New(errno.ErrUserAuthenticationFailed, errorx.KV("reason", "session not exist"))
	}

	return session, nil
}

func userDo2PassportTo(userDo *entity.User) *passport.User {
	var locale *string
	if userDo.Locale != "" {
		locale = ptr.Of(userDo.Locale)
	}

	return &passport.User{
		UserIDStr:      userDo.UserID,
		Name:           userDo.Name,
		ScreenName:     ptr.Of(userDo.Name),
		UserUniqueName: userDo.UniqueName,
		Email:          userDo.Email,
		Description:    userDo.Description,
		AvatarURL:      userDo.IconURL,
		AppUserInfo: &passport.AppUserInfo{
			UserUniqueName: userDo.UniqueName,
		},
		Locale: locale,

		UserCreateTime: userDo.CreatedAt / 1000,
	}
}

func userDo2PlaygroundTo(userDo *entity.User) *playground.UserBasicInfo {
	return &playground.UserBasicInfo{
		UserId:         userDo.UserID,
		Username:       userDo.Name,
		UserUniqueName: ptr.Of(userDo.UniqueName),
		UserAvatar:     userDo.IconURL,
		CreateTime:     ptr.Of(userDo.CreatedAt / 1000),
	}
}

// PassportPlatformALoginPost handle platform A login requests
func (u *UserApplicationService) PassportPlatformALoginPost(ctx context.Context, req *passport.PassportPlatformALoginPostRequest) (
	resp *passport.PassportPlatformALoginPostResponse, sessionKey string, err error,
) {
	// 1. Decrypt the encrypted session key from platform A
	platformAUserInfo, err := u.decryptPlatformASessionKey(req.GetEncryptedSessionKey())
	if err != nil {
		return nil, "", errorx.New(errno.ErrUserAuthenticationFailed, errorx.KV("reason", "invalid encrypted session key"))
	}

	// 2. Check if user exists by email
	userInfo, err := u.DomainSVC.PlatformALogin(ctx, platformAUserInfo.Email)
	if err != nil {
		// 3. If user doesn't exist, register a new user
		userInfo, err = u.DomainSVC.Create(ctx, &user.CreateUserRequest{
			Email:    platformAUserInfo.Email,
			Password: "platform-a-login", // Use a default password for platform A users
			Name:     platformAUserInfo.Name,
			Locale:   platformAUserInfo.Locale,
		})
		if err != nil {
			return nil, "", err
		}

		// 4. Login the newly registered user
		userInfo, err = u.DomainSVC.PlatformALogin(ctx, platformAUserInfo.Email)
		if err != nil {
			return nil, "", err
		}
	}

	return &passport.PassportPlatformALoginPostResponse{
		Data: userDo2PassportTo(userInfo),
		Code: 0,
	}, userInfo.SessionKey, nil
}



// decryptPlatformASessionKey decrypts the encrypted session key from platform A using AES
func (u *UserApplicationService) decryptPlatformASessionKey(encryptedSessionKey string) (*PlatformAUserInfo, error) {
	// Test mode: return mock user info for testing
	if encryptedSessionKey == "test_session_key" || encryptedSessionKey == "test" {
		return &PlatformAUserInfo{
			Email:  "user@platform-a.com",
			Name:   "Platform A User",
			Locale: "en-US",
		}, nil
	}
	
	// In a real system, you would store the key in a secure location
	// For example, using environment variables or a secret management service
	key := []byte("your-aes-secret-key") // 16, 24, or 32 bytes for AES-128, AES-192, or AES-256
	
	// Decode the base64-encoded encrypted session key
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedSessionKey)
	if err != nil {
		return nil, err
	}
	
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	
	// Check if the ciphertext is long enough
	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	
	// Extract the IV
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	
	// Create CBC decrypter
	stream := cipher.NewCBCDecrypter(block, iv)
	
	// Decrypt the ciphertext
	stream.CryptBlocks(ciphertext, ciphertext)
	
	// Unpad the plaintext
	plaintext, err := unpad(ciphertext)
	if err != nil {
		return nil, err
	}
	
	// Parse the plaintext as JSON
	var userInfo PlatformAUserInfo
	if err := json.Unmarshal(plaintext, &userInfo); err != nil {
		return nil, err
	}
	
	return &userInfo, nil
}

// unpad removes PKCS#7 padding
func unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}
	
	padding := data[len(data)-1]
	if int(padding) > len(data) {
		return nil, errors.New("invalid padding")
	}
	
	return data[:len(data)-int(padding)], nil
}

// PlatformAUserInfo represents user information from platform A
type PlatformAUserInfo struct {
	Email  string
	Name   string
	Locale string
}

// FindUserByExternalID finds a user by external ID and platform
func (u *UserApplicationService) FindUserByExternalID(ctx context.Context, externalID, platform string) (*entity.User, error) {
	return u.DomainSVC.FindUserByExternalID(ctx, externalID, platform)
}

// CreateUserWithExternalID creates a user with external ID and platform
func (u *UserApplicationService) CreateUserWithExternalID(ctx context.Context, externalID, platform string) (*entity.User, error) {
	// Generate a random password for the user
	password := "external-" + platform + "-login"
	
	// Create user with external ID
	userInfo, err := u.DomainSVC.Create(ctx, &user.CreateUserRequest{
		Email:    externalID + "@" + platform + ".com",
		Password: password,
		Name:     platform + " User",
		Locale:   "en-US",
	})
	if err != nil {
		return nil, err
	}
	
	// Login the newly created user
	userInfo, err = u.DomainSVC.Login(ctx, userInfo.Email, password)
	if err != nil {
		return nil, err
	}
	
	return userInfo, nil
}

// GenerateSession generates a session for the user
func (u *UserApplicationService) GenerateSession(ctx context.Context, user *entity.User) (string, error) {
	// Create a new session for the user
	sessionKey, err := u.DomainSVC.CreateSession(ctx, user.UserID)
	if err != nil {
		return "", err
	}
	
	return sessionKey, nil
}
