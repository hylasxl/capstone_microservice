package handlers

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
	"user_service/constants"
	"user_service/models"
	"user_service/proto/user_service"
)

type UserService struct {
	user_service.UnimplementedUserServiceServer
	DB               *gorm.DB
	CloudinaryClient *CloudinaryService
}

func (svc *UserService) Login(ctx context.Context, in *user_service.LoginRequest) (*user_service.LoginResponse, error) {
	var account models.Account

	if in.Username == "" {
		return &user_service.LoginResponse{Error: "Username cannot be empty"}, nil
	}
	if in.Password == "" {
		return &user_service.LoginResponse{Error: "Password cannot be empty"}, nil
	}

	if err := svc.DB.Where("username = ?", in.Username).First(&account).Error; err != nil {
		return &user_service.LoginResponse{Error: "The username is not correct"}, nil
	}

	if account.IsSelfDeleted {
		return &user_service.LoginResponse{Error: "The account is deleted"}, nil
	}
	if account.IsRestricted {
		return &user_service.LoginResponse{Error: "The account is restricted by admin"}, nil
	}
	if account.IsBanned {
		return &user_service.LoginResponse{Error: "The account is banned by admin"}, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(in.Password)); err != nil {
		return &user_service.LoginResponse{Error: "The password is not correct"}, nil
	}

	return &user_service.LoginResponse{
		UserId: strconv.Itoa(int(account.ID)),
		RoleId: strconv.Itoa(int(account.AccountRoleID)),
	}, nil
}

