package photosave

import (
	"mime/multipart"
	"net/http"
)

type SaveObj struct {
	R                 *http.Request
	Reader            multipart.File
	ValueName         string
	FilePath          string
	FileName          string
	FileNameRnd       bool
	WatermarkPath     string
	WatermarkX        int
	WatermarkXFromMax bool
	WatermarkY        int
	WatermarkYFromMax bool
}
