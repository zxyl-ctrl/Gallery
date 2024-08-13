package models

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
)

var (
	ErrNotFound   = errors.New("models: resource could not be found")
	ErrEmailTaken = errors.New("models: email address is already in use")
)

type FileError struct {
	Issue string
}

func (fe FileError) Error() string {
	return fmt.Sprintf("invalid file: %v", fe.Issue)
}

// 由于seek在读取前几位后(魔数)重置文件，以验证图像类型是否有效，这限制的图像类型
// 为了支持在DropBox中对应的所有图像类型，应当改进这一点
// 一种解决
// func checkContentType(r io.ReadSeeker, allowTypes []string) error {
// 	testBytes := make([]byte, 512)
// 	_, err := r.Read(testBytes)
// 	if err != nil {
// 		return fmt.Errorf("checking content type: %w", err)
// 	}

// 	_, err = r.Seek(0, 0)
// 	if err != nil {
// 		return fmt.Errorf("checkingg content type: %w", err)
// 	}

// 	contentType := http.DetectContentType(testBytes)
// 	for _, t := range allowTypes {
// 		if contentType == t {
// 			return nil
// 		}
// 	}

// 	return FileError{
// 		Issue: fmt.Sprintf("invalid content type: %v", contentType),
// 	}
// }

// 仅仅返回读取的字节
func checkContentType(r io.Reader, allowedTypes []string) ([]byte, error) {
	testBytes := make([]byte, 512)
	n, err := r.Read(testBytes)
	if err != nil {
		return nil, fmt.Errorf("checking content type: %w", err)
	}
	contentType := http.DetectContentType(testBytes)
	for _, t := range allowedTypes {
		if contentType == t {
			return testBytes[:n], nil
		}
	}
	return nil, FileError{
		Issue: fmt.Sprintf("invalid content type: %v", contentType),
	}
}

func checkExtension(filename string, allowedExtensions []string) error {
	if !hasExtension(filename, allowedExtensions) {
		return FileError{
			Issue: fmt.Sprintf("invalid extension: %v", filepath.Ext(filename)),
		}
	}
	return nil
}
