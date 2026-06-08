package handler

import (
	"fmt"
	"image"
	"image/jpeg"
	_ "image/jpeg" // Register JPEG decoder
	"image/png"
	_ "image/png" // Register PNG decoder
	"net/http"
	"path/filepath"

	"change-type-image.com/internal/utils"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
)

func ConvertImages(w http.ResponseWriter, r *http.Request) {

	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Bad Request", "Cannot parse form")
		fmt.Println(err.Error())
		return
	}

	typeName := r.FormValue("type")
	typeValue, err := utils.ValidateImageType(typeName)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Bad Request", err.Error())
		fmt.Println(err.Error())
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Bad Request", "File not found")
		fmt.Println(err.Error())
		return
	}
	defer file.Close()

	// ! READ 512 bytes for header checking
	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Internal server error", "Failed to read input file")
		fmt.Println(err.Error())
		return
	}

	// ! Reset Pointer
	file.Seek(0, 0)

	// ! Detect content type only accept image
	contentType := http.DetectContentType(buff)

	if contentType != "image/png" && contentType != "image/jpeg" && contentType != "image/webp" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Bad Request", "Inputted image is not valid (jpg,png,jpeg,webp)")
		fmt.Println("Error inputing image")
		return
	}

	// ! Decode image into generic image file
	img, _, err := image.Decode(file)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Something went wrong !", "Failed to decode image !")
		fmt.Println(err.Error())
		return
	}

	ext := filepath.Ext(header.Filename)
	name := header.Filename[:len(header.Filename)-len(ext)]

	// ! name_image_converted.jpg
	fileName := name + "_converted" + "." + typeValue

	// ! CONVERTING IMAGE
	switch typeValue {

	case "png":

		w.Header().Set(
			"Content-Disposition",
			`attachment; filename="`+fileName+`"`,
		)
		w.Header().Set("Content-Type", "image/png")

		err := png.Encode(w, img)
		if err != nil {
			http.Error(w, err.Error(), 500)
			fmt.Println(err.Error())
		}

	case "jpg", "jpeg":
		w.Header().Set(
			"Content-Disposition",
			`attachment; filename="`+fileName+`"`,
		)
		w.Header().Set("Content-Type", "image/jpeg")

		err := jpeg.Encode(w, img, &jpeg.Options{Quality: 80})
		if err != nil {
			http.Error(w, err.Error(), 500)
			fmt.Println(err.Error())

		}
	case "webp":
		w.Header().Set(
			"Content-Disposition",
			`attachment; filename="`+fileName+`"`,
		)
		w.Header().Set("Content-Type", "image/webp")

		options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 75)
		if err != nil {
			http.Error(w, err.Error(), 500)
			fmt.Println(err.Error())
		}

		if err := webp.Encode(w, img, options); err != nil {
			http.Error(w, err.Error(), 500)
			fmt.Println(err.Error())
		}

	default:
		utils.ErrorResponse(w, http.StatusBadRequest, "Failed to convert to"+typeName, "Failed converting image")
	}

}
