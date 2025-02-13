package handlers

import (
	"context"
	"errors"
	"fmt"
	"friend_service/models"
	"friend_service/proto/friend_service"
	"gorm.io/gorm"
	"sort"
	"strconv"
)

type FriendService struct {
	friend_service.UnimplementedFriendServiceServer
	DB *gorm.DB
}

func NewFriendService(db *gorm.DB) *FriendService {
	return &FriendService{
		DB: db,
	}
}

func (svc *FriendService) SendFriend(ctx context.Context, in *friend_service.SendFriendRequest) (*friend_service.SendFriendResponse, error) {
	tx := svc.DB.Begin()

	var existingFriend models.FriendList
	if err := tx.Where(
		"((first_account_id = ? AND second_account_id = ?  AND is_valid = true) OR (first_account_id = ? AND second_account_id = ?  AND is_valid = true))",
		in.FromAccountID, in.ToAccountID, in.ToAccountID, in.FromAccountID,
	).First(&existingFriend).Error; err == nil {
		tx.Rollback()
		return &friend_service.SendFriendResponse{
			Error: "Accounts have already friends",
		}, nil
	}

	var blockedFriend models.FriendBlock
	if err := tx.Where(
		"(first_account_id = ? AND second_account_id = ?  AND is_blocked = true) OR (first_account_id = ? AND second_account_id = ?  AND is_blocked = true)",
		in.FromAccountID, in.ToAccountID, in.ToAccountID, in.FromAccountID,
	).First(&blockedFriend).Error; err == nil {
		tx.Rollback()
		return &friend_service.SendFriendResponse{
			Error: "Accounts have already blocked friends",
		}, nil
	}

	var existingRequest models.FriendListRequest
	if err := tx.Where(
		"((sender_account_id = ? AND receiver_account_id = ? AND request_status = 'pending' AND is_recalled = false) OR (sender_account_id = ? AND receiver_account_id = ? AND request_status = 'pending' AND is_recalled = false))",
		in.FromAccountID, in.ToAccountID, in.ToAccountID, in.FromAccountID,
	).First(&existingRequest).Error; err == nil {
		tx.Rollback()
		return &friend_service.SendFriendResponse{
			Error: "Request existed",
		}, nil
	}

	senderID, err := strconv.ParseUint(in.FromAccountID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid FromAccountID: %v", err)
	}

	receiverID, err := strconv.ParseUint(in.ToAccountID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid ToAccountID: %v", err)
	}

	newRequest := models.FriendListRequest{
		SenderAccountID:   uint(senderID),
		ReceiverAccountID: uint(receiverID),
	}

	if err := tx.Create(&newRequest).Error; err != nil {
		tx.Rollback()
		return &friend_service.SendFriendResponse{
			Error: "Failed to create friend request",
		}, nil
	}

	tx.Commit()

	return &friend_service.SendFriendResponse{
		Success:   true,
		RequestID: uint64(newRequest.ID),
	}, nil
}

