package handlers

import (
	"context"
	"log/slog"

	"github.com/minio/minio-go/v7"
	"github.com/vladgrskkh/movie_recomendation_system/imageservice/internal/service"

	pb "github.com/vladgrskkh/movie-recommender-contracts/v1/imageservice"
)

type ImageHandler struct {
	logger       *slog.Logger
	imageService *service.ImageService
}

func NewImageHandler(logger *slog.Logger, imageService *service.ImageService) *ImageHandler {
	return &ImageHandler{
		logger:       logger,
		imageService: imageService,
	}
}

func (h *ImageHandler) Upload(ctx context.Context, r *pb.ImageUploadRequest) (*pb.ImageUploadResponse, error) {
	image := r.GetImage()
	err := h.imageService.UploadImage(ctx, image.GetBucketName(), image.GetObjectName(), r.GetImageBytes(), minio.PutObjectOptions{})
}
