package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	if len(os.Args) == 1 {
		log.Println("Please pass in names of mods as arguments.")
	}
	for _, mod := range os.Args[1:] {
		err := getMod(mod)
		if err != nil {
			log.Println(err)
		}
	}
}
func getMod(modName string) error {
	modPage := fmt.Sprintf("https://mods.factorio.com/mod/%s", modName)

	req, err := http.NewRequest("GET", modPage, nil)
	if err != nil {
		return err
	}
	headers(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	pageBody := string(data)

	modDl := "href=\"/download/" + modName + "/"
	i := strings.Index(pageBody, modDl)
	if i == -1 {
		return fmt.Errorf("NO download found for %s", modName)
	}
	i = i + len(modDl)
	modLink := pageBody[i-len(modDl)+6 : i+40]
	modLink = strings.TrimSpace(modLink)
	modLink = modLink[:len(modLink)-2]
	log.Println(modLink)

	req, err = http.NewRequest("GET", fmt.Sprintf("https://mods.factorio.com/%s", modLink), nil)
	if err != nil {
		return err
	}
	headers(req)
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := ioutil.TempFile("", modName+".zip")
	if err != nil {
		return err
	}

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	f.Close()

	name, err := ZipName(f.Name())
	if err != nil {
		return err
	}
	log.Println(name + ".zip")
	return os.Rename(f.Name(), name+".zip")
}

func headers(req *http.Request) {
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Cookie", os.Getenv("FACTORIO_AUTH_COOKIE"))
	req.Header.Set("Connection", "keep-alive")
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		defer rc.Close()

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)
		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {

			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)

		} else {

			// Make File
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return filenames, err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, err
			}

			_, err = io.Copy(outFile, rc)

			// Close the file without defer to close before next iteration of loop
			outFile.Close()

			if err != nil {
				return filenames, err
			}

		}
	}
	return filenames, nil
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func ZipName(src string) (string, error) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return "", err
	}
	defer r.Close()
	return path.Dir(r.File[0].Name), nil
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