func (svc *FriendService) ResolveFriendRequest(ctx context.Context, in *friend_service.FriendListResolveRequest) (*friend_service.FriendListResolveResponse, error) {
	tx := svc.DB.Begin()

	if in.Action != "accept" && in.Action != "reject" {
		return &friend_service.FriendListResolveResponse{
			Error: "Invalid action",
		}, nil
	}

	request := &models.FriendListRequest{}
	if err := tx.Where(
		"receiver_account_id = ? AND id = ? AND is_recalled = false AND request_status = 'pending'",
		in.ReceiverID, in.RequestID,
	).First(&request).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return &friend_service.FriendListResolveResponse{
				Error: "Request not found",
			}, nil
		} else if err != nil {
			tx.Rollback()
			return &friend_service.FriendListResolveResponse{
				Error: "Request getting request",
			}, nil
		}
	}

	var SenderID = request.SenderAccountID

	switch in.Action {
	case "accept":
		friendList := &models.FriendList{}
		if err := tx.Where(
			"((first_account_id = ? AND second_account_id = ?) OR (first_account_id = ? AND second_account_id = ?)) AND is_valid = true",
			SenderID, in.ReceiverID, in.ReceiverID, SenderID,
		).First(friendList).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				friendList := &models.FriendList{
					FirstAccountID:  uint(in.ReceiverID),
					SecondAccountID: SenderID,
				}
				if err := tx.Create(friendList).Error; err != nil {
					tx.Rollback()
					return &friend_service.FriendListResolveResponse{
						Error: "Failed to create friend list",
					}, nil
				}
			} else {
				tx.Rollback()
				if err := tx.Create(friendList).Error; err != nil {
					tx.Rollback()
					return &friend_service.FriendListResolveResponse{
						Error: "Failed to query friend list table",
					}, nil
				}
			}
		} else {
			if err := tx.Model(friendList).
				Update("is_valid", true).
				Where("id = ?", friendList.ID).
				Error; err != nil {
				tx.Rollback()
				return &friend_service.FriendListResolveResponse{
					Error: "Failed to update friend list",
				}, nil
			}
		}

		friendFollow := &models.FriendFollow{
			FirstAccountID:  SenderID,
			SecondAccountID: uint(in.ReceiverID),
		}
		if err := tx.Where(
			"first_account_id = ? AND second_account_id = ?",
			SenderID, in.ReceiverID,
		).First(friendFollow).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(friendFollow).Error; err != nil {
					tx.Rollback()
					return &friend_service.FriendListResolveResponse{
						Error: "Error when creating friend follow",
					}, nil
				}
			} else {
				tx.Rollback()
				return &friend_service.FriendListResolveResponse{
					Error: "Failed to query friend follow table",
				}, nil
			}
		} else {
			if err := tx.Model(friendFollow).Where("id = ?", friendFollow.ID).Update("is_followed", true).Error; err != nil {
				tx.Rollback()
				return &friend_service.FriendListResolveResponse{
					Error: "Failed to update friend follow table",
				}, nil
			}
		}

		friendFollowReversed := &models.FriendFollow{
			FirstAccountID:  uint(in.ReceiverID),
			SecondAccountID: SenderID,
		}
		if err := tx.Where(
			"first_account_id = ? AND second_account_id = ?",
			in.ReceiverID, SenderID,
		).First(friendFollowReversed).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(friendFollowReversed).Error; err != nil {
					tx.Rollback()
					return &friend_service.FriendListResolveResponse{
						Error: "Error when creating friend follow",
					}, nil
				}
			} else {
				tx.Rollback()
				return &friend_service.FriendListResolveResponse{
					Error: "Failed to query friend follow table",
				}, nil
			}
		} else {
			if err := tx.Model(friendFollowReversed).Where("id = ?", friendFollowReversed.ID).Update("is_followed", true).Error; err != nil {
				tx.Rollback()
				return &friend_service.FriendListResolveResponse{
					Error: "Failed to update friend follow table",
				}, nil
			}
		}

		friendBlock := &models.FriendBlock{
			FirstAccountID:  SenderID,
			SecondAccountID: uint(in.ReceiverID),
		}

		if err := tx.Where(
			"(first_account_id = ? AND second_account_id = ?)",
			SenderID, in.ReceiverID,
		).First(friendBlock).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(friendBlock).Error; err != nil {
					tx.Rollback()
					return &friend_service.FriendListResolveResponse{
						Error: "Error when creating friend block",
					}, nil
				}
			} else {
				tx.Rollback()
				return &friend_service.FriendListResolveResponse{
					Error: "Error query friend block",
				}, nil
			}
		} else {
			if err := tx.Model(friendBlock).Where("id = ?", friendBlock.ID).Update("is_blocked", false).Error; err != nil {
				tx.Rollback()
				return &friend_service.FriendListResolveResponse{
					Error: "Error updating friend block",
				}, nil
			}
		}

		friendBlockReversed := &models.FriendBlock{
			FirstAccountID:  uint(in.ReceiverID),
			SecondAccountID: SenderID,
		}

		if err := tx.Where(
			"(first_account_id = ? AND second_account_id = ?)",
			in.ReceiverID, SenderID,
		).First(friendBlockReversed).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(friendBlockReversed).Error; err != nil {
					tx.Rollback()
					return &friend_service.FriendListResolveResponse{
						Error: "Error when creating friend block",
					}, nil
				}
			} else {
				tx.Rollback()
				return &friend_service.FriendListResolveResponse{
					Error: "Error query friend block",
				}, nil
			}
		} else {
			if err := tx.Model(friendBlockReversed).Where("id = ?", friendBlockReversed.ID).Update("is_blocked", false).Error; err != nil {
				tx.Rollback()
				return &friend_service.FriendListResolveResponse{
					Error: "Error updating friend block",
				}, nil
			}
		}

		if err := tx.Model(request).Where("id = ?", request.ID).Update("request_status", "approved").Error; err != nil {
			tx.Rollback()
			return &friend_service.FriendListResolveResponse{
				Error: "Failed to update friend request status",
			}, nil
		}
		break
	case "reject":
		if err := tx.Model(request).Where("id = ?", request.ID).Update("request_status", "rejected").Error; err != nil {
			tx.Rollback()
			return &friend_service.FriendListResolveResponse{
				Error: "Failed to update friend request status",
			}, nil
		}
		break
	}

	tx.Commit()
	return &friend_service.FriendListResolveResponse{
		Success: true,
	}, nil
}

