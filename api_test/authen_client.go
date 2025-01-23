package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const baseURL = "http://localhost:8080"

type Client struct {
	httpClient *http.Client
	idToken    string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{},
	}
}

func (c *Client) SignUp(email, password string) error {
	payload := map[string]string{
		"email":    email,
		"password": password,
	}
	return c.sendRequest("POST", "/signup", payload, nil)
}

func (c *Client) SignIn(email, password string) error {
	payload := map[string]string{
		"email":    email,
		"password": password,
	}
	var response map[string]interface{}
	err := c.sendRequest("POST", "/signin", payload, &response)
	if err != nil {
		return err
	}
	c.idToken = response["idToken"].(string)
	return nil
}

func (c *Client) ChangePassword(newPassword string) error {
	payload := map[string]string{
		"idToken":  c.idToken,
		"password": newPassword,
	}
	return c.sendRequest("POST", "/changepassword", payload, nil)
}

func (c *Client) GetCurrentUser() error {
	return c.sendRequest("GET", "/getCurrentUser", nil, nil)
}

func (c *Client) SignOut(uid string) error {
	payload := map[string]string{
		"uid": uid,
	}
	return c.sendRequest("POST", "/signout", payload, nil)
}

func (c *Client) DeleteAccount() error {
	return c.sendRequest("DELETE", "/deleteAccount", nil, nil)
}

func (c *Client) sendRequest(method, endpoint string, payload interface{}, response interface{}) error {
	var body []byte
	var err error

	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest(method, baseURL+endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.idToken != "" && (endpoint != "/signin" && endpoint != "/signup") {
		req.Header.Set("Authorization", "Bearer "+c.idToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	fmt.Printf("Response for %s %s:\n%s\n\n", method, endpoint, string(respBody))

	if response != nil {
		return json.Unmarshal(respBody, response)
	}

	return nil
}

func main() {
	client := NewClient()

	// Sign up
	err := client.SignUp("test@example.com", "testpassword123")
	if err != nil {
		fmt.Println("SignUp error:", err)
		return
	}

	// Sign in
	err = client.SignIn("testtest@gmail.com", "Namning1678")
	if err != nil {
		fmt.Println("SignIn error:", err)
		return
	}

	// Change password
	err = client.ChangePassword("newpassword123")
	if err != nil {
		fmt.Println("ChangePassword error:", err)
		return
	}

	// Get current user
	err = client.GetCurrentUser()
	if err != nil {
		fmt.Println("GetCurrentUser error:", err)
		return
	}

	// Sign out (you need to provide a valid UID here)
	err = client.SignOut("some-valid-uid")
	if err != nil {
		fmt.Println("SignOut error:", err)
		return
	}

	// Delete account
	err = client.DeleteAccount()
	if err != nil {
		fmt.Println("DeleteAccount error:", err)
		return
	}
}
