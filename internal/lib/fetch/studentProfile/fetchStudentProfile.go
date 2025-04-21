package studentProfile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Group struct {
	Name string `json:"name"`
}

type Student struct {
	Token   string `json:"token"`
	Barcode int    `json:"barcode"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Group   Group  `json:"group"`
}

func FetchStudentProfile(token string) (*Student, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.yeunikey.dev/v1/auth/profile", nil)
	if err != nil {
		return nil, err
	}

	authHeader := "Bearer " + token
	req.Header.Set("Authorization", authHeader)
	fmt.Println("Auth header:", authHeader)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\nResponse Body: %s\n", resp.StatusCode, string(bodyBytes))

	if resp.StatusCode == 400 || resp.StatusCode == 401 {
		return nil, errors.New("unauthorized: invalid or expired token")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var res struct {
		Data Student `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &res.Data, nil
}
