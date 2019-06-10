///////////////////////////////////////////////////
//
// Copyright 2019 Broadcom Inc.
//
///////////////////////////////////////////////////

package server

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"

	"gopkg.in/go-playground/validator.v9"
)

func isSkipValidation(t reflect.Type) bool {
	//log.Printf("IsSkipValidation type: %v\n", t)
	if t == reflect.TypeOf([]int32{}) {
		return true
	} else {
		return false
	}
}

// RequestValidate performas payload validation of request body.
func RequestValidate(r *http.Request, v interface{}) error {
	contentType := r.Header.Get("Content-Type")
	if contentType == "application/json" {
		return validateRequestJSON(r, v)
	}

	log.Printf("Skipping payload validation for content-type '%s'", contentType)
	return nil
}

// validateRequestJSON performs payload validation for JSON data
func validateRequestJSON(r *http.Request, v interface{}) error {
	jsn, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error while reading request")
		return err
	}

	err = json.Unmarshal(jsn, v)
	if err != nil {
		log.Printf("decoding error %s\n", jsn)
		return err
	}

	//log.Printf("Received data: %s\n", jsn)
	//log.Printf("Type is: %T, Value is:%v\n", v, v)
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
	}

	if !isSkipValidation(val.Type()) {
		log.Println("Going to validate request")
		validate := validator.New()
		if val.Kind() == reflect.Slice {
			//log.Println("Validate using Var")
			err = validate.Var(v, "dive")
		} else {
			//log.Println("Validate using Struct")
			err = validate.Struct(v)
		}
		if err != nil {
			log.Printf("validation failed: %s\n", err.Error())
			return err
		}
	} else {
		log.Printf("Skipping payload validation for dataType %v", val.Type())
	}

	// Get sanitized json by marshalling validated body. Removes
	// extra fields if any..
	newBody, err := json.Marshal(v)
	if err != nil {
		log.Printf("Failed to marshall; %v", err)
		return err
	}

	// Put sanitized body back into http request object
	r.Body = ioutil.NopCloser(bytes.NewBuffer(newBody))

	return nil
}
