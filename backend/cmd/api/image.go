package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"path"

	"github.com/go-chi/chi/v5"
	"github.com/vladgrskkh/movie-recommender-contracts/common"
	pb "github.com/vladgrskkh/movie-recommender-contracts/v1/imageservice"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UploadImageHandler godoc
// @Summary Upload image
// @Description Upload an image file (jpg or png). Multipart form: field name `image`.
// @Tags images
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Image file"
// @Success 201 {object} map[string]string "Created | Example {\"message\": \"...\"}"
// @Failure 400 {object} map[string]string "Bad Request | Example {\"error\": \"...\"}"
// @Failure 401 {object} map[string]string "Unauthorized | Example {\"error\": \"this resourse avaliable only for authenticated users\"}"
// @Failure 415 {object} map[string]string "Unsupported Media Type | Example {\"error\": \"only jpeg and png images are allowed\"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {"error": "server encountered a problem and could not process your request"}"
// @Security BearerAuth
// @Router /images [post]
func (app *application) UploadImageHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	file, header, err := r.FormFile("image")
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	app.deferClose(file.Close, err)

	// need 512 bytes to determent content type
	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if valid := app.validateImage(buf); !valid {
		app.imageUnsupportedMediaTypeResponse(w, r)
		return
	}

	// set to beggining after validating content type
	_, err = file.Seek(0, 0)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.logger.Info("file metadata", slog.Int64("size", header.Size), slog.String("filename", header.Filename))

	stream, err := app.imageClient.Upload(context.Background())
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	ext := path.Ext(header.Filename)
	if ext == "" {
		app.badRequestResponse(w, r, err)
	}

	err = stream.Send(&pb.ImageUploadRequest{
		Payload: &pb.ImageUploadRequest_Image{
			Image: &common.Image{
				ObjectName: header.Filename,
				BucketName: "images",
				Format:     ext[1:],
				Size:       header.Size,
			},
		},
	})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	buf = make([]byte, 1024*32)

forLoop:
	for {
		n, err := file.Read(buf)
		if err != nil {
			switch {
			case errors.Is(err, io.EOF):
				break forLoop
			default:
				app.serverErrorResponse(w, r, err)
				return
			}
		}

		err = stream.Send(&pb.ImageUploadRequest{
			Payload: &pb.ImageUploadRequest_Chunk{
				Chunk: &pb.ImageChunk{Chunk: buf[:n]},
			},
		})
		if err != nil {
			switch {
			case status.Code(err) == codes.InvalidArgument:
				app.badRequestResponse(w, r, err)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"message": res.Message}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// GetImageHandler godoc
//
// @Summary Get image
// @Description Retrieve an image by providing file name
// @Tags images
// @Produce image/png, image/jpg, application/json
// @Param imageID path string true "Image ID"
// @Success 200 {file} file "Image file"
// @Failure 400 {object} map[string]string "Bad Request | Example {\"error\": \"...\"}"
// @Failure 401 {object} map[string]string "Unauthorized | Example {\"error\": \"this resourse avaliable only for authenticated users\"}"
// @Failure 404 {object} map[string]string "Not Found | Example {\"error\": \"requested resource could not be found\"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {\"error\": \"server encountered a problem and could not process your request\"}"
// @Security BearerAuth
// @Router /images/{imageID} [get]
func (app *application) GetImageHandler(w http.ResponseWriter, r *http.Request) {
	imageID := chi.URLParam(r, "imageID")

	stream, err := app.imageClient.Get(context.Background(), &common.Image{ObjectName: imageID, BucketName: "images"})
	if err != nil {
		switch {
		case status.Code(err) == codes.NotFound:
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	ext := path.Ext(imageID)
	if ext == "" || ext[1:] != "jpg" && ext[1:] != "png" {
		app.badRequestResponse(w, r, errors.New("only jpg and png images are allowed"))
		return
	}

	w.Header().Set("Content-Type", "image/"+ext[1:])
	w.WriteHeader(http.StatusOK)

forLoop:
	for {
		buf, err := stream.Recv()
		if err != nil {
			switch {
			case errors.Is(err, io.EOF):
				break forLoop
			default:
				app.serverErrorResponse(w, r, err)
				return
			}
		}

		_, err = w.Write(buf.GetChunk().Chunk)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}
}

// DeleteImageHandler godoc
//
// @Summary Delete image
// @Description Delete image by providing file name
// @Tags images
// @Produce json
// @Param imageID path string true "Image ID"
// @Success 200 {object} map[string]string "OK | Example {\"message\": \"image deleted\"}"
// @Failure 400 {object} map[string]string "Bad Request | Example {\"error\": \"...\"}"
// @Failure 401 {object} map[string]string "Unauthorized | Example {\"error\": \"this resourse avaliable only for authenticated users\"}"
// @Failure 404 {object} map[string]string "Not Found | Example {\"error\": \"requested resource could not be found\"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {\"error\": \"server encountered a problem and could not process your request\"}"
// @Security BearerAuth
// @Router /images/{imageID} [delete]
func (app *application) DeleteImageHandler(w http.ResponseWriter, r *http.Request) {
	// check for appropriate id
	imageID := chi.URLParam(r, "imageID")

	res, err := app.imageClient.Delete(context.Background(), &common.Image{ObjectName: imageID, BucketName: "images"})
	if err != nil {
		switch {
		case status.Code(err) == codes.NotFound:
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusNotImplemented, envelope{"message": res.Message}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
