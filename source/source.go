package source

import (
	"context"
	"io"
	"net/http"
)

type Source interface {
	GetFile(ctx context.Context, path string) (data string, err error)
}

type Func func(ctx context.Context, path string) (data string, err error)

func (s Func) GetFile(ctx context.Context, path string) (data string, err error) {
	return s(ctx, path)
}

func NewHTTPSource() Source {
	return Func(func(ctx context.Context, path string) (data string, err error) {
		r, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
		if err != nil {
			return "", err
		}

		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		return string(bytes), nil
	})
}
