package app

import (
	"errors"
	"io"

	"namefreq/internal/config"
)

var ErrPipelineNotImplemented = errors.New("external aggregation pipeline is not implemented yet")

func Run(cfg config.Config, out io.Writer) error {
	_ = cfg
	_ = out

	return ErrPipelineNotImplemented
}
