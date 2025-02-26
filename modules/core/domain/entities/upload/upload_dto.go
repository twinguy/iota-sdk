package upload

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"path/filepath"

	"github.com/gabriel-vasile/mimetype"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/iota-uz/iota-sdk/pkg/configuration"
	"github.com/iota-uz/iota-sdk/pkg/constants"
)

type CreateDTO struct {
	File io.ReadSeeker `validate:"required"`
	Name string        `validate:"required"`
	Size int           `validate:"required"`
	Type string
}

func (d *CreateDTO) Ok(l ut.Translator) (map[string]string, bool) {
	errorMessages := map[string]string{}
	errs := constants.Validate.Struct(d)
	if errs == nil {
		return errorMessages, true
	}

	for _, err := range errs.(validator.ValidationErrors) {
		errorMessages[err.Field()] = err.Translate(l)
	}
	return errorMessages, len(errorMessages) == 0
}

func (d *CreateDTO) ToEntity() (Upload, []byte, error) {
	conf := configuration.Use()
	bytes, err := io.ReadAll(d.File)
	if err != nil {
		return nil, nil, err
	}
	mdsum := md5.Sum(bytes)
	hash := hex.EncodeToString(mdsum[:])
	ext := filepath.Ext(d.Name)
	return New(
		hash,
		filepath.Join(conf.UploadsPath, hash+ext),
		d.Size,
		mimetype.Detect(bytes),
	), bytes, nil
}
