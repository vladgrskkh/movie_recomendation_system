package handlers

import (
	"context"
	"errors"
	"io"
	"log/slog"

	"github.com/vladgrskkh/movie_recomendation_system/imageservice/internal/domain"
	"github.com/vladgrskkh/movie_recomendation_system/imageservice/internal/service"
	"google.golang.org/grpc"
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

func (h *ImageHandler) Upload(stream grpc.ClientStreamingServer[pb.ImageUploadRequest, pb.ImageUploadResponse]) error {
	pbImageResp, err := stream.Recv()
	if err != nil {
		return status.Error(codes.Internal, "server encountered a problem and could not process your request")
	}

	pbImageMetadata := pbImageResp.GetImage()
	imageMetadata := domain.NewImageMetadata(pbImageMetadata.GetObjectName(), pbImageMetadata.GetBucketName(), pbImageMetadata.GetFormat(), 0)
	// TODO: get size from response
	imageBytes := make([]byte, 1024*1024*3)
forLoop:
	for {
		buf, err := stream.Recv()
		if err != nil {
			switch {
			case errors.Is(err, io.EOF):
				break forLoop
			default:
				return status.Error(codes.Internal, "error processing image chunk")
			}
		}

		imageBytes = append(imageBytes, buf.GetChunk().Chunk...)
	}

	imageMetadata.Size = int64(len(imageBytes))
	err = h.imageService.UploadImage(context.Background(), imageMetadata, []byte{})
	if err != nil {
		return status.Error(codes.Internal, "server encountered a problem and could not process your request")
	}

	return nil
}

// TODO: change proto contract so that request is not common.Image instead pb.ImageGetRequest
// TODO: change proto contract (message field)
func (h *ImageHandler) Get(r *common.Image, stream grpc.ServerStreamingServer[pb.ImageGetResponse]) error {
	imageMetadata := domain.NewImageMetadata(r.GetObjectName(), r.GetBucketName(), r.GetFormat(), 0)
	reader, err := h.imageService.GetImage(context.Background(), imageMetadata)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrorImageNotFound):
			return status.Error(codes.NotFound, "image not found")
		default:
			return status.Error(codes.Internal, "server encountered a problem and could not process your request")
		}
	}

	buf := make([]byte, 1024*32)
forLoop:
	for {
		n, err := reader.Read(buf)
		if err != nil {
			switch {
			case errors.Is(err, io.EOF):
				break forLoop
			default:
				return status.Error(codes.Internal, "server encountered a problem and could not process your request")
			}
		}

		// TODO: do it better
		err = stream.Send(&pb.ImageGetResponse{
			Payload: &pb.ImageGetResponse_Chunk{
				Chunk: &pb.ImageChunk{
					Chunk: buf[:n],
				},
			},
		})
		if err != nil {
			return status.Error(codes.Internal, "server encountered a problem and could not process your request")
		}
	}

	err = stream.Send(&pb.ImageGetResponse{
		Payload: &pb.ImageGetResponse_Message{
			Message: "successfully send image",
		},
	})
	if err != nil {
		return status.Error(codes.Internal, "server encountered a problem and could not process your request")
	}

	return nil
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
