package studentProfile

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Group struct {
	Name string `json:"name"`
}
type Student struct {
	Token string `json:"token"`

	Barcode string `json:"barcode"`
	Name    string `json:"name"`
	Surname string `json:"surname"`

	Group Group `json:"group"`
}

func FetchStudentProfile(token string) (*Student, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.yeunikey.dev/v1/auth/profile", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 400 || resp.StatusCode == 401 {
		return nil, errors.New("unauthorized")
	}

	var res struct {
		Data Student `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &res.Data, nil
}
