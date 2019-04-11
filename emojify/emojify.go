package emojify

import (
	"fmt"
	"image"
	"image/draw"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/emojify-app/face-detection/client"

	"github.com/nfnt/resize"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Emojify defines an interface for emojify operations
type Emojify interface {
	GetFaces(f io.ReadSeeker) ([]image.Rectangle, error)
	Emojimise(image.Image, []image.Rectangle) (image.Image, error)
	Health() (int, error)
}

// Impl implements the Emojify interface
type Impl struct {
	emojis         []image.Image
	fetcher        Fetcher
	fd             *client.FaceDetect
	faceboxAddress string
}

// NewEmojify creates a new Emojify instance
func NewEmojify(fetcher Fetcher, address, imagePath string) Emojify {
	emojis := loadEmojis(imagePath)

	fd := client.NewClient(fmt.Sprintf("http://%s", address))

	return &Impl{
		emojis:  emojis,
		fetcher: fetcher,
		fd:      fd,
	}
}

// GetFaces finds the faces in an image
func (e *Impl) GetFaces(r io.ReadSeeker) ([]image.Rectangle, error) {
	_, err := r.Seek(0, os.SEEK_SET)
	if err != nil {
		return nil, err
	}

	// ok to continue
	f, err := e.fd.DetectFaces(r)
	if err != nil {
		return nil, err
	}

	return f.Faces, nil
}

// Emojimise detects faces in an image and replaces them with emoji
func (e *Impl) Emojimise(src image.Image, faces []image.Rectangle) (image.Image, error) {
	dstImage := image.NewRGBA(src.Bounds())
	draw.Draw(dstImage, src.Bounds(), src, image.ZP, draw.Src)

	for _, face := range faces {
		m := resize.Resize(uint(face.Size().Y), uint(face.Size().X), e.randomEmoji(), resize.Lanczos3)
		sp2 := image.Point{face.Min.X, face.Min.Y}
		r2 := image.Rectangle{sp2, sp2.Add(m.Bounds().Size())}

		draw.Draw(
			dstImage,
			r2,
			m,
			image.ZP,
			draw.Over)
	}
	return dstImage, nil
}

// Health returns health info about facebox
func (e *Impl) Health() (int, error) {
	return http.StatusOK, nil
}

func (e *Impl) randomEmoji() image.Image {
	return e.emojis[rand.Intn(len(e.emojis))]
}

func loadEmojis(path string) []image.Image {
	images := make([]image.Image, 0)
	root := path

	filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			reader, err := os.Open(root + f.Name())
			if err != nil {
				return err
			}
			defer reader.Close()

			i, _, err := image.Decode(reader)
			if err == nil {
				images = append(images, i)
			}
		}

		return nil
	})

	return images
}
