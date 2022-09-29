package downloader

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
)

type usersJson struct {
	Users []user `json:"response"`
}

type user struct {
	Id              uint64 `json:"id"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	CanAccessClosed bool   `json:"can_access_closed"`
}
type photosJson struct {
	Response response `json:"response"`
}

type response struct {
	Count int    `json:"count"`
	Items []item `json:"items"`
}

type item struct {
	Id     uint64  `json:"id"`
	Photos []photo `json:"sizes"`
}

type photo struct {
	Url  string `json:"url"`
	Size string `json:"type"`
}

func DownloadPhotos(vkId, token string) error {
	url := fmt.Sprintf("https://api.vk.com/method/users.get?"+
		"user_ids=%v&access_token=%v&v=5.131", vkId, token)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	usersJson := usersJson{}
	err = json.NewDecoder(resp.Body).Decode(&usersJson)
	if err != nil {
		return err
	}

	url = fmt.Sprintf(fmt.Sprintf("https://api.vk.com/method/photos.getAll?"+
		"owner_id=%v&access_token=%v&v=5.131", usersJson.Users[0].Id, token))
	resp, err = http.Get(url)
	photosJson := photosJson{}
	err = json.NewDecoder(resp.Body).Decode(&photosJson)

	countWorkers := runtime.NumCPU()
	w8 := sync.WaitGroup{}
	w8.Add(countWorkers)
	ch := make(chan item, len(photosJson.Response.Items))

	for i := 0; i < countWorkers; i++ {
		go worker(ch, &w8, i)
	}

	for _, item := range photosJson.Response.Items {
		ch <- item
	}

	close(ch)
	w8.Wait()
	return nil
}

func worker(ch chan item, w8 *sync.WaitGroup, workerId int) {
	for photos := range ch {
		fmt.Printf("Worker %v takes task\n", workerId)
		idInStr := strconv.FormatUint(photos.Id, 10)
		err := os.Mkdir(idInStr, 0777)
		if err != nil {
			fmt.Printf("Worker %v err: %v\n", workerId, err)
			continue
		}
		for _, photo := range photos.Photos {
			file, err := os.Create(idInStr + "/" + idInStr + photo.Size)
			if err != nil {
				fmt.Printf("Worker %v err: %v\n", workerId, err)
				continue
			}
			resp, err := http.Get(photo.Url)
			if err != nil {
				fmt.Printf("Worker %v err: %v\n", workerId, err)
				continue
			}
			io.Copy(file, resp.Body)
		}
		fmt.Printf("Worker %v finished task\n", workerId)
	}
	fmt.Printf("Worker %v quit\n", workerId)
	w8.Done()
}
