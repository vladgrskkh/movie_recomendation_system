package domain

type ImageMetadata struct {
	Name   string
	Bucket string
	Format string
	Size   int64
}

func NewImageMetadata(name string, bucket string, format string, size int64) *ImageMetadata {
	return &ImageMetadata{
		Name:   name,
		Bucket: bucket,
		Size:   size,
		Format: format,
	}
}