func (svc *UserService) Signup(ctx context.Context, in *user_service.SignupRequest) (*user_service.SignupResponse, error) {

	if err := validateSignupInput(in); err != nil {
		return &user_service.SignupResponse{Error: err.Error()}, nil
	}

	tx := svc.DB.Begin()

	if exists, _ := recordExists(svc.DB, "accounts", "username = ?", in.Username); exists {
		tx.Rollback()
		return &user_service.SignupResponse{Error: "Duplicated username"}, nil
	}

	if exists, _ := recordExists(svc.DB, "account_infos", "email = ?", in.Email); exists {
		tx.Rollback()
		return &user_service.SignupResponse{Error: "Duplicated email"}, nil
	}

	if in.PhoneNumber != "" {
		if exists, _ := recordExists(svc.DB, "account_infos", "phone_number = ?", in.PhoneNumber); exists {
			tx.Rollback()
			return &user_service.SignupResponse{Error: "Duplicated phone number"}, nil
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		return &user_service.SignupResponse{Error: "Failed to generate password"}, nil
	}

	newAccount := &models.Account{
		Username:               in.Username,
		Password:               string(hashedPassword),
		AccountRoleID:          1,
		AccountCreatedByMethod: "normal",
	}

	if err := tx.Create(newAccount).Error; err != nil {
		tx.Rollback()
		return &user_service.SignupResponse{Error: "Failed to create account"}, nil
	}

	avatarUrl := constants.AVATARDEFAULTURL
	if len(in.Avatar) > 0 {
		uploadedAvatarUrl, err := svc.CloudinaryClient.UploadAvatar(in.Avatar)
		if err != nil {
			tx.Rollback()
			return &user_service.SignupResponse{Error: "Failed to upload avatar"}, nil
		}
		avatarUrl = uploadedAvatarUrl
	}

	accountAvatar := &models.AccountAvatar{
		AccountID: newAccount.ID,
		AvatarURL: avatarUrl,
	}
	if err := tx.Create(accountAvatar).Error; err != nil {
		tx.Rollback()
		return &user_service.SignupResponse{Error: "Failed to create account avatar"}, nil
	}

	newAccountInfo := &models.AccountInfo{
		Email:       in.Email,
		PhoneNumber: in.PhoneNumber,
		FirstName:   in.FirstName,
		LastName:    in.LastName,
		Gender:      in.Gender,
		DateOfBirth: time.Unix(in.Birthday, 0),
		AccountID:   newAccount.ID,
		AvatarID:    accountAvatar.ID,
	}
	if err := tx.Create(newAccountInfo).Error; err != nil {
		tx.Rollback()
		return &user_service.SignupResponse{Error: "Failed to create account"}, nil
	}

	tx.Commit()
	return &user_service.SignupResponse{
		AccountId: strconv.Itoa(int(newAccount.ID)),
	}, nil
}

func (svc *UserService) LoginWithGoogle(ctx context.Context, in *user_service.LoginWithGoogleRequest) (*user_service.LoginWithGoogleResponse, error) {
	var accountInfo models.AccountInfo
	var account models.Account

	//checkToken

	// Check if email exists in the system
	if err := svc.DB.Where("email = ?", in.Email).First(&accountInfo).Error; err != nil {
		// If email does not exist, create a new account
		tx := svc.DB.Begin()

		username, err := generateUsernameFromDisplayName(in.DisplayName, tx)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		newAccount := &models.Account{
			Username:               username, // No username for Google login
			Password:               "",       // No password for Google login
			AccountRoleID:          1,
			AccountCreatedByMethod: "google",
		}
		if err := tx.Create(newAccount).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		newAvatar := &models.AccountAvatar{
			AccountID: newAccount.ID,
			AvatarURL: in.PhotoURL,
		}

		if err := tx.Create(newAvatar).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		firstName, lastName := extractFirstAndLastName(in.DisplayName)

		newAccountInfo := &models.AccountInfo{
			Email:       in.Email,
			AccountID:   newAccount.ID,
			AvatarID:    newAvatar.ID,
			DateOfBirth: time.Now(),
			FirstName:   firstName,
			LastName:    lastName,
			Gender:      "other",
		}
		if err := tx.Create(newAccountInfo).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		tx.Commit()
		return &user_service.LoginWithGoogleResponse{
			Success:   true,
			AccountID: uint64(newAccount.ID),
		}, nil
	}

	// Check account linked to email
	if err := svc.DB.Where("id = ?", accountInfo.AccountID).First(&account).Error; err != nil {
		return nil, err
	}

	// Ensure the account is created via Google
	if account.AccountCreatedByMethod != "google" {
		return nil, errors.New("account is not created via Google")
	}

	// Successful Google login
	return &user_service.LoginWithGoogleResponse{
		Success:   true,
		AccountID: uint64(account.ID),
	}, nil
}

func generateUsernameFromDisplayName(displayName string, db *gorm.DB) (string, error) {
	// Normalize the display name
	username := strings.ToLower(displayName)
	username = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(username, "_") // Replace non-alphanumeric chars with "_"
	username = strings.Trim(username, "_")                                      // Trim leading/trailing "_"

	// Ensure uniqueness
	originalUsername := username
	for i := 1; ; i++ {
		var count int64
		db.Model(&models.Account{}).Where("username = ?", username).Count(&count)
		if count == 0 {
			break // Username is unique
		}
		username = originalUsername + "_" + strconv.Itoa(i)
	}

	return username, nil
}

func extractFirstAndLastName(displayName string) (string, string) {
	// Split the display name into words
	nameParts := strings.Fields(displayName)

	// If there are no parts, return empty strings
	if len(nameParts) == 0 {
		return "", ""
	}

	// First name is the first part
	firstName := nameParts[0]

	// Last name is the last part if there are multiple parts, otherwise empty
	lastName := ""
	if len(nameParts) > 1 {
		lastName = nameParts[len(nameParts)-1]
	}

	return firstName, lastName
}

func validateSignupInput(in *user_service.SignupRequest) error {
	if in.Username == "" || in.Password == "" {
		return errors.New("username and password cannot be empty")
	}
	if in.FirstName == "" || in.LastName == "" {
		return errors.New("first name and last name cannot be empty")
	}
	if in.Email == "" {
		return errors.New("email cannot be empty")
	}
	if in.Gender == "" {
		return errors.New("gender cannot be empty")
	}
	return nil
}

func recordExists(db *gorm.DB, table, query string, args ...interface{}) (bool, error) {
	var count int64
	if err := db.Table(table).Where(query, args...).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (svc *UserService) CheckExistingUsername(ctx context.Context, in *user_service.CheckExistingUsernameRequest) (*user_service.CheckExistingUsernameResponse, error) {

	isRecordExisted, err := recordExists(svc.DB, "accounts", "username = ?", in.Username)
	if err != nil {
		return nil, err
	}

	return &user_service.CheckExistingUsernameResponse{
		IsExisting: isRecordExisted,
	}, nil
}

func (svc *UserService) CheckExistingEmail(ctx context.Context, in *user_service.CheckExistingEmailRequest) (*user_service.CheckExistingEmailResponse, error) {

	isRecordExisted, err := recordExists(svc.DB, "account_infos", "email = ?", in.Email)
	if err != nil {
		return nil, err
	}

	return &user_service.CheckExistingEmailResponse{
		IsExisting: isRecordExisted,
	}, nil

}

func (svc *UserService) CheckExistingPhone(ctx context.Context, in *user_service.CheckExistingPhoneRequest) (*user_service.CheckExistingPhoneResponse, error) {

	isRecordExisted, err := recordExists(svc.DB, "account_infos", "phone_number = ?", in.Phone)
	if err != nil {
		return nil, err
	}

	return &user_service.CheckExistingPhoneResponse{
		IsExisting: isRecordExisted,
	}, nil

}

func (svc *UserService) CheckValidUser(ctx context.Context, in *user_service.CheckValidUserRequest) (*user_service.CheckValidUserResponse, error) {
	var account models.Account
	if err := svc.DB.Where("id = ? AND is_banned = ? AND is_self_deleted = ?", in.UserId, false, false).First(&account).Error; err != nil {
		return &user_service.CheckValidUserResponse{
			IsValid: false,
		}, nil
	}
	return &user_service.CheckValidUserResponse{
		IsValid: true,
	}, nil
}

func (svc *UserService) GetListAccountDisplayInfo(ctx context.Context, in *user_service.GetListAccountDisplayInfoRequest) (*user_service.GetListAccountDisplayInfoResponse, error) {
	tx := svc.DB.Begin()

	response := make([]*user_service.SingleDisplayInfo, len(in.IDs))

	for i, record := range in.IDs {
		data := &user_service.SingleDisplayInfo{}
		var accountInfo models.AccountInfo
		if err := tx.Where("account_id = ?", record).First(&accountInfo).Error; err != nil {
			tx.Rollback()
			return &user_service.GetListAccountDisplayInfoResponse{
				Error: "Cannot fetch account info",
			}, nil
		}
		displayedName := ""
		if accountInfo.NameDisplayType == "first_name_first" {
			displayedName = accountInfo.FirstName + " " + accountInfo.LastName
		} else {
			displayedName = accountInfo.LastName + " " + accountInfo.FirstName
		}

		var accountAvatar models.AccountAvatar
		if err := tx.Where("account_id = ? AND is_in_used = ?", record, true).First(&accountAvatar).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				accountAvatar = models.AccountAvatar{
					AvatarURL: constants.AVATARDEFAULTURL,
					AccountID: uint(record),
					IsInUsed:  true,
					IsDeleted: false,
				}
			} else {
				tx.Rollback()
				return &user_service.GetListAccountDisplayInfoResponse{
					Error: "Cannot fetch account avatar",
				}, err
			}
		}

		data.AccountID = record
		data.DisplayName = displayedName
		data.AvatarURL = accountAvatar.AvatarURL
		response[i] = data
	}

	defer tx.Rollback()

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return &user_service.GetListAccountDisplayInfoResponse{}, err
	}

	return &user_service.GetListAccountDisplayInfoResponse{
		Infos: response,
	}, nil
}

