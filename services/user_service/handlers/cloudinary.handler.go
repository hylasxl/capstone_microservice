package handlers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
)

type CloudinaryService struct {
	Client *cloudinary.Cloudinary
	Ctx    context.Context
}

func (cs *CloudinaryService) CompressImage(data []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}
	img = resize.Thumbnail(1200, 1200, img, resize.Lanczos3)

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 75})
	if err != nil {
		return nil, fmt.Errorf("failed to compress image: %v", err)
	}
	return buf.Bytes(), nil
}

func (cs *CloudinaryService) UploadAvatar(data []byte) (string, error) {
	if len(data) > 10*1024*1024 {
		var compressData []byte
		var err error
		compressData, err = cs.CompressImage(data)
		if err != nil {
			return "", err
		}
		data = compressData
	}
	imageReader := bytes.NewReader(data)
	uploadParams := uploader.UploadParams{
		Folder:         "user_profile_image",
		Transformation: "c_limit,w_1200,h_1200,q_auto:low",
	}
	uploadResult, err := cs.Client.Upload.Upload(cs.Ctx, imageReader, uploadParams)
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %v", err)
	}
	return uploadResult.SecureURL, nil
}
