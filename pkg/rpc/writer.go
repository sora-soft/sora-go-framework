package rpc

type ResponseWriter struct {
	headers map[string]string
}

func newResponseWriter() *ResponseWriter {
	return &ResponseWriter{
		headers: make(map[string]string),
	}
}

func (w *ResponseWriter) GetHeader(key string) string {
	return w.headers[key]
}

func (w *ResponseWriter) GetHeaders() map[string]string {
	return w.headers
}

func (w *ResponseWriter) SetHeader(key string, value string) {
	w.headers[key] = value
}

func (w *ResponseWriter) SetHeaders(headers map[string]string) {
	for k, v := range headers {
		w.headers[k] = v
	}
}

func (w *ResponseWriter) Headers() map[string]string {
	return w.headers
}
