package handlers

import (
	"context"
	"errors"
	"fmt"
	"gopkg.in/mail.v2"
	"gorm.io/gorm"
	"math/rand"
	"otp_service/models"
	"otp_service/proto/otp_service"
	"strconv"
	"time"
)

type OTPService struct {
	otp_service.UnimplementedOTPServiceServer
	DB *gorm.DB
}

func (s *OTPService) SendForgetPasswordOTP(context context.Context, in *otp_service.SendForgetPasswordOTPRequest) (*otp_service.SendForgetPasswordOTPResponse, error) {

	m := mail.NewMessage()

	m.SetHeader("From", "noreply@example.com")
	m.SetHeader("To", in.Email)
	m.SetHeader("Subject", "Password Reset OTP")

	otp := generateOTP()

	htmlBody := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<style>
			body {
				font-family: Arial, sans-serif;
				background-color: #f9f9f9;
				color: #333;
				padding: 20px;
			}
			.container {
				max-width: 600px;
				margin: 0 auto;
				background-color: #ffffff;
				border-radius: 8px;
				box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
				padding: 20px;
			}
			.header {
				text-align: center;
				padding-bottom: 10px;
				border-bottom: 1px solid #eeeeee;
			}
			.header h1 {
				color: #4CAF50;
			}
			.content {
				margin: 20px 0;
				text-align: center;
			}
			.otp {
				font-size: 24px;
				font-weight: bold;
				color: #4CAF50;
			}
			.footer {
				text-align: center;
				font-size: 12px;
				color: #aaaaaa;
				margin-top: 20px;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h1>Password Reset Request</h1>
			</div>
			<div class="content">
				<p>Hello,</p>
				<p>We received a request to reset your password. Use the OTP below to reset your password:</p>
				<p class="otp">%s</p>
				<p>If you didn't request this, please ignore this email.</p>
			</div>
			<div class="footer">
				<p>&copy; 2025 SyncIO. All rights reserved.</p>
			</div>
		</div>
	</body>
	</html>
	`, otp)

	m.SetBody("text/html", htmlBody)

	d := mail.NewDialer("smtp.gmail.com", 587, "syncio78@gmail.com", "lzvf ibca fdsq uwsy")

	if err := d.DialAndSend(m); err != nil {
		return &otp_service.SendForgetPasswordOTPResponse{
			Success: false,
		}, err
	}

	expiredAt := time.Now().Add(5 * time.Minute)
	duration := time.Until(expiredAt)

	newData := &models.OTPRetakePassword{
		OTP:       otp,
		AccountID: uint(in.AccountID),
		ExpiredAt: expiredAt,
	}

	if err := s.DB.Model(models.OTPRetakePassword{}).Create(newData).Error; err != nil {
		return nil, err
	}

	if time.Now().After(expiredAt) {
		return nil, errors.New("Cannot reset OTP after " + duration.String())
	}

	go func() {
		time.Sleep(duration)
		if err := s.DB.Model(&models.OTPRetakePassword{}).Where("id = ?", newData.ID).Update("status", "expired").Error; err != nil {
		}
	}()

	return &otp_service.SendForgetPasswordOTPResponse{
		Success: true,
	}, nil
}

func (s *OTPService) CheckValidOTP(context context.Context, in *otp_service.CheckValidOTPRequest) (*otp_service.CheckValidOTPResponse, error) {
	var invalidAttempts int64
	timeWindow := time.Now().Add(-24 * time.Hour)

	// Count invalid attempts within the last 24 hours
	if err := s.DB.Model(&models.OTPInputs{}).
		Where("account_id = ? AND status = ? AND created_at > ?", in.AccountID, "rejected", timeWindow).
		Count(&invalidAttempts).Error; err != nil {
		return &otp_service.CheckValidOTPResponse{
			IsValid:  false,
			Attempts: 0,
		}, nil
	}

	if invalidAttempts >= 5 {
		fmt.Printf("attempts: %d\n", invalidAttempts)
		return &otp_service.CheckValidOTPResponse{
			IsValid:  false,
			Attempts: uint32(invalidAttempts),
		}, nil
	}

	// Create OTP input record for processing
	otpInput := models.OTPInputs{
		AccountID:  uint(in.AccountID),
		OTP:        strconv.FormatUint(in.OTP, 10),
		Status:     "pending",
		OTPSection: "reset password",
	}

	tx := s.DB.Begin()

	if err := tx.Create(&otpInput).Error; err != nil {
		tx.Rollback()
		return &otp_service.CheckValidOTPResponse{
			Attempts: uint32(invalidAttempts + 1),
			IsValid:  false,
		}, nil
	}

	// Find the OTP record in the database
	var otpRecord models.OTPRetakePassword
	if err := tx.Where("account_id = ? AND otp = ? AND status = ?", in.AccountID, in.OTP, "valid").First(&otpRecord).Error; err != nil {
		// If OTP is invalid, reject and return response
		tx.Model(&otpInput).Update("status", "rejected")
		tx.Commit()
		return &otp_service.CheckValidOTPResponse{
			Attempts: uint32(invalidAttempts + 1),
			IsValid:  false,
		}, nil
	}

	// If the OTP has expired, reject it and return response
	if time.Now().After(otpRecord.ExpiredAt) {
		tx.Model(&otpInput).Update("status", "rejected")
		tx.Model(&otpRecord).Update("status", "expired")
		tx.Commit()
		return &otp_service.CheckValidOTPResponse{
			Attempts: uint32(invalidAttempts + 1),
			IsValid:  false,
		}, nil
	}

	tx.Model(&otpInput).Update("status", "approved")
	tx.Model(&otpRecord).Update("status", "invalid")
	tx.Commit()

	return &otp_service.CheckValidOTPResponse{
		Attempts: uint32(invalidAttempts), // Ensure the correct value is returned
		IsValid:  true,
	}, nil
}

func generateOTP() string {
	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(1000000)
	return fmt.Sprintf("%06d", otp)
}
