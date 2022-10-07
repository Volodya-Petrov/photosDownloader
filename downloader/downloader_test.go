package downloader

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

const pathToTestData = "./testData"
const pathToResult = "./result"

func TestDownloadPhotos(t *testing.T) {
	token, ok := os.LookupEnv("TOKEN")
	if !ok {
		t.Fatalf("Error: access token not found")
	}
	vkId := "lovenaitivixod"

	err := DownloadPhotos(vkId, token)
	if err != nil {
		t.Fatalf("DownoladPhotos returned error: %v", err)
	}
	expectedPhotos, err := ioutil.ReadDir(pathToTestData)
	if err != nil {
		t.Fatalf("Error when reading testData directory: %v", err)
	}
	actualPhotos, err := ioutil.ReadDir(pathToResult)
	if err != nil {
		t.Fatalf("Error when reading result directory: %v", err)
	}

	if len(expectedPhotos) != len(actualPhotos) {
		t.Fatalf("Count of result photos is not equal to count of actual photos, result:%v != actual:%v", len(expectedPhotos), len(actualPhotos))
	}

	for _, expectedPhoto := range expectedPhotos {
		var actualPhoto os.FileInfo
		if !findFile(actualPhotos, expectedPhoto.Name(), &actualPhoto) {
			t.Errorf("Photo %v not found in result directory\n", expectedPhoto.Name())
			continue
		}

		expectedPhotosFiles, err := ioutil.ReadDir(pathToTestData + "/" + expectedPhoto.Name())
		if err != nil {
			t.Errorf("Error when reading %v photo directory in test data", expectedPhoto.Name())
			continue
		}

		actualPhotoFiles, err := ioutil.ReadDir(pathToResult + "/" + actualPhoto.Name())
		if err != nil {
			t.Errorf("Error when reading %v photo directory in result", actualPhoto.Name())
			continue
		}

		for _, expectedFileInfo := range expectedPhotosFiles {
			var actualFileInfo os.FileInfo
			if !findFile(actualPhotoFiles, expectedFileInfo.Name(), &actualFileInfo) {
				t.Errorf("File %v not found in result directory\n", expectedFileInfo.Name())
				continue
			}
			pathToExpectedFile := pathToTestData + "/" + expectedPhoto.Name() + "/" + expectedFileInfo.Name()
			expectedFile, err := ioutil.ReadFile(pathToExpectedFile)
			if err != nil {
				t.Errorf("Couldn't read file %v from testData. Error:%v\n", pathToExpectedFile, err)
				continue
			}

			pathToActualFile := pathToResult + "/" + actualPhoto.Name() + "/" + actualFileInfo.Name()
			actualFile, err := ioutil.ReadFile(pathToActualFile)
			if err != nil {
				t.Errorf("Couldn't read file %v from result data. Error:%v\n", pathToActualFile, err)
				continue
			}

			if !bytes.Equal(actualFile, expectedFile) {
				t.Errorf("File %v not equat to %v\n", pathToExpectedFile, pathToActualFile)
			}
		}
	}
}

func findFile(files []os.FileInfo, name string, out *os.FileInfo) bool {
	for _, file := range files {
		if file.Name() == name {
			*out = file
			return true
		}
	}
	return false
}