func (svc *UserService) GetAccountInfo(ctx context.Context, in *user_service.GetAccountInfoRequest) (*user_service.GetAccountInfoResponse, error) {
	tx := svc.DB.Begin()

	var existingAccount models.Account
	var account *user_service.Account
	if err := tx.Model(models.Account{}).Where("id = ?", in.AccountID).First(&existingAccount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return &user_service.GetAccountInfoResponse{
				Error: "Account doesnt exist",
			}, errors.New("account doesnt exist")
		} else {
			tx.Rollback()
			return &user_service.GetAccountInfoResponse{
				Error: "Failed to get account",
			}, errors.New("failed to get account")
		}
	} else {
		account = &user_service.Account{
			Username:      existingAccount.Username,
			RoleID:        uint32(existingAccount.AccountRoleID),
			CreateMethod:  existingAccount.AccountCreatedByMethod,
			IsBanned:      existingAccount.IsBanned,
			IsRestricted:  existingAccount.IsRestricted,
			IsSelfDeleted: existingAccount.IsSelfDeleted,
		}
	}

	var accountInfo models.AccountInfo
	var info *user_service.AccountInfo
	if err := tx.Model(models.AccountInfo{}).Where("account_id = ?", existingAccount.ID).First(&accountInfo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return &user_service.GetAccountInfoResponse{
				Error: "Account info doesnt exist",
			}, errors.New("account info doesnt exist")
		} else {
			tx.Rollback()
			return &user_service.GetAccountInfoResponse{
				Error: "Failed to get account info",
			}, errors.New("failed to get account info")
		}
	} else {
		info = &user_service.AccountInfo{
			FirstName:       accountInfo.FirstName,
			LastName:        accountInfo.LastName,
			Email:           accountInfo.Email,
			DateOfBirth:     accountInfo.DateOfBirth.Unix(),
			Gender:          accountInfo.Gender,
			MaterialStatus:  accountInfo.MaritalStatus,
			PhoneNumber:     accountInfo.PhoneNumber,
			NameDisplayType: accountInfo.NameDisplayType,
			Bio:             accountInfo.Bio,
		}
	}

	var existingAccountAvatar models.AccountAvatar
	var avatar *user_service.Avatar
	if err := tx.Model(models.AccountAvatar{}).Where("account_id = ? AND is_in_used = ?", existingAccount.ID, true).First(&existingAccountAvatar).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			avatar = &user_service.Avatar{
				ID:        0,
				AvatarURL: constants.AVATARDEFAULTURL,
				IsDeleted: false,
				IsInUse:   true,
			}
		} else {
			tx.Rollback()
			return &user_service.GetAccountInfoResponse{
				Error: "Failed to get account avatar",
			}, errors.New("failed to get account avatar")
		}
	} else {
		avatar = &user_service.Avatar{
			ID:        uint32(existingAccountAvatar.ID),
			AvatarURL: existingAccountAvatar.AvatarURL,
			IsDeleted: false,
			IsInUse:   true,
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return &user_service.GetAccountInfoResponse{
			Error: "failed to commit transaction",
		}, errors.New("failed to commit transaction")
	}

	return &user_service.GetAccountInfoResponse{
		AccountID:     in.AccountID,
		Account:       account,
		AccountInfo:   info,
		AccountAvatar: avatar,
	}, nil
}
func (svc *UserService) GetProfileInfo(ctx context.Context, in *user_service.GetProfileInfoRequest) (*user_service.GetProfileInfoResponse, error) {
	tx := svc.DB.Begin()

	if in.IsBlocked {
		return nil, errors.New("user is blocked")
	}

	isSelf := in.TargetAccountID == in.RequestAccountID
	var accountInfo models.AccountInfo
	var accountAvatar models.AccountAvatar
	var account models.Account

	// Fetch AccountInfo
	if err := tx.Model(&models.AccountInfo{}).Where("account_id = ?", in.TargetAccountID).First(&accountInfo).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Fetch AccountAvatar
	if err := tx.Model(&models.AccountAvatar{}).Where("account_id = ? AND is_in_used = ?", in.TargetAccountID, true).First(&accountAvatar).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			accountAvatar = models.AccountAvatar{
				AvatarURL: constants.AVATARDEFAULTURL,
				IsDeleted: false,
				AccountID: uint(in.TargetAccountID),
				IsInUsed:  true,
			}
		} else {
			tx.Rollback()
			return nil, err
		}
	}

	// Fetch Account
	if err := tx.Model(&models.Account{}).Where("id = ?", in.TargetAccountID).First(&account).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Apply privacy filter
	filteredAccountInfo := accountInfo
	if !isSelf {
		filteredAccountInfo = ApplyPrivacyFilter(&accountInfo, in.Privacy, in.IsFriend)
	}

	response := &user_service.GetProfileInfoResponse{
		AccountID: in.TargetAccountID,
		Account: &user_service.Account{
			Username:      account.Username,
			RoleID:        uint32(account.AccountRoleID),
			CreateMethod:  account.AccountCreatedByMethod,
			IsBanned:      account.IsBanned,
			IsRestricted:  account.IsRestricted,
			IsSelfDeleted: account.IsSelfDeleted,
		},
		AccountAvatar: &user_service.Avatar{
			AvatarURL: accountAvatar.AvatarURL,
			IsDeleted: false,
			IsInUse:   accountAvatar.IsInUsed,
			ID:        uint32(accountAvatar.ID),
		},
		AccountInfo: &user_service.AccountInfo{
			FirstName:       filteredAccountInfo.FirstName,
			LastName:        filteredAccountInfo.LastName,
			Email:           filteredAccountInfo.Email,
			DateOfBirth:     filteredAccountInfo.DateOfBirth.Unix(),
			Gender:          filteredAccountInfo.Gender,
			MaterialStatus:  filteredAccountInfo.MaritalStatus,
			PhoneNumber:     filteredAccountInfo.PhoneNumber,
			NameDisplayType: filteredAccountInfo.NameDisplayType,
			Bio:             filteredAccountInfo.Bio,
		},
		Privacy:   in.Privacy,
		IsFriend:  in.IsFriend,
		IsBlocked: in.IsBlocked,
		IsFollow:  in.IsFollow,
		Timestamp: time.Now().UTC().Unix(),
	}

	tx.Commit()
	return response, nil
}

