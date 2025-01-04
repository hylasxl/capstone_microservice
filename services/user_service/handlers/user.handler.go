package handlers

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
	if err := svc.DB.Where("id = ?", in.UserId).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &user_service.CheckValidUserResponse{
				IsValid: false,
			}, nil
		} else {
			return &user_service.CheckValidUserResponse{
				IsValid: false,
			}, nil
		}
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
