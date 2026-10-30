package xhttp

import (
	"context"
	"io"
	"net/http"
	"time"
)

// Stream requires a route with go-zero's buffered timeout disabled. The context
// bounds storage reads and the write deadline bounds a slow client independently.
func Stream(ctx context.Context, w http.ResponseWriter, body io.ReadCloser, deadline time.Time) error {
	control := http.NewResponseController(w)
	if err := control.SetWriteDeadline(deadline); err != nil {
		return err
	}
	defer control.SetWriteDeadline(time.Time{})
	stop := context.AfterFunc(ctx, func() { _ = body.Close() })
	defer stop()
	buf := make([]byte, 32*1024)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, err := body.Read(buf)
		if n > 0 {
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			if flushErr := control.Flush(); flushErr != nil {
				return flushErr
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}
