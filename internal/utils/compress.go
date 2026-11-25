package utils

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
)

// Kesalahan dari sisi input user (bukan kesalahan server).
var (
	ErrUnsupportedImage = errors.New("unsupported image type (jpg, jpeg, png, webp)")
	ErrInvalidImage     = errors.New("image is corrupt or cannot be read")
)

type CompressResult struct {
	Data        []byte
	ContentType string
}

// CompressImage mengompres gambar tanpa mengubah formatnya.
//
// quality (1 - 100) dipakai langsung oleh encoder JPEG dan WebP.
// Untuk PNG, quality menentukan jumlah warna (lihat pngColors);
// quality 100 berarti lossless.
func CompressImage(imgData []byte, contentType string, quality int) (*CompressResult, error) {
	switch contentType {
	case "image/jpeg", "image/png", "image/webp":
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedImage, contentType)
	}

	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidImage, err)
	}

	var buf bytes.Buffer

	switch contentType {
	case "image/jpeg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})

	case "image/png":
		enc := png.Encoder{CompressionLevel: png.BestCompression}
		if quality < 100 {
			img = quantizePNG(img, pngColors(quality))
		}
		err = enc.Encode(&buf, img)

	case "image/webp":
		var options *encoder.Options
		options, err = encoder.NewLossyEncoderOptions(encoder.PresetDefault, float32(quality))
		if err == nil {
			err = webp.Encode(&buf, img, options)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode %s: %w", contentType, err)
	}

	// Jangan pernah mengembalikan file yang lebih besar dari aslinya.
	data := buf.Bytes()
	if len(data) >= len(imgData) {
		data = imgData
	}

	return &CompressResult{Data: data, ContentType: contentType}, nil
}
