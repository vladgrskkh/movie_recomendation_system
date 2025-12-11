package handlers

import (
	"context"
	"errors"
	"log/slog"

	"github.com/vladgrskkh/movie_recomendation_system/imageservice/internal/domain"
	"github.com/vladgrskkh/movie_recomendation_system/imageservice/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	pbImage := r.GetImage()

	imageMetadata := domain.NewImageMetadata(pbImage.ObjectName, pbImage.BucketName, pbImage.Format, int64(len(r.ImageBytes)))
	err := h.imageService.UploadImage(ctx, imageMetadata, r.ImageBytes)
	if err != nil {
		return nil, status.Error(codes.Internal, "server encountered a problem and could not process your request")
	}

	return &pb.ImageUploadResponse{
		Image:   pbImage,
		Message: "Successfully uploaded image",
	}, nil
}

// TODO: change proto contract so that request is not common.Image instead pb.ImageGetRequest
func (h *ImageHandler) Get(ctx context.Context, r *common.Image) (*pb.ImageGetResponse, error) {
	imageMetadata := domain.NewImageMetadata(r.GetObjectName(), r.GetBucketName(), r.GetFormat(), 0)
	image, err := h.imageService.GetImage(ctx, imageMetadata)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrorImageNotFound):
			return nil, status.Error(codes.NotFound, "image not found")
		default:
			return nil, status.Error(codes.Internal, "server encountered a problem and could not process your request")
		}
	}

	return &pb.ImageGetResponse{
		ImageBytes: image,
		Message:    "Successfully got image",
	}, nil
}

func (h *ImageHandler) Delete(ctx context.Context, r *common.Image) (*pb.ImageDeleteResponse, error) {
	imageMetadata := domain.NewImageMetadata(r.GetObjectName(), r.GetBucketName(), r.GetFormat(), 0)
	err := h.imageService.DeleteImage(ctx, imageMetadata)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrorImageNotFound):
			return nil, status.Error(codes.NotFound, "image not found")
		default:
			return nil, status.Error(codes.Internal, "server encountered a problem and could not process your request")
		}
	}

	return &pb.ImageDeleteResponse{
		Message: "Successfully deleted image",
	}, nil
}
