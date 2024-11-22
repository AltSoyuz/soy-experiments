package forms

import (
	"fmt"
	"net/http"
)

type CodeForm struct {
	Code string `form:"code"`
}

func CodeFrom(r *http.Request) (CodeForm, error) {
	err := r.ParseForm()
	if err != nil {
		return CodeForm{}, err
	}
	code := r.FormValue("code")

	if code == "" {
		return CodeForm{}, fmt.Errorf("no code provided")
	}

	return CodeForm{Code: code}, nil
}