func (svc *FriendService) RecallFriendRequest(ctx context.Context, in *friend_service.RecallRequest) (*friend_service.RecallResponse, error) {
	tx := svc.DB.Begin()
	var request models.FriendListRequest

	if err := tx.Where("id = ? AND sender_account_id = ? AND request_status = 'pending' AND is_recalled = false",
		in.RequestID, in.SenderID).First(&request).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return &friend_service.RecallResponse{
				Error: "Request not found",
			}, nil
		}
	}

	if err := tx.Model(request).Update("is_recalled", true).Error; err != nil {
		tx.Rollback()
		return &friend_service.RecallResponse{
			Error: "Can't recall request",
		}, nil
	}

	tx.Commit()

	return &friend_service.RecallResponse{
		Success: true,
	}, nil
}

func (svc *FriendService) Unfriend(ctx context.Context, in *friend_service.UnfriendRequest) (*friend_service.UnfriendResponse, error) {
	tx := svc.DB.Begin()

	fromAccountID, err := strconv.ParseUint(in.FromAccountID, 10, 64)
	if err != nil {
		tx.Rollback()
		return &friend_service.UnfriendResponse{
			Error: "Invalid From Account ID",
		}, nil
	}

	toAccountID, err := strconv.ParseUint(in.ToAccountID, 10, 64)
	if err != nil {
		tx.Rollback()
		return &friend_service.UnfriendResponse{
			Error: "Invalid To Account ID",
		}, nil
	}

	relation := &models.FriendList{}
	if err := tx.Where(""+
		"((first_account_id = ? AND second_account_id = ?  AND is_valid = true) OR (first_account_id = ? AND second_account_id = ?  AND is_valid = true))",
		fromAccountID, toAccountID, toAccountID, fromAccountID,
	).First(&relation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return &friend_service.UnfriendResponse{
				Error: "Not friend already",
			}, nil
		} else {
			tx.Rollback()
			return &friend_service.UnfriendResponse{
				Error: "Error getting relation",
			}, nil
		}
	}

	if err := tx.Model(relation).Where("id = ?", relation.ID).Update("is_valid", false).Error; err != nil {
		tx.Rollback()
		return &friend_service.UnfriendResponse{
			Error: "Error updating relation",
		}, nil
	}

	if err := tx.Model(relation).Where("id = ?", relation.ID).Delete(&models.FriendList{}).Error; err != nil {
		tx.Rollback()
		return &friend_service.UnfriendResponse{
			Error: "Error unfriend",
		}, nil
	}

	if err := tx.Model(&models.FriendFollow{}).
		Where("first_account_id = ? AND second_account_id = ?", fromAccountID, toAccountID).
		Delete(&models.FriendFollow{}).
		Error; err != nil {
		tx.Rollback()
		return &friend_service.UnfriendResponse{
			Error: "Error unfollow",
		}, nil
	}

	if err := tx.Model(&models.FriendFollow{}).
		Where("first_account_id = ? AND second_account_id = ?", toAccountID, fromAccountID).
		Update("is_followed", false).
		Error; err != nil {
		tx.Rollback()
		return &friend_service.UnfriendResponse{
			Error: "Error unfollow",
		}, nil
	}

	tx.Commit()

	return &friend_service.UnfriendResponse{
		Success: true,
	}, nil
}

