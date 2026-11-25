package handler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"change-type-image.com/internal/utils"
)

func CompressImage(w http.ResponseWriter, r *http.Request) {

	// Batasi ukuran request maksimal 32 MB.
	// Jika client mengirim lebih besar dari ini, request akan ditolak.
	const MaxUploadSize = 32 << 20
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)

	// Membuat MultipartReader.
	// MultipartReader membaca request multipart secara streaming
	// sehingga kita tidak perlu memuat seluruh file ke memory.
	reader, err := r.MultipartReader()
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Bad Request", "Cannot parse form")
		log.Println(err.Error())
		return
	}

	// State yang dikumpulkan dari multipart fields
	var (
		fileData    []byte
		contentType string
		fileName    string
		quality     int
	)

	for {
		part, err := reader.NextPart()

		if err == io.EOF {
			break
		}

		if err != nil {
			utils.ErrorResponse(w, http.StatusBadRequest, "Bad Request", err.Error())
			log.Println(err.Error())
			return
		}

		switch part.FormName() {

		case "file":
			// Pastikan user benar benar memasukkan file
			if part.FileName() == "" {
				utils.ErrorResponse(w, http.StatusBadRequest, "Bad Request", "File is required !")
				return
			}

			fileName = part.FileName()

			// Buffer untuk membaca signature file
			header := make([]byte, 512)

			n, err := io.ReadFull(part, header)
			// io.ErrUnexpectedEOF artinya file kurang dari 512 byte,
			// tapi tetap bisa jadi file yang valid.
			if err != nil && err != io.ErrUnexpectedEOF {
				utils.ErrorResponse(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
				fmt.Println(err.Error())
				return
			}

			contentType = http.DetectContentType(header[:n])
			if len(contentType) < 6 || contentType[:6] != "image/" {
				utils.ErrorResponse(w, http.StatusBadRequest, "Bad Request", "Only image files can be compressed")
				fmt.Println("Only image file can be compressed for now !")
				return
			}

			// Gabungkan 512 byte header tadi dengan sisa stream,
			// lalu baca seluruh file ke memory.
			fullReader := io.MultiReader(bytes.NewReader(header[:n]), part)
			fileData, err = io.ReadAll(fullReader)
			if err != nil {
				utils.ErrorResponse(w, http.StatusBadRequest, "Bad Request !", err.Error())

				return
			}

		case "quality":

			// READ INPUT FOR quality
			raw, err := io.ReadAll(part)
			if err != nil {
				utils.ErrorResponse(w, http.StatusBadRequest, "Bad Request", err.Error())
				return
			}

			quality, err = strconv.Atoi(strings.TrimSpace(string(raw)))
			if err != nil || quality < 1 || quality > 100 {
				utils.ErrorResponse(w, http.StatusBadRequest, "Bad Request", "quality must be a number between 1 and 100")
				return
			}

		}

	}

	// Validasi file data tidak boleh 0
	if len(fileData) == 0 {
		utils.ErrorResponse(w, http.StatusBadRequest, "Bad Request", "File is required")
		return
	}

	// Validasi quality wajib diisi
	if quality == 0 {
		utils.ErrorResponse(w, http.StatusBadRequest, "Bad Request", "quality is required")
		return
	}

	// Kompresi Image
	result, err := utils.CompressImage(fileData, contentType, quality)
	if err != nil {
		log.Println(err.Error())

		if errors.Is(err, utils.ErrUnsupportedImage) || errors.Is(err, utils.ErrInvalidImage) {
			utils.ErrorResponse(w, http.StatusUnsupportedMediaType, "Unsupported Media Type", err.Error())
			return
		}

		utils.ErrorResponse(w, http.StatusInternalServerError, "Internal Server Error", "Failed to compress image")
		return
	}

	// foto.jpg -> foto-compressed.jpg
	ext := filepath.Ext(fileName)
	outName := strings.TrimSuffix(fileName, ext) + "-compressed" + ext

	// Tanda kutip dan backslash akan merusak header Content-Disposition
	outName = strings.NewReplacer(`"`, "_", `\`, "_").Replace(outName)

	w.Header().Set("Content-Type", result.ContentType)
	w.Header().Set("Content-Disposition", `attachment; filename="`+outName+`"`)
	w.Header().Set("Content-Length", strconv.Itoa(len(result.Data)))

	if _, err := w.Write(result.Data); err != nil {
		log.Println(err.Error())
	}
}
