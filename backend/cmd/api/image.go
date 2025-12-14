package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path"

	"github.com/vladgrskkh/movie-recommender-contracts/common"
	pb "github.com/vladgrskkh/movie-recommender-contracts/v1/imageservice"
)

func (app *application) UploadImageHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	contentType := r.Header.Get("Content-Type")
	if contentType != "image/jpg" && contentType != "image/png" {
		app.imageUnsupportedMediaTypeResponse(w, r)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// TODO: better error handling, for now its okey
	defer func() {
		e := file.Close()
		if err != nil && e != nil {
			app.logger.Error(fmt.Errorf("previous error: %w; close error: %w", err, e).Error())
		} else if e != nil {
			app.logger.Error(e.Error())
		}
	}()

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
			},
		},
	})
	// TODO: think about what error server can return
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	buf := make([]byte, 1024*32)

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
			app.serverErrorResponse(w, r, err)
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

func (app *application) GetImageHandler(w http.ResponseWriter, r *http.Request) {
	imageID := r.URL.Query().Get("imageID")

	stream, err := app.imageClient.Get(context.Background(), &common.Image{ObjectName: imageID, BucketName: "images"})
	// TODO: check what error is returned here from the server(it can be not found or something else so i need to
	// handle it better)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	ext := path.Ext(imageID)
	if ext == "" {
		app.badRequestResponse(w, r, err)
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

func (app *application) DeleteImageHandler(w http.ResponseWriter, r *http.Request) {
	imageID := r.URL.Query().Get("imageID")

	// TODO: better error handling (not found etc)
	res, err := app.imageClient.Delete(context.Background(), &common.Image{ObjectName: imageID, BucketName: "images"})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusNotImplemented, envelope{"message": res.Message}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