func (svc *FriendService) ResolveFriendFollow(ctx context.Context, in *friend_service.FriendFollowResolveRequest) (*friend_service.FriendFollowResolveResponse, error) {
	tx := svc.DB.Begin()

	if in.Action != "follow" && in.Action != "unfollow" {
		tx.Rollback()
		return &friend_service.FriendFollowResolveResponse{
			Error: "Invalid Action",
		}, nil
	}

	fromAccountID, err := strconv.ParseUint(in.FromAccountID, 10, 64)
	if err != nil {
		tx.Rollback()
		return &friend_service.FriendFollowResolveResponse{
			Error: "Invalid From Account ID",
		}, nil
	}

	toAccountID, err := strconv.ParseUint(in.ToAccountID, 10, 64)
	if err != nil {
		tx.Rollback()
		return &friend_service.FriendFollowResolveResponse{
			Error: "Invalid To Account ID",
		}, nil
	}

	switch in.Action {
	case "follow":
		relation := &models.FriendFollow{}
		err := tx.Where("first_account_id = ? AND second_account_id = ?", fromAccountID, toAccountID).First(&relation).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			relation = &models.FriendFollow{
				FirstAccountID:  uint(fromAccountID),
				SecondAccountID: uint(toAccountID),
			}
			if err := tx.Create(&relation).Error; err != nil {
				tx.Rollback()
				return &friend_service.FriendFollowResolveResponse{
					Error: "Error creating relation",
				}, nil
			}
		} else if err == nil {
			if err := tx.Model(&relation).Update("is_followed", true).Error; err != nil {
				tx.Rollback()
				return &friend_service.FriendFollowResolveResponse{
					Error: "Error updating relation",
				}, nil
			}
		} else {
			tx.Rollback()
			return &friend_service.FriendFollowResolveResponse{
				Error: "Error retrieving relation",
			}, nil
		}
		break
	case "unfollow":
		relation := &models.FriendFollow{}
		err := tx.Where("first_account_id = ? AND second_account_id = ?", fromAccountID, toAccountID).First(&relation).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			relation = &models.FriendFollow{
				FirstAccountID:  uint(fromAccountID),
				SecondAccountID: uint(toAccountID),
				IsFollowed:      false,
			}
			if err := tx.Create(&relation).Error; err != nil {
				tx.Rollback()
				return &friend_service.FriendFollowResolveResponse{
					Error: "Error creating relation",
				}, nil
			}
		} else if err == nil {
			if err := tx.Model(&relation).Update("is_followed", false).Error; err != nil {
				tx.Rollback()
				return &friend_service.FriendFollowResolveResponse{
					Error: "Error updating relation",
				}, nil
			}
		} else {
			tx.Rollback()
			return &friend_service.FriendFollowResolveResponse{
				Error: "Error retrieving relation",
			}, nil
		}
		break
	}
	tx.Commit()
	return &friend_service.FriendFollowResolveResponse{
		Success: true,
	}, nil
}
func (svc *FriendService) ResolveFriendBlock(ctx context.Context, in *friend_service.FriendBlockResolveRequest) (*friend_service.FriendBlockResolveResponse, error) {
	tx := svc.DB.Begin()

	if in.Action != "block" && in.Action != "unblock" {
		tx.Rollback()
		return &friend_service.FriendBlockResolveResponse{
			Error: "Invalid Action",
		}, nil
	}

	fromAccountID, err := strconv.ParseUint(in.FromAccountID, 10, 64)
	if err != nil {
		tx.Rollback()
		return &friend_service.FriendBlockResolveResponse{
			Error: "Invalid From Account ID",
		}, nil
	}

	toAccountID, err := strconv.ParseUint(in.ToAccountID, 10, 64)
	if err != nil {
		tx.Rollback()
		return &friend_service.FriendBlockResolveResponse{
			Error: "Invalid To Account ID",
		}, nil
	}

	relation := &models.FriendBlock{}
	err = tx.Where("first_account_id = ? AND second_account_id = ?", fromAccountID, toAccountID).First(&relation).Error

	switch in.Action {
	case "block":
		if errors.Is(err, gorm.ErrRecordNotFound) {
			relation = &models.FriendBlock{
				FirstAccountID:  uint(fromAccountID),
				SecondAccountID: uint(toAccountID),
				IsBlocked:       true,
			}
			if err := tx.Create(&relation).Error; err != nil {
				tx.Rollback()
				return &friend_service.FriendBlockResolveResponse{
					Error: "Error creating relation",
				}, nil
			}
		} else if err == nil {
			if err := tx.Model(&relation).Update("is_blocked", true).Error; err != nil {
				tx.Rollback()
				return &friend_service.FriendBlockResolveResponse{
					Error: "Error blocking",
				}, nil
			}
		} else {
			tx.Rollback()
			return &friend_service.FriendBlockResolveResponse{
				Error: "Error retrieving relation",
			}, nil
		}

		var friend models.FriendList
		if err := tx.Model(&friend).Where("(first_account_id = ? AND second_account_id = ?) OR (first_account_id = ? AND second_account_id = ?)",
			fromAccountID, toAccountID, toAccountID, fromAccountID).First(&friend).Error; err != nil {
		} else {
			if err := tx.Model(&friend).Delete(&friend).Error; err != nil {
				tx.Rollback()
				return &friend_service.FriendBlockResolveResponse{
					Error: "Error deleting relation",
				}, nil
			}
		}

	case "unblock":
		if errors.Is(err, gorm.ErrRecordNotFound) {
			relation = &models.FriendBlock{
				FirstAccountID:  uint(fromAccountID),
				SecondAccountID: uint(toAccountID),
				IsBlocked:       false,
			}
			if err := tx.Create(&relation).Error; err != nil {
				tx.Rollback()
				return &friend_service.FriendBlockResolveResponse{
					Error: "Error creating relation",
				}, nil
			}
		} else if err == nil {
			if err := tx.Model(&relation).Update("is_blocked", false).Error; err != nil {
				tx.Rollback()
				return &friend_service.FriendBlockResolveResponse{
					Error: "Error unblocking",
				}, nil
			}
		} else {
			tx.Rollback()
			return &friend_service.FriendBlockResolveResponse{
				Error: "Error retrieving relation",
			}, nil
		}
	}

	tx.Commit()

	return &friend_service.FriendBlockResolveResponse{
		Success: true,
	}, nil
}
func (svc *FriendService) GetPendingList(ctx context.Context, in *friend_service.GetPendingListRequest) (*friend_service.GetPendingListResponse, error) {
	tx := svc.DB.Begin()
	accountID, err := strconv.ParseUint(in.AccountID, 10, 64)
	if err != nil {
		tx.Rollback()
		return &friend_service.GetPendingListResponse{
			Error: "Invalid AccountID",
		}, nil
	}

	page := in.Page
	if page == 0 {
		page = 1
	}
	pageSize := int64(10)

	offset := (page - 1) * pageSize

	var pendingList []models.FriendListRequest

	if err := tx.Model(&models.FriendListRequest{}).
		Where("receiver_account_id = ? AND request_status = 'pending' AND is_recalled = false", accountID).
		Limit(int(pageSize)).
		Offset(int(offset)).
		Find(&pendingList).Error; err != nil {
		tx.Rollback()
		return &friend_service.GetPendingListResponse{
			Error: "Error getting pending list",
		}, nil
	}

	responseType := make([]*friend_service.PendingData, len(pendingList))
	for i, record := range pendingList {
		mutualCount, err := svc.countMutualFriends(record.SenderAccountID, uint(accountID))
		if err != nil {
			mutualCount = 0
		}
		responseType[i] = &friend_service.PendingData{
			RequestID:     uint64(record.ID),
			AccountID:     uint64(record.SenderAccountID),
			CreatedAt:     record.CreatedAt.Unix(),
			MutualFriends: mutualCount,
		}
	}

	tx.Commit()

	return &friend_service.GetPendingListResponse{
		Page:        in.Page,
		ListPending: responseType,
	}, nil
}

