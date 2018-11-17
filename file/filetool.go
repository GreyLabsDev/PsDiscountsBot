package file

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

var fileCreateOpenError = "File creating/opening error"

type DownloadCallback func()

func OpenFile(fileName string) *os.File {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println(fileCreateOpenError)
		panic(err)
	}
	return file
}

func CreateFile(fileName string) *os.File {
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println(fileCreateOpenError)
		panic(err)
	}
	return file
}

func CloseFile(file *os.File) {
	file.Close()
}

func AppendToFile(file *os.File, stringToAppend string) {
	file.WriteString(stringToAppend)
}

func AppendToFileAndClose(file *os.File, stringToAppend string) {
	file.WriteString(stringToAppend)
	file.Close()
}

func ReadImage(imagePath string) (outImage image.Image) {
	imageFile := OpenFile(imagePath)
	defer CloseFile(imageFile)

	var decodeError error
	outImage, decodeError = png.Decode(imageFile)
	if decodeError != nil {
		return
	}

	return outImage
}

func SaveImage(imagePath string, targetImage image.Image) {
	imageFile := CreateFile(imagePath)
	defer CloseFile(imageFile)

	png.Encode(imageFile, targetImage)
}

func DownloadImage(imageURL string, imageName string, callback DownloadCallback) {
	imageFile := CreateFile(imageName)

	resp, err := http.Get(imageURL)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	size, err2 := io.Copy(imageFile, resp.Body)
	if err2 != nil {
		log.Fatal(err2)
	}

	fmt.Println("Image loaded, size = " + strconv.FormatInt(size, 10))

	CloseFile(imageFile)
	callback()
}
