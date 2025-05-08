package common

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	C     *gin.Context
	Err   error
	Pages interface{}
	Data  interface{}
}

func (r *Response) Send() {
	defer func() {
		if r := recover(); r != nil {
			Log.Error(errors.New(fmt.Sprint(r)))
		}
	}()
	var success bool
	var message string
	if r.Err == nil {
		success = true
		message = "ok"
	} else {
		message = r.Err.Error()
		r.Data = nil
	}
	if r.Pages == nil {
		r.C.JSON(http.StatusOK, gin.H{"success": success, "message": message, "data": r.Data})
	} else {
		r.C.JSON(http.StatusOK, gin.H{"pages": r.Pages, "success": success, "message": message, "data": r.Data})
	}
}
func (r *Response) Unauthorized() {
	defer func() {
		if r := recover(); r != nil {
			Log.Error(errors.New(fmt.Sprint(r)))
		}
	}()
	r.C.JSON(http.StatusUnauthorized, gin.H{
		"success": false, "message": "Unauthorized", "data": map[string]string{}})
}