func (svc *FriendService) GetListFriend(ctx context.Context, in *friend_service.GetListFriendRequest) (*friend_service.GetListFriendResponse, error) {
	tx := svc.DB.Begin()

	accountID, err := strconv.ParseUint(in.AccountID, 10, 64)
	if err != nil {
		tx.Rollback()
		return &friend_service.GetListFriendResponse{
			Error: "Invalid AccountID",
		}, nil
	}

	var friendList []models.FriendList
	if err := tx.Model(friendList).
		Select("first_account_id, second_account_id").
		Where(
			"(first_account_id = ? OR second_account_id = ?) AND is_valid = true",
			accountID, accountID,
		).Find(&friendList).Error; err != nil {
		tx.Rollback()
		return &friend_service.GetListFriendResponse{
			Error: "Error getting friends",
		}, nil
	}

	friendIDs := make(map[uint64]struct{})
	for _, friend := range friendList {
		if friend.FirstAccountID == uint(accountID) {
			friendIDs[uint64(friend.SecondAccountID)] = struct{}{}
		} else {
			friendIDs[uint64(friend.FirstAccountID)] = struct{}{}
		}
	}

	var blockList []models.FriendBlock
	if err := tx.Model(&models.FriendBlock{}).
		Select("first_account_id, second_account_id").
		Where("(first_account_id = ? OR second_account_id = ?) AND is_blocked = true", accountID, accountID).
		Find(&blockList).Error; err != nil {
		tx.Rollback()
		return &friend_service.GetListFriendResponse{
			Error: "Error retrieving block list",
		}, nil
	}

	blockedIDs := make(map[uint64]struct{})
	for _, block := range blockList {
		if block.FirstAccountID == uint(accountID) {
			blockedIDs[uint64(block.SecondAccountID)] = struct{}{}
		} else {
			blockedIDs[uint64(block.FirstAccountID)] = struct{}{}
		}
	}

	var unblockList []uint64
	for id := range friendIDs {
		if _, isBlocked := blockedIDs[id]; !isBlocked {
			unblockList = append(unblockList, id)
		}
	}

	responseFriends := make([]string, len(unblockList))
	for i, id := range unblockList {
		responseFriends[i] = fmt.Sprintf("%d", id)
	}

	tx.Commit()

	return &friend_service.GetListFriendResponse{
		ListFriendIDs: responseFriends,
	}, nil
}

