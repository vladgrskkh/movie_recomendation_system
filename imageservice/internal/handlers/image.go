package handlers

import (
	"context"
	"log/slog"

	"github.com/minio/minio-go/v7"
	"github.com/vladgrskkh/movie_recomendation_system/imageservice/internal/service"

	"github.com/vladgrskkh/movie-recommender-contracts/common"
	pb "github.com/vladgrskkh/movie-recommender-contracts/v1/imageservice"
)

type ImageHandler struct {
	logger       *slog.Logger
	imageService *service.ImageService
	pb.UnimplementedImageServer
}

func NewImageHandler(logger *slog.Logger, imageService *service.ImageService) *ImageHandler {
	return &ImageHandler{
		logger:       logger,
		imageService: imageService,
	}
}

func (h *ImageHandler) Upload(ctx context.Context, r *pb.ImageUploadRequest) (*pb.ImageUploadResponse, error) {
	image := r.GetImage()
	// TODO: discard minio opts
	err := h.imageService.UploadImage(ctx, image.GetBucketName(), image.GetObjectName(), r.GetImageBytes(), minio.PutObjectOptions{})
	if err != nil {
		// TODO: look into grpc status codes
		return nil, err
	}

	return &pb.ImageUploadResponse{
		Image:   image,
		Message: "Successfully uploaded image",
	}, nil
}

// TODO: change proto contract so that request is not common.Image instead pb.ImageGetRequest
func (h *ImageHandler) Get(ctx context.Context, r *common.Image) (*pb.ImageGetResponse, error) {
	// TODO: discard minio opts
	image, err := h.imageService.GetImage(ctx, r.GetBucketName(), r.GetObjectName(), minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return &pb.ImageGetResponse{
		ImageBytes: image,
		Message:    "Successfully got image",
	}, nil
}

func (h *ImageHandler) Delete(ctx context.Context, r *common.Image) (*pb.ImageDeleteResponse, error) {
	// TODO: discard minio opts
	err := h.imageService.DeleteImage(ctx, r.GetBucketName(), r.GetObjectName(), minio.RemoveObjectOptions{})
	if err != nil {
		return nil, err
	}

	return &pb.ImageDeleteResponse{
		Message: "Successfully deleted image",
	}, nil
}

// TODO: better error handling (grpc status codes)
