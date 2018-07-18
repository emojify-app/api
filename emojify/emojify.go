package emojify

import (
	"image"
	"image/draw"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/machinebox/sdk-go/facebox"
	"github.com/nfnt/resize"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Emojify interface {
	Emojimise(image.Image, []facebox.Face) (image.Image, error)
	GetFaces(f io.ReadSeeker) ([]facebox.Face, error)
}

type EmojifyImpl struct {
	emojis  []image.Image
	fetcher Fetcher
}

func NewEmojify(fetcher Fetcher, imagePath string) Emojify {
	emojis := loadEmojis(imagePath)
	return &EmojifyImpl{
		emojis:  emojis,
		fetcher: fetcher,
	}
}

func (e *EmojifyImpl) Emojimise(src image.Image, faces []facebox.Face) (image.Image, error) {
	dstImage := image.NewRGBA(src.Bounds())
	draw.Draw(dstImage, src.Bounds(), src, image.ZP, draw.Src)

	for _, face := range faces {
		m := resize.Resize(uint(face.Rect.Height), uint(face.Rect.Width), e.randomEmoji(), resize.Lanczos3)
		sp2 := image.Point{face.Rect.Left, face.Rect.Top}
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

func (e *EmojifyImpl) GetFaces(r io.ReadSeeker) ([]facebox.Face, error) {
	_, err := r.Seek(0, os.SEEK_SET)
	if err != nil {
		return nil, err
	}

	fb := facebox.New(os.Getenv("FACEBOX"))
	fb.HTTPClient.Timeout = 30000 * time.Millisecond

	return fb.Check(r)
}

func (e *EmojifyImpl) randomEmoji() image.Image {
	return e.emojis[rand.Intn(len(e.emojis))]
}

func loadEmojis(path string) []image.Image {
	images := make([]image.Image, 0)
	root := path

	filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			log.Println("Processing: ", root+f.Name())
			reader, err := os.Open(root + f.Name())
			if err != nil {
				log.Println("Unable to open image")
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
