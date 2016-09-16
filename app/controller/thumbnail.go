package controller

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
	"golang.org/x/image/webp"
)

func decodeImage(r io.Reader, format string) (img image.Image) {
	switch format {
	case "gif":
		img, _ = gif.Decode(r)
	case "jpeg":
		img, _ = jpeg.Decode(r)
	case "png":
		img, _ = png.Decode(r)
	case "webp":
		img, _ = webp.Decode(r)
	}
	return
}

func Thumbnail(w http.ResponseWriter, r *http.Request) {
	filename := "static/thumbnails/" + mux.Vars(r)["Id"]
	root := http.Dir("static/thumbnails")
	httpFile, err := root.Open(mux.Vars(r)["Id"])
	if err == nil {
		io.Copy(w, httpFile)
		return
	}
	filename = "static/uploads/" + mux.Vars(r)["Id"]
	f, _ := os.Open(filename)
	config, format, _ := image.DecodeConfig(f)
	f.Seek(0, 0)
	img := decodeImage(f, format)
	if img == nil {
		w.WriteHeader(404)
		w.Write([]byte("404 page not found"))
		return
	}
	var dim uint = 300
	var width uint = 0
	var height uint = dim
	if config.Height > config.Width {
		width = dim
		height = 0
	}
	img = resize.Resize(width, height, img, resize.MitchellNetravali)
	img, _ = cutter.Crop(img, cutter.Config{
		Width:  int(dim),
		Height: int(dim),
		Mode:   cutter.Centered,
	})
	var jpegBytes bytes.Buffer
	jpeg.Encode(&jpegBytes, img, &jpeg.Options{
		Quality: 90,
	})
	w.Write(jpegBytes.Bytes())
	f, _ = os.Create("static/thumbnails/" + mux.Vars(r)["Id"])
	f.Write(jpegBytes.Bytes())
}
