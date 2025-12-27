package handlers

import "fmt"

type ResponseWriter interface {
	Write(data []byte) error
	SetStatus(code int)
}

type JSONResponse struct {
	statusCode int
}

func (j *JSONResponse) Write(data []byte) error {
	fmt.Printf("Writing JSON %s\n", string(data))
	return nil
}

func (j *JSONResponse) SetStatus(code int) {
	j.statusCode = code
}

type XMLResponse struct {
	statusCode int
}

func (x *XMLResponse) Write(data []byte) error {
	fmt.Printf("Writing JSON %s\n", string(data))
	return nil
}

func (x *XMLResponse) SetStatus(code int) {
	x.statusCode = code
}

func SendResponse(w ResponseWriter, data []byte) {
	w.SetStatus(200)
	w.Write(data)
}
