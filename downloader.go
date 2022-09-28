package photos_downloader

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type usersJson struct {
	users []user `json:"response"`
}

type user struct {
	id              string `json:"id"`
	firstName       string `json:"first_name"`
	lastName        string `json:"last_name"`
	canAccessClosed bool   `json:"can_access_closed"`
}

type photosJson struct {
	count int    `json:"count"`
	items []item `json:"items"`
}

type item struct {
	sizes []photo `json:"sizes"`
}

type photo struct {
	url  string `json:"url"`
	size string `json:"type"`
}

func DownloadPhotos(vkId string) error {
	token, exist := os.LookupEnv("TOKEN")
	if !exist {
		fmt.Println("Token doesnt exist")
		return nil
	}
	url := fmt.Sprintf("https://api.vk.com/method/users.get?"+
		"user_ids=%v&access_token=%v&v=5.131", vkId, token)
	resp, err := http.Get(url)

	if err != nil {
		return err
	}
	var usersJson usersJson
	err = json.NewDecoder(resp.Body).Decode(&usersJson)

	if err != nil {
		return err
	}

	url = fmt.Sprintf(fmt.Sprintf("https://api.vk.com/method/users.get?"+
		"owner_id=%v&access_token=%v&v=5.131", usersJson.users[0].id, token))

	resp, err = http.Get(url)

	var photosJson photosJson
	err = json.NewDecoder(resp.Body).Decode(&photosJson)
	return nil
}
