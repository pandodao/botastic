package render

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/oxtoacart/bpool"
)

// Response internal error msg as hint
var ResponseErrorMessageAsHint bool

func init() {
	v := os.Getenv("RESPONSE_ERROR_MESSAGE_AS_HINT")
	ResponseErrorMessageAsHint, _ = strconv.ParseBool(v)
}

type wrapResponse struct {
	status int
	header http.Header
	buf    *bytes.Buffer
}

func (w *wrapResponse) Header() http.Header {
	return w.header
}

func (w *wrapResponse) WriteHeader(statusCode int) {
	w.status = statusCode
}

func (w *wrapResponse) Write(data []byte) (int, error) {
	return w.buf.Write(data)
}

func (w *wrapResponse) isJsonContent() bool {
	typ := w.header.Get("Content-Type")
	return strings.HasPrefix(typ, "application/json")
}

type dataResponse struct {
	Timestamp int64           `json:"ts,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

type errorResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Hint string `json:"hint,omitempty"`
}

func WrapResponse(wrapData bool) func(http.Handler) http.Handler {
	bufferPool := bpool.NewSizedBufferPool(64, 16*1024)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			buf := bufferPool.Get()
			defer bufferPool.Put(buf)

			ww := &wrapResponse{
				header: http.Header{},
				buf:    buf,
			}

			next.ServeHTTP(ww, r)

			if ww.status >= 300 {
				var (
					response errorResponse
				)

				response.Code = ww.status
				response.Msg = http.StatusText(ww.status)

				buf.Reset()
				_ = json.NewEncoder(buf).Encode(response)
			} else if wrapData && ww.isJsonContent() {
				r := dataResponse{
					Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
					Data:      buf.Bytes(),
				}
				buf.Reset()
				_ = json.NewEncoder(buf).Encode(r)
			}

			// reset content length
			ww.header.Set("Content-Length", strconv.Itoa(buf.Len()))

			for key := range ww.header {
				w.Header().Set(key, ww.header.Get(key))
			}

			w.WriteHeader(ww.status)
			_, _ = w.Write(buf.Bytes())
		})
	}
}
