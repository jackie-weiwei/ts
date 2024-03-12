package ts

import (
	"encoding/base64"
	"os"

	"github.com/gofrs/uuid"
)

func WriteImageFile(path string, base64_image_content string) (bool, string) {

	fileType := "png"
	id, _ := uuid.NewV4()
	ids := id.String()

	var fileName string = path + "/" + ids + "." + fileType
	byte, _ := base64.StdEncoding.DecodeString(base64_image_content)

	err := os.WriteFile(fileName, byte, 0666)

	return err == nil, ids + "." + fileType
}