func ApplyPrivacyFilter(accountInfo *models.AccountInfo, privacyIndices *user_service.PrivacyIndices, isFriend bool) models.AccountInfo {
	filtered := *accountInfo

	if privacyIndices.DateOfBirth == "public" || (isFriend && privacyIndices.DateOfBirth == "friend_only") {
	} else {
		filtered.DateOfBirth = time.Time{}
	}

	if privacyIndices.Gender == "public" || (isFriend && privacyIndices.Gender == "friend_only") {
	} else {
		filtered.Gender = ""
	}

	if privacyIndices.MaterialStatus == "public" || (isFriend && privacyIndices.MaterialStatus == "friend_only") {
	} else {
		filtered.MaritalStatus = ""
	}

	if privacyIndices.Phone == "public" || (isFriend && privacyIndices.Phone == "friend_only") {

	} else {
		filtered.PhoneNumber = ""
	}

	if privacyIndices.Email == "public" || (isFriend && privacyIndices.Email == "friend_only") {

	} else {
		filtered.Email = ""
	}

	if privacyIndices.Bio == "public" || (isFriend && privacyIndices.Bio == "friend_only") {
		// Include Bio
	} else {
		filtered.Bio = ""
	}

	return filtered
}

func (svc *UserService) ChangeAccountInfo(ctx context.Context, in *user_service.ChangeAccountDataRequest) (*user_service.ChangeAccountDataResponse, error) {

	if len(strings.TrimSpace(in.DataFieldName)) == 0 {
		return nil, errors.New("invalid data field name")
	}

	tx := svc.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if in.DataFieldName == "date_of_birth" {
		bdInt, err := strconv.ParseInt(in.Data, 10, 64)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		if err := tx.Model(&models.AccountInfo{}).
			Where("account_id = ?", in.AccountID).
			Update(in.DataFieldName, time.Unix(bdInt, 0)).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else {
		if err := tx.Model(&models.AccountInfo{}).
			Where("account_id = ?", in.AccountID).
			Update(in.DataFieldName, in.Data).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &user_service.ChangeAccountDataResponse{
		Success: true,
	}, nil
}

func (svc *UserService) ChangeAvatar(ctx context.Context, in *user_service.ChangeAvatarRequest) (*user_service.ChangeAvatarResponse, error) {
	tx := svc.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	if err := tx.Model(models.AccountAvatar{}).Where("account_id = ?", in.AccountID).Update("is_in_used", false).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	avatarUrl := constants.AVATARDEFAULTURL
	uploadedStatus := "uploaded"
	if len(in.Avatar) > 0 {
		uploadedAvatarUrl, err := svc.CloudinaryClient.UploadAvatar(in.Avatar)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		avatarUrl = uploadedAvatarUrl
		uploadedStatus = "failed"
	}

	accountAvatar := &models.AccountAvatar{
		AccountID: uint(in.AccountID),
		AvatarURL: avatarUrl,
	}
	if err := tx.Create(accountAvatar).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	accountAvatarHistory := &models.AccountAvatarHistory{
		AccountID:    uint(in.AccountID),
		AvatarURL:    avatarUrl,
		UploadStatus: uploadedStatus,
	}

	if err := tx.Create(accountAvatarHistory).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return &user_service.ChangeAvatarResponse{
		Success: true,
	}, nil
}

func (svc *UserService) VerifyUsernameAndEmail(ctx context.Context, in *user_service.VerifyUsernameAndEmailRequest) (*user_service.VerifyUsernameAndEmailResponse, error) {

	var existingUsername models.Account
	if err := svc.DB.Where("username = ?", in.Username).First(&existingUsername).Error; err != nil {
		return &user_service.VerifyUsernameAndEmailResponse{
			Success: false,
		}, nil
	}

	var existingEmail models.AccountInfo
	if err := svc.DB.Where("email = ? AND account_id = ?", in.Email, existingUsername.ID).First(&existingEmail).Error; err != nil {
		return &user_service.VerifyUsernameAndEmailResponse{
			Success: false,
		}, nil
	}

	return &user_service.VerifyUsernameAndEmailResponse{
		Success: true,
		UserID:  int64(existingUsername.ID),
	}, nil
}

func (svc *UserService) ChangePassword(ctx context.Context, in *user_service.ChangePasswordRequest) (*user_service.ChangePasswordResponse, error) {

	if in.AccountID <= 0 {
		return &user_service.ChangePasswordResponse{
			Success: false,
		}, errors.New("invalid account id")
	}

	var existingUsername models.Account
	if err := svc.DB.Where("id = ?", in.AccountID).First(&existingUsername).Error; err != nil {
		return &user_service.ChangePasswordResponse{
			Success: false,
		}, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return &user_service.ChangePasswordResponse{
			Success: false,
		}, err
	}

	existingUsername.Password = string(hashedPassword)
	if err := svc.DB.Save(&existingUsername).Error; err != nil {
		return &user_service.ChangePasswordResponse{
			Success: false,
		}, err
	}

	return &user_service.ChangePasswordResponse{
		Success: true,
	}, nil

}

func (svc *UserService) CustomDeleteAccount(ctx context.Context, in *user_service.CustomDeleteAccountRequest) (*user_service.CustomDeleteAccountResponse, error) {
	if in.AccountID <= 0 {
		return nil, errors.New("invalid account id")
	}

	var existingAccount models.Account
	if err := svc.DB.Where("id = ?", in.AccountID).First(&existingAccount).Error; err != nil {
		return nil, errors.New("invalid account id")
	}

	switch in.Method {
	case "admin":
		{
			if err := svc.DB.Model(&models.Account{}).Where("id = ?", existingAccount.ID).Update("is_banned", true).Error; err != nil {
				return nil, errors.New("failed to ban account")
			}
			break
		}
	case "self":
		{
			if err := svc.DB.Model(&models.Account{}).Where("id = ?", existingAccount.ID).Update("is_self_deleted", true).Error; err != nil {
				return nil, errors.New("failed to ban account")
			}
			break
		}
	default:
		return nil, errors.New("invalid method")
	}

	return &user_service.CustomDeleteAccountResponse{Success: true}, nil
}

func (svc *UserService) SearchAccount(ctx context.Context, in *user_service.SearchAccountRequest) (*user_service.SearchAccountResponse, error) {
	if in.Page < 1 {
		in.Page = 1
	}
	if in.PageSize < 1 {
		in.PageSize = 10
	}

	// Temporary struct for GORM
	type AccountDisplayInfo struct {
		AccountID uint64 `json:"account_id" gorm:"column:account_id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		AvatarURL string `json:"avatar_url"`
	}

	query := svc.DB.Model(&models.Account{}).
		Select(`
			accounts.id AS account_id, 
			account_infos.first_name, 
			account_infos.last_name, 
			account_avatars.avatar_url
		`).
		Joins("LEFT JOIN account_infos ON accounts.id = account_infos.account_id").
		Joins("LEFT JOIN account_avatars ON accounts.id = account_avatars.account_id AND account_avatars.is_in_used = true").
		Where("accounts.is_banned = ? AND accounts.is_self_deleted = ?", false, false)

	in.BlockedList = append(in.BlockedList, in.RequestAccountID)

	if len(in.BlockList) > 0 {
		query = query.Where("accounts.id NOT IN (?)", gorm.Expr("?", in.BlockList))
	}
	if len(in.BlockedList) > 0 {
		query = query.Where("accounts.id NOT IN (?)", gorm.Expr("?", in.BlockedList))
	}

	if in.QueryString != "" {
		searchTerm := "%" + in.QueryString + "%"
		query = query.Where("accounts.username LIKE ? OR account_infos.first_name LIKE ? OR account_infos.last_name LIKE ?", searchTerm, searchTerm, searchTerm)
	}

	offset := (int(in.Page) - 1) * int(in.PageSize)
	query = query.Limit(int(in.PageSize)).Offset(offset)

	var rawAccounts []AccountDisplayInfo
	err := query.Scan(&rawAccounts).Error
	if err != nil {
		return nil, err
	}

	if len(rawAccounts) == 0 {
		log.Println("No accounts found")
	}

	// Convert to Protobuf struct
	var accounts []*user_service.SingleDisplayInfo
	for _, acc := range rawAccounts {
		accounts = append(accounts, &user_service.SingleDisplayInfo{
			AccountID:   acc.AccountID,
			DisplayName: acc.FirstName + " " + acc.LastName,
			AvatarURL:   acc.AvatarURL,
		})
	}

	return &user_service.SearchAccountResponse{
		Account:  accounts,
		Page:     in.Page,
		PageSize: in.PageSize,
	}, nil
}

func (svc *UserService) GetNewRegisterationData(ctx context.Context, in *user_service.GetNewRegisterationDataRequest) (*user_service.GetNewRegisterationDataResponse, error) {
	response := &user_service.GetNewRegisterationDataResponse{
		RequestAccountID: in.RequestAccountID,
		PeriodLabel:      in.PeriodLabel,
	}

	var data []*user_service.DataTerms
	var totalUsers int64

	currentYear, currentMonth, currentDay := time.Now().Date()

	if in.PeriodLabel == "year" {
		// Get the requested year
		requestedYear := int(in.PeriodData)

		for i := 1; i <= 12; i++ {
			monthLabel := fmt.Sprintf("%d-%02d", requestedYear, i) // YYYY-MM
			var count int64

			if requestedYear < currentYear || (requestedYear == currentYear && i <= int(currentMonth)) {
				svc.DB.Model(&models.Account{}).
					Where("YEAR(created_at) = ? AND MONTH(created_at) = ? AND account_role_id = ?", requestedYear, i, 1).
					Count(&count)
				totalUsers += count
			} else {
				count = 0
			}

			data = append(data, &user_service.DataTerms{
				Label: monthLabel,
				Count: uint64(count),
			})
		}

	} else if in.PeriodLabel == "month" {
		// Get the requested year & month
		requestedYear := int(in.PeriodData / 100)  // Extract year (e.g., 202402 → 2024)
		requestedMonth := int(in.PeriodData % 100) // Extract month (e.g., 202402 → 2)

		// Get the last day of the requested month
		firstDay := time.Date(requestedYear, time.Month(requestedMonth), 1, 0, 0, 0, 0, time.UTC)
		lastDay := firstDay.AddDate(0, 1, -1).Day() // Get last day of the month

		for i := 1; i <= lastDay; i++ {
			dayLabel := fmt.Sprintf("%d-%02d-%02d", requestedYear, requestedMonth, i) // YYYY-MM-DD
			var count int64

			// Only fetch data for past and current days
			if requestedYear < currentYear ||
				(requestedYear == currentYear && requestedMonth < int(currentMonth)) ||
				(requestedYear == currentYear && requestedMonth == int(currentMonth) && i <= currentDay) {
				svc.DB.Model(&models.Account{}).
					Where("DATE(created_at) = ? AND account_role_id = ?", dayLabel, 1).
					Count(&count)
				totalUsers += count
			} else {
				count = 0 // Future days return 0
			}

			data = append(data, &user_service.DataTerms{
				Label: dayLabel,
				Count: uint64(count),
			})
		}
	}

	response.TotalUsers = uint64(totalUsers)
	response.Data = data
	return response, nil
}

func (svc *UserService) CountUserType(ctx context.Context, in *user_service.CountTypeUserRequest) (*user_service.CountTypeUserResponse, error) {

	response := &user_service.CountTypeUserResponse{
		RequestAccountID: in.RequestAccountID,
	}

	var totalUsers int64
	if err := svc.DB.Unscoped().Model(&models.Account{}).Count(&totalUsers).Error; err != nil {
		return nil, err
	}

	// Count banned users (including soft-deleted)
	var bannedUsers int64
	if err := svc.DB.Unscoped().Model(&models.Account{}).Where("is_banned = ?", true).Count(&bannedUsers).Error; err != nil {
		return nil, err
	}

	// Count deleted users (soft-deleted or self-deleted)
	var deletedUsers int64
	if err := svc.DB.Unscoped().Model(&models.Account{}).
		Where("deleted_at IS NOT NULL OR is_self_deleted = ?", true).
		Count(&deletedUsers).Error; err != nil {
		return nil, err
	}

	// Assign values to the response
	response.TotalUsers = uint64(totalUsers)
	response.BannedUsers = uint64(bannedUsers)
	response.DeletedUsers = uint64(deletedUsers)

	return response, nil
}

func (svc *UserService) GetAccountList(ctx context.Context, in *user_service.GetAccountListRequest) (*user_service.GetAccountListResponse, error) {
	var accounts []models.Account

	result := svc.DB.Model(&models.Account{}).Order("created_at DESC").
		Limit(int(in.PageSize)).
		Offset(int((in.Page - 1) * in.PageSize)).
		Find(&accounts)

	if result.Error != nil {
		return nil, result.Error
	}

	accountList := make([]*user_service.AccountRawInfo, len(accounts))
	for i, acc := range accounts {
		accountList[i] = &user_service.AccountRawInfo{
			AccountID:     uint32(acc.ID),
			Username:      acc.Username,
			IsBanned:      acc.IsBanned,
			Method:        acc.AccountCreatedByMethod,
			IsSelfDeleted: acc.IsSelfDeleted,
		}
	}

	// Return paginated response
	return &user_service.GetAccountListResponse{
		Accounts: accountList,
		Page:     in.Page,
		PageSize: in.PageSize,
	}, nil
}

func (svc *UserService) SearchAccountList(ctx context.Context, in *user_service.SearchAccountListRequest) (*user_service.SearchAccountListResponse, error) {
	var accounts []models.Account
	query := svc.DB

	if in.QueryString != "" {
		query = query.Where("username LIKE ?", "%"+in.QueryString+"%")
	}

	// Pagination
	offset := (in.Page - 1) * in.PageSize
	query = query.Limit(int(in.PageSize)).Offset(int(offset))

	// Execute query
	if err := query.Order("created_at DESC").Find(&accounts).Error; err != nil {
		return nil, err
	}

	// Convert database models to response format
	var accountList []*user_service.AccountRawInfo
	for _, account := range accounts {
		accountList = append(accountList, &user_service.AccountRawInfo{
			AccountID: uint32(account.ID),
			Username:  account.Username,
			Method:    account.AccountCreatedByMethod,
			IsBanned:  account.IsBanned,
		})
	}

	// Build response
	response := &user_service.SearchAccountListResponse{
		Accounts: accountList,
		Page:     in.Page,
		PageSize: in.PageSize,
	}

	return response, nil
}

func (svc *UserService) ResolveBan(ctx context.Context, in *user_service.ResolveBanRequest) (*user_service.ResolveBanResponse, error) {
	// Find the account
	var account models.Account
	if err := svc.DB.First(&account, in.AccountID).Error; err != nil {
		return &user_service.ResolveBanResponse{Success: false}, err
	}

	switch in.Action {
	case "ban":
		account.IsBanned = true
	case "activate":
		account.IsBanned = false
	default:
		return &user_service.ResolveBanResponse{Success: false}, errors.New("invalid action")
	}

	// Update account status
	if err := svc.DB.Save(&account).Error; err != nil {
		return &user_service.ResolveBanResponse{Success: false}, err
	}

	// Return success response
	return &user_service.ResolveBanResponse{Success: true}, nil
}

func (svc *UserService) GetUsername(ctx context.Context, in *user_service.GetUsernameRequest) (*user_service.GetUsernameResponse, error) {
	var account models.Account
	if err := svc.DB.First(&account, in.AccountID).Error; err != nil {
		return nil, err
	}

	return &user_service.GetUsernameResponse{
		Username: account.Username,
	}, nil
}