func (svc *FriendService) CountPending(ctx context.Context, in *friend_service.CountPendingRequest) (*friend_service.CountPendingResponse, error) {
	var quantity int64

	err := svc.DB.Model(&models.FriendListRequest{}).
		Where("receiver_account_id = ? AND is_recalled = ? AND request_status = ?",
			in.AccountID, false, "pending").
		Count(&quantity).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count pending friend requests: %w", err)
	}

	return &friend_service.CountPendingResponse{
		Quantity: int32(quantity),
	}, nil
}

func (svc *FriendService) countMutualFriends(account1 uint, account2 uint) (int64, error) {
	var mutualFriendsCount int64

	friendsOfAccount1 := svc.DB.Model(&models.FriendList{}).
		Select("CASE WHEN first_account_id = ? THEN second_account_id ELSE first_account_id END AS friend_id", account1).
		Where("(first_account_id = ? OR second_account_id = ?) AND is_valid = true", account1, account1)

	friendsOfAccount2 := svc.DB.Model(&models.FriendList{}).
		Select("CASE WHEN first_account_id = ? THEN second_account_id ELSE first_account_id END AS friend_id", account2).
		Where("(first_account_id = ? OR second_account_id = ?) AND is_valid = true", account2, account2)

	err := svc.DB.Table("(?) AS friends1", friendsOfAccount1).
		Joins("JOIN (?) AS friends2 ON friends1.friend_id = friends2.friend_id", friendsOfAccount2).
		Count(&mutualFriendsCount).Error

	return mutualFriendsCount, err

}

