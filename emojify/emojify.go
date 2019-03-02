package emojify

import (
	"fmt"
	"image"
	"image/draw"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/machinebox/sdk-go/boxutil"
	"github.com/machinebox/sdk-go/facebox"
	"github.com/nfnt/resize"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// FindFaceResponse is returned from find faces method
type FindFaceResponse struct {
	Faces []facebox.Face
	Error error
}

// Emojify defines an interface for emojify operations
type Emojify interface {
	GetFaces(f io.ReadSeeker, resp chan FindFaceResponse)
	Emojimise(image.Image, []facebox.Face) (image.Image, error)
	Health() (*boxutil.Info, error)
}

// Impl implements the Emojify interface
type Impl struct {
	emojis         []image.Image
	fetcher        Fetcher
	fb             *facebox.Client
	faceboxAddress string
	faceboxWorkers int32
	activeWorkers  int32
	workerTimeout  time.Duration
}

// NewEmojify creates a new Emojify instance
func NewEmojify(fetcher Fetcher, address, imagePath string, workers int32, timeout time.Duration) Emojify {
	emojis := loadEmojis(imagePath)

	fb := facebox.New(fmt.Sprintf("http://%s", address))
	fb.HTTPClient.Timeout = 30 * time.Second

	return &Impl{
		emojis:         emojis,
		fetcher:        fetcher,
		fb:             fb,
		faceboxWorkers: workers,
		activeWorkers:  0,
		workerTimeout:  timeout,
	}
}

// GetFaces finds the faces in an image
func (e *Impl) GetFaces(r io.ReadSeeker, resp chan FindFaceResponse) {
	go func() {
		_, err := r.Seek(0, os.SEEK_SET)
		if err != nil {
			resp <- FindFaceResponse{nil, err}
		}

		// block if we have exhausted workers
		st := time.Now()
		for atomic.LoadInt32(&e.activeWorkers) >= e.faceboxWorkers {
			if time.Now().Sub(st) > e.workerTimeout {
				fmt.Println("timeout", time.Now().Sub(st))
				// fail due to timeout
				resp <- FindFaceResponse{nil, fmt.Errorf("Timeout waiting for worker")}
				break
			}
		}

		// ok to continue
		atomic.AddInt32(&e.activeWorkers, 1)
		faces, err := e.fb.Check(r)
		resp <- FindFaceResponse{faces, err}
		atomic.AddInt32(&e.activeWorkers, -1)
	}()
}

// Emojimise detects faces in an image and replaces them with emoji
func (e *Impl) Emojimise(src image.Image, faces []facebox.Face) (image.Image, error) {
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

// Health returns health info about facebox
func (e *Impl) Health() (*boxutil.Info, error) {
	return e.fb.Info()
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
