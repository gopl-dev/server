package file

import (
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif" // register GIF decoder for image.Decode
	"image/jpeg"
	_ "image/png" // register PNG decoder for image.Decode
	"io"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gopl-dev/server/app"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp" // register WEBP decoder for image.Decode
)

var (
	// ErrPreviewNotSupported ...
	ErrPreviewNotSupported = errors.New("preview not supported")

	// ErrInvalidImageDimensions ...
	ErrInvalidImageDimensions = errors.New("invalid image dimensions")

	// ErrIImageResolutionIsTooLarge ...
	ErrIImageResolutionIsTooLarge = errors.New("image resolution is too large")
)

// ResizableImages is the list of file extensions for which we can generate previews.
var ResizableImages = []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

// CheckImageDimensions validates image resolution.
func CheckImageDimensions(r io.ReadSeeker) error {
	conf := app.Config().Files

	dc, _, err := image.DecodeConfig(r)
	if err != nil {
		return err
	}
	_, _ = r.Seek(0, io.SeekStart)

	if dc.Width <= 0 || dc.Height <= 0 {
		return ErrInvalidImageDimensions
	}

	if dc.Width > conf.ImageMaxWidth || dc.Height > conf.ImageMaxHeight {
		return fmt.Errorf("%w: max is %dx%d; your is %dx%d", ErrIImageResolutionIsTooLarge, conf.ImageMaxHeight, conf.ImageMaxWidth, dc.Width, dc.Height)
	}

	return nil
}

// CreatePreview generates a preview image for the given source object using default settings
// from config (preview width/height and JPEG quality).
//
// The preview is stored next to the source in a "preview/" subdirectory using the same base name.
// It returns the destination key/path for the generated preview.
func CreatePreview(ctx context.Context, source string) (string, error) {
	conf := app.Config().Files

	if !IsResizableImage(source) {
		return "", ErrPreviewNotSupported
	}

	w, h := conf.PreviewWidth, conf.PreviewHeight

	dir := filepath.Dir(source)
	base := filepath.Base(source)
	previewName := filepath.Join(dir, "preview", base)

	err := CreatePreviewCustom(ctx, source, previewName, w, h, 85) //nolint:mnd

	return previewName, err
}

// CreatePreviewCustom creates an image preview with the provided settings.
//
// Input formats: JPEG, PNG, GIF, WEBP (decoded via image.Decode; supported formats are enabled
// by the blank imports above).
// Output format: JPEG.
//
// The image is resized to fit within maxW x maxH while preserving aspect ratio.
// If maxW/maxH are invalid, they are clamped to configured maximums.
func CreatePreviewCustom(ctx context.Context, srcKey, dstKey string, maxW, maxH int, quality int) error {
	conf := app.Config().Files

	if maxW <= 0 || maxW > conf.PreviewWidth {
		maxW = conf.PreviewWidth
	}
	if maxH <= 0 || maxH > conf.PreviewHeight {
		maxH = conf.PreviewHeight
	}

	rc, _, err := Open(ctx, srcKey)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := rc.Close()
		if closeErr != nil {
			fmt.Println("CLOSE FILE ERROR:", closeErr)
		}
	}()

	img, _, err := image.Decode(rc)
	if err != nil {
		return err
	}

	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	newW, newH := fit(w, h, maxW, maxH)

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, b, draw.Over, nil)

	pr, pw := io.Pipe()

	encodeErrCh := make(chan error, 1)
	go func() {
		errEncode := jpeg.Encode(pw, dst, &jpeg.Options{Quality: quality})
		_ = pw.CloseWithError(errEncode)
		encodeErrCh <- errEncode
	}()

	_, err = Store(ctx, pr, dstKey)
	if err != nil {
		_ = pr.Close()
		<-encodeErrCh
		return err
	}

	err = <-encodeErrCh
	if err != nil {
		return err
	}

	return nil
}

// fit computes a new size (nw, nh) so that an image of size w x h fits within maxW x maxH
// while preserving aspect ratio. It never returns dimensions less than 1x1.
func fit(w, h, maxW, maxH int) (int, int) {
	if w <= maxW && h <= maxH {
		return w, h
	}

	rw := float64(maxW) / float64(w)
	rh := float64(maxH) / float64(h)

	scale := rw
	if rh < rw {
		scale = rh
	}

	nw := int(float64(w)*scale + 0.5) //nolint:mnd
	nh := int(float64(h)*scale + 0.5) //nolint:mnd

	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}

	return nw, nh
}

// IsResizableImage reports whether a preview can be generated for the given filename
// by checking its extension against ResizableImages.
func IsResizableImage(filename string) bool {
	return slices.Contains(
		ResizableImages,
		strings.ToLower(filepath.Ext(filename)),
	)
}
