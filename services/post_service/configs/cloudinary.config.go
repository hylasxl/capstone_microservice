package configs

import (
	"context"
	"fmt"
	"github.com/cloudinary/cloudinary-go/v2"
	"os"
)

type CloudinaryService struct {
	Client *cloudinary.Cloudinary
	Ctx    context.Context
}

func InitCloudinary(ctx context.Context) (*CloudinaryService, error) {
	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudinary: %v", err)
	}
	return &CloudinaryService{Client: cld, Ctx: ctx}, nil
}
