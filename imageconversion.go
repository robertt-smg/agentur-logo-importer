package main

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/gif"
	jpeg "image/jpeg"

	"image/png"
	"io"
	"os"

	m "github.com/h2non/filetype/matchers"
	"github.com/mdouchement/hdr"
	_ "github.com/mdouchement/hdr/codec/rgbe"
	"github.com/mdouchement/hdr/tmo"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

const (
	GIF     = "GIF"
	PNG     = "PNG"
	BMP     = "BMP"
	JPEG    = "JPEG"
	TIFF    = "TIFF"
	CR2     = "CR2"
	INVALID = ""
)

var FAILED_TO_OPEN = errors.New("FAILED_TO_OPEN")
var UNSUPPORTED_FILE_FORMAT = errors.New("UNSUPPORTED_FILE_FORMAT")
var FAILED_TO_CONVERT = errors.New("FAILED_TO_CONVERT")

// StreamToByte StreamToByte
func StreamToByte(stream io.Reader) []byte {
	var buf bytes.Buffer
	buf.ReadFrom(stream)
	return buf.Bytes()
}

// IsValidFormat IsValidFormat
func IsValidFormat(image []byte) (string, error) {
	if len(image) == 0 {
		return INVALID, fmt.Errorf("Empty file")
	}

	if m.Gif(image) {
		return GIF, nil
	}
	if m.Png(image) {
		return PNG, nil
	}
	if m.Bmp(image) {
		return BMP, nil
	}
	if m.Jpeg(image) {
		return JPEG, nil
	}
	if m.CR2(image) {
		return CR2, nil
	}
	if m.Tiff(image) {
		return TIFF, nil
	}
	header := image[:4]
	return INVALID, fmt.Errorf("Invalid format (%s %s)", hex.Dump(header), string(header))
}

// ReadImageFile ReadImageFile
func ReadImageFile(path string) (*os.File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func pngToJpeg(sourcePath, destinationPath string) error {
	file, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	src, err := png.Decode(file)
	if err != nil {
		return err
	}
	outfile, _ := os.Create(destinationPath)
	defer outfile.Close()
	jpeg.Encode(outfile, src, &jpeg.Options{
		Quality: 80,
	})
	return nil
}

func gifToJpeg(sourcePath, destinationPath string) error {
	file, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	src, err := gif.Decode(file)
	if err != nil {
		return err
	}
	outfile, _ := os.Create(destinationPath)
	defer outfile.Close()
	jpeg.Encode(outfile, src, &jpeg.Options{
		Quality: 80,
	})
	return nil
}

func bmpToJpeg(sourcePath, destinationPath string) error {
	file, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	src, err := bmp.Decode(file)
	if err != nil {
		return err
	}
	outfile, _ := os.Create(destinationPath)
	defer outfile.Close()
	jpeg.Encode(outfile, src, &jpeg.Options{
		Quality: 80,
	})
	return nil
}
func cr2ToJpeg(sourcePath, destinationPath string) error {
	file, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	m, _, err := image.Decode(file)

	if hdrm, ok := m.(hdr.Image); ok {

		t := tmo.NewLinear(hdrm)
		// t := tmo.NewLogarithmic(hdrm)
		// t := tmo.NewDefaultDrago03(hdrm)
		// t := tmo.NewDefaultDurand(hdrm)
		// t := tmo.NewDefaultCustomReinhard05(hdrm)
		// t := tmo.NewDefaultReinhard05(hdrm)
		// t := tmo.NewDefaultICam06(hdrm)
		m = t.Perform()
	}

	fo, err := os.Create(destinationPath)
	if err != nil {
		return err
	}

	jpeg.Encode(fo, m, &jpeg.Options{
		Quality: 80,
	})

	return nil
}
func tiffToJpeg(sourcePath, destinationPath string) error {
	file, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	src, err := tiff.Decode(file)
	if err != nil {
		return err
	}
	outfile, _ := os.Create(destinationPath)
	defer outfile.Close()
	jpeg.Encode(outfile, src, &jpeg.Options{
		Quality: 80,
	})
	return nil
}

func jpegToJpeg(sourcePath, destinationPath string) error {
	file, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	src, err := jpeg.Decode(file)
	if err != nil {
		return err
	}
	outfile, _ := os.Create(destinationPath)
	defer outfile.Close()
	jpeg.Encode(outfile, src, &jpeg.Options{
		Quality: 80,
	})
	return nil
}

// ConvertImageToJPEG ConvertImageToJPEG
func ConvertImageToJPEG(sourcePath, destinationPath string) (formatOk bool, err error) {
	infile, err := ReadImageFile(sourcePath)
	if err != nil {
		return false, FAILED_TO_OPEN
	}
	defer infile.Close()

	binaryFile := StreamToByte(infile)
	fileFormat, err := IsValidFormat(binaryFile)
	if err != nil {
		return false, err
	}
	switch fileFormat {
	case PNG:
		err := pngToJpeg(sourcePath, destinationPath)
		if err != nil {
			return false, err
		}
		formatOk = true
	case GIF:
		err := gifToJpeg(sourcePath, destinationPath)
		if err != nil {
			return false, err
		}
		formatOk = true
	case BMP:
		err := bmpToJpeg(sourcePath, destinationPath)
		if err != nil {
			return false, err
		}
		formatOk = true
	case CR2:
		err := cr2ToJpeg(sourcePath, destinationPath)
		if err != nil {
			return false, err
		}
		formatOk = true
	case TIFF:
		err := tiffToJpeg(sourcePath, destinationPath)
		if err != nil {
			return false, err
		}
		formatOk = true
	case JPEG:
		err := jpegToJpeg(sourcePath, destinationPath)
		if err != nil {
			return false, err
		}
		formatOk = true

	default:
		return false, UNSUPPORTED_FILE_FORMAT
	}
	return formatOk, nil

}

/*func main1() {
    if len(os.Args) < 3 {
        fmt.Println("Insufficient number of params 1.Infile_path 2.outfile_path")
        return
    }
    sourcePath:= os.Args[1]
    destinationPath:= os.Args[2]
    infile, err:= readImageFile(sourcePath)
    if err != nil {
        fmt.Println("failed to read file")
        return
    }
    defer infile.Close()
    b:= StreamToByte(infile)
    fileFormat, _:= IsValidFormat(b)
    fmt.Println(fileFormat)
    err = convertImageToJPEG(sourcePath, destinationPath)
    if err != nil {
        fmt.Println("Failed to convert image to jpeg")
    }
}*/