func (svc *FriendService) CheckIsFriend(ctx context.Context, in *friend_service.CheckIsFriendRequest) (*friend_service.CheckIsFriendResponse, error) {
	var friendList models.FriendList

	err := svc.DB.Where("(first_account_id = ? AND second_account_id = ?) OR (first_account_id = ? AND second_account_id = ?)",
		in.FirstAccountID, in.SecondAccountID, in.SecondAccountID, in.FirstAccountID).
		Where("is_valid = ?", true).
		First(&friendList).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &friend_service.CheckIsFriendResponse{
				IsFriend: false,
			}, nil
		}
		return nil, err
	}

	return &friend_service.CheckIsFriendResponse{
		IsFriend: true,
	}, nil
}

func (svc *FriendService) CheckIsBlock(ctx context.Context, in *friend_service.CheckIsBlockedRequest) (*friend_service.CheckIsBlockedResponse, error) {
	if in.FirstAccountID == 0 || in.SecondAccountID == 0 {
		return &friend_service.CheckIsBlockedResponse{
			Error: "Invalid account IDs",
		}, errors.New("invalid account IDs")
	}
	fmt.Printf("FA: %v, SA: %v", in.FirstAccountID, in.SecondAccountID)

	var friendBlock models.FriendBlock
	if err := svc.DB.Model(models.FriendBlock{}).Where("(first_account_id = ? AND second_account_id = ? AND is_blocked = ?) OR (first_account_id = ? AND second_account_id = ? AND is_blocked = ?)",
		in.FirstAccountID, in.SecondAccountID, true, in.SecondAccountID, in.FirstAccountID, true).
		First(&friendBlock).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &friend_service.CheckIsBlockedResponse{
				IsBlocked: false,
			}, nil
		}
		return &friend_service.CheckIsBlockedResponse{
			IsBlocked: true,
		}, nil
	}

	return &friend_service.CheckIsBlockedResponse{
		IsBlocked: true,
	}, nil
}

func (svc *FriendService) CheckIsFollow(ctx context.Context, in *friend_service.CheckIsFollowRequest) (*friend_service.CheckIsFollowResponse, error) {
	// Validate input IDs
	if in.FromAccountID <= 0 || in.ToAccountID <= 0 {
		return &friend_service.CheckIsFollowResponse{
			Error: "Invalid account IDs",
		}, errors.New("invalid account IDs")
	}

	fmt.Printf("Checking if account %d follows account %d\n", in.FromAccountID, in.ToAccountID)

	// Check if a follow relationship exists
	var exists bool
	err := svc.DB.Model(&models.FriendFollow{}).
		Select("COUNT(*) > 0").
		Where("first_account_id = ? AND second_account_id = ? AND is_followed = ?", in.FromAccountID, in.ToAccountID, true).
		Find(&exists).Error

	// Handle database error
	if err != nil {
		fmt.Printf("Database error: %v\n", err)
		return &friend_service.CheckIsFollowResponse{
			IsFollow: false,
		}, err
	}

	// Return the follow status
	return &friend_service.CheckIsFollowResponse{
		IsFollow: exists,
	}, nil
}

