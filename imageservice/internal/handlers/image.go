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
	imageMetadata := domain.NewImageMetadata(pbImageMetadata.GetObjectName(), pbImageMetadata.GetBucketName(), pbImageMetadata.GetFormat(), pbImageMetadata.GetSize())

	pr, pw := io.Pipe()
	errChan := make(chan error)
	go func() {
		// TODO: need to check if error occur will it close reader buffer
		errChan <- h.imageService.UploadImage(context.Background(), imageMetadata, pr)
	}()

forLoop:
	for {
		buf, err := stream.Recv()
		if err != nil {
			switch {
			case errors.Is(err, io.EOF):
				// race condition here if i don't write empty buffer(close pipe before minio can read from it
				// resulting in mising last chunk)
				n, err := pw.Write([]byte{})
				if err != nil {
					return status.Error(codes.Internal, "server encountered a problem and could not process your request")
				}
				h.logger.Info("zero chunk written to pipe", slog.Int("size", n))
				break forLoop
			default:
				err = pw.CloseWithError(err)
				if err != nil {
					h.logger.Error("error closing pipe with error", slog.String("error", err.Error()))
				}
				return status.Error(codes.Internal, "error processing image chunk")
			}
		}

		n, err := pw.Write(buf.GetChunk().Chunk)
		h.logger.Info("image chunk written to pipe", slog.Int("size", n))
		if err != nil {
			h.logger.Error("error writting image chunk to pipe", slog.Any("error", err))
			break forLoop
		}
	}

	err = pw.Close()
	if err != nil {
		return status.Error(codes.Internal, "server encountered a problem and could not process your request")
	}
	if err = <-errChan; err != nil {
		switch {
		case errors.Is(err, service.ErrImageCredentials):
			return status.Error(codes.InvalidArgument, "non existing bucket or invalid image creds")
		default:
			return status.Error(codes.Internal, "server encountered a problem and could not process your request")
		}
	}

	err = stream.SendAndClose(&pb.ImageUploadResponse{
		Message: "succesfully upload image",
	})
	if err != nil {
		return status.Error(codes.Internal, "server encountered a problem and could not process your request")
	}
	return nil
}

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
		h.logger.Info("image chunk read from minio reader", slog.Int("size", n))
		if err != nil {
			switch {
			case errors.Is(err, io.EOF):
				// import here to write last chunk to stream(it signals EOF when NEXT chunk is empty)
				// as far as I get it right
				h.logger.Info("EOF reached", slog.Int("size", n))
				err = stream.Send(&pb.ImageGetResponse{
					Payload: &pb.ImageGetResponse_Chunk{
						Chunk: &pb.ImageChunk{Chunk: buf[:n]},
					},
				})
				if err != nil {
					return status.Error(codes.Internal, "server encountered a problem and could not process your request")
				}
				break forLoop
			default:
				return status.Error(codes.Internal, "server encountered a problem and could not process your request")
			}
		}

		err = stream.Send(&pb.ImageGetResponse{
			Payload: &pb.ImageGetResponse_Chunk{
				Chunk: &pb.ImageChunk{Chunk: buf[:n]},
			},
		})
		if err != nil {
			return status.Error(codes.Internal, "server encountered a problem and could not process your request")
		}
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
		Message: "successfully delete image",
	}, nil
}
