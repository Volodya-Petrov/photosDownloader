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
	realId, err := getId(vkId, token)
	if err != nil {
		return err
	}

	response, err := getPhotos(realId, token, 0)
	if err != nil {
		return err
	}

	err = os.Mkdir("result", 0777)
	if err != nil && !os.IsExist(err) {
		return err
	}

	photosCount := response.Count
	countWorkers := runtime.NumCPU()
	w8 := sync.WaitGroup{}
	w8.Add(countWorkers)
	ch := make(chan item, len(response.Items)*3)

	for i := 0; i < countWorkers; i++ {
		go worker(ch, &w8, i)
	}

	offset := 0
	for offset < photosCount {
		response, err := getPhotos(realId, token, offset)
		if err != nil {
			fmt.Printf("err downloading photos: %v\n", err)
			offset = photosCount
			continue
		}
		for _, item := range response.Items {
			ch <- item
		}
		offset += len(response.Items)
	}

	close(ch)
	w8.Wait()
	return nil
}

func worker(ch chan item, w8 *sync.WaitGroup, workerId int) {
	for photos := range ch {
		fmt.Printf("Worker %v takes task\n", workerId)
		idInStr := strconv.FormatUint(photos.Id, 10)
		err := os.Mkdir("result/"+idInStr, 0777)
		if err != nil {
			fmt.Printf("Worker %v err: %v\n", workerId, err)
			if !os.IsExist(err) {
				continue
			}
		}
		for _, photo := range photos.Photos {
			file, err := os.Create("result/" + idInStr + "/" + photo.Size)
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

func getPhotos(id uint64, token string, offset int) (response, error) {
	url := fmt.Sprintf(fmt.Sprintf("https://api.vk.com/method/photos.getAll?"+
		"owner_id=%v&offset=%v&access_token=%v&v=5.131", id, offset, token))
	resp, err := http.Get(url)
	if err != nil {
		return response{}, err
	}
	photosJson := photosJson{}
	err = json.NewDecoder(resp.Body).Decode(&photosJson)
	if err != nil {
		return response{}, err
	}
	return photosJson.Response, nil
}

func getId(vkId, token string) (uint64, error) {
	url := fmt.Sprintf("https://api.vk.com/method/users.get?"+
		"user_ids=%v&access_token=%v&v=5.131", vkId, token)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	usersJson := usersJson{}
	err = json.NewDecoder(resp.Body).Decode(&usersJson)
	if err != nil {
		return 0, err
	}
	return usersJson.Users[0].Id, nil
}