func (svc *FriendService) GetUserInteraction(ctx context.Context, in *friend_service.GetUserInteractionRequest) (*friend_service.GetUserInteractionResponse, error) {
	var interactions []*friend_service.InteractionScore

	// Query all the users the account (in.AccountID) could have interactions with
	var allUsers []models.FriendList
	err := svc.DB.WithContext(ctx).Where("first_account_id = ? OR second_account_id = ?", in.AccountID, in.AccountID).Find(&allUsers).Error
	if err != nil {

	}

	// Map to track processed users and their interaction score
	processedUsers := make(map[uint64]bool)

	// Iterate over all users to calculate the interaction score
	for _, interaction := range allUsers {
		var targetAccountID uint
		// Determine the target account ID (the other user in the interaction)
		if uint64(interaction.FirstAccountID) == in.AccountID {
			targetAccountID = interaction.SecondAccountID
		} else {
			targetAccountID = interaction.FirstAccountID
		}

		// Skip if this target account has already been processed
		if processedUsers[uint64(targetAccountID)] {
			continue
		}
		processedUsers[uint64(targetAccountID)] = true

		score := int64(0)

		var blockStatus models.FriendBlock
		err := svc.DB.WithContext(ctx).Where("(first_account_id = ? AND second_account_id = ?) OR (second_account_id = ? AND first_account_id = ?)", in.AccountID, targetAccountID, targetAccountID, in.AccountID).First(&blockStatus).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			blockStatus.IsBlocked = false
		}

		if blockStatus.IsBlocked {
			score = -10000
			continue
		} else {
			var followStatus models.FriendFollow
			err := svc.DB.WithContext(ctx).Where("first_account_id = ? AND second_account_id = ? AND is_followed = ?", in.AccountID, targetAccountID, true).First(&followStatus).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					followStatus.IsFollowed = false
					continue
				}
				return nil, err
			}

			if !followStatus.IsFollowed {
				score = -10000
			} else {
				score += 200
				var friendStatus models.FriendList
				err := svc.DB.WithContext(ctx).Where("(first_account_id = ? AND second_account_id = ?) OR (first_account_id = ? AND second_account_id = ?)", in.AccountID, targetAccountID, targetAccountID, in.AccountID).First(&friendStatus).Error
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
					} else {
						return nil, err
					}
				}

				if friendStatus.IsValid {
					score += 100
				}
			}
		}

		if score > 0 {
			interactions = append(interactions, &friend_service.InteractionScore{
				AccountID: uint64(targetAccountID),
				Score:     uint64(score),
			})
		}
	}

	sort.Slice(interactions, func(i, j int) bool {
		return interactions[i].Score > interactions[j].Score
	})

	return &friend_service.GetUserInteractionResponse{
		Interactions: interactions,
	}, nil
}

func (svc *FriendService) CheckExistingRequest(ctx context.Context, in *friend_service.CheckExistingRequestRequest) (*friend_service.CheckExistingRequestResponse, error) {
	if in.ToAccountID <= 0 || in.FromAccountID <= 0 {
		return nil, errors.New("invalid account IDs")
	}

	var request models.FriendListRequest
	if err := svc.DB.Model(request).Where("sender_account_id = ? AND receiver_account_id = ? AND is_recalled = ? AND request_status = ?", in.FromAccountID, in.ToAccountID, false, "pending").First(&request).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &friend_service.CheckExistingRequestResponse{
				IsExisting: false,
				RequestID:  0,
			}, nil
		} else {
			return nil, err
		}
	}

	return &friend_service.CheckExistingRequestResponse{
		IsExisting: true,
		RequestID:  uint64(request.ID),
	}, nil
}

func (svc *FriendService) GetBlockList(ctx context.Context, in *friend_service.GetBlockListRequest) (*friend_service.BlockListResponse, error) {
	if in.AccountID == 0 {
		return nil, errors.New("invalid AccountID")
	}

	var blockedIDs []uint32
	err := svc.DB.Model(&models.FriendBlock{}).
		Where("first_account_id = ? AND is_blocked = ?", in.AccountID, true).
		Pluck("second_account_id", &blockedIDs).Error

	if err != nil {
		return nil, err
	}

	return &friend_service.BlockListResponse{IDs: blockedIDs}, nil
}

func (svc *FriendService) GetBlockedList(ctx context.Context, in *friend_service.GetBlockedListRequest) (*friend_service.BlockListResponse, error) {
	if in.AccountID == 0 {
		return nil, errors.New("invalid AccountID")
	}

	var blockedByIDs []uint32
	err := svc.DB.Model(&models.FriendBlock{}).
		Where("second_account_id = ? AND is_blocked = ?", in.AccountID, true).
		Pluck("first_account_id", &blockedByIDs).Error

	if err != nil {
		return nil, err
	}

	return &friend_service.BlockListResponse{IDs: blockedByIDs}, nil
}
