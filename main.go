package main

import (
	"errors"
	"fmt"
	"github.com/Poppington/Downloader"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

const BASE_FOLDER = "download/"

var (
	URLREGEX = regexp.MustCompile(`www\.imagefap\.com\/(pictures|gallery)\/([\d]+)`)
)

type Gallery struct {
	Id         int
	GalleryUrl string
	OnePageUrl string
	FolderName string
}

func main() {
	os.MkdirAll(BASE_FOLDER, os.ModeDir)
	if len(os.Args) < 2 {
		fmt.Println("please enter the gallery/galleries url(s)")
		return
	}
	galleries := []Gallery{}
	for _, arg := range os.Args {
		gallery, err := AddArgument(arg)
		if err == nil {
			galleries = append(galleries, gallery)
		}
	}
	errors := DownloadGalleries(galleries)
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Println(err.Error())
		}
	}
}

func DownloadGalleries(galleries []Gallery) []error {
	errored := []error{}
	for _, gallery := range galleries {
		fmt.Println(gallery.OnePageUrl + " -> " + gallery.FolderName)
		doc, err := goquery.NewDocument(gallery.OnePageUrl)
		if err != nil {
			errored = append(errored, err)
			continue
		}
		doc.Find("#gallery form table tbody").Each(func(i int, s *goquery.Selection) {
			s.Find("tr").Each(func(i int, s *goquery.Selection) {
				s.Find("td").Each(func(i int, s *goquery.Selection) {
					row := s.Find("table tbody tr").First()
					link := row.Find("td a").First()
					href, _ := link.Attr("href")
					if href != "" {
						url := "http://www.imagefap.com" + href
						// fmt.Printf("%s\n%s\n%s\n%s\n%s\n\n", gallery.FolderName, gallery.OnePageUrl, url, doc.Url.String(), BASE_FOLDER+gallery.FolderName)
						DownloadImageFromPage(url, BASE_FOLDER+gallery.FolderName)
					}
				})
			})
		})
	}
	return errored
}

func DownloadImageFromPage(url, folder string) error {
	get, err := http.Get(url)
	if err != nil {
		return err
	}
	defer get.Body.Close()

	contents, err := ioutil.ReadAll(get.Body)
	if err != nil {
		return err
	}
	reg := regexp.MustCompile(`\"contentUrl\": \"([\S]+)\"\,`)
	find := reg.FindStringSubmatch(string(contents))
	imgsrc := find[len(find)-1]
	os.MkdirAll(folder, os.ModeDir)
	fileloc := folder + path.Base(imgsrc)
	downloader.DownloadFile(fileloc, imgsrc, "*")
	return nil
}

func AddArgument(arg string) (Gallery, error) {
	gallery := Gallery{}
	find := URLREGEX.FindStringSubmatch(arg)
	if find == nil {
		return gallery, errors.New("No ImageFap Gallery Found")
	}

	idstr := find[len(find)-1]
	id, err := strconv.Atoi(idstr)
	if err != nil {
		return gallery, err
	}
	gallery.Id = id
	gallery.GalleryUrl = fmt.Sprintf("http://www.imagefap.com/gallery/%d", gallery.Id)
	head, err := http.Head(gallery.GalleryUrl)
	if err != nil {
		return gallery, err
	}
	split := strings.Split(head.Request.URL.String(), "/")
	gallery.FolderName = fmt.Sprintf("%d-%s/", gallery.Id, split[len(split)-1])
	gallery.OnePageUrl = head.Request.URL.String() + fmt.Sprintf("?gid=%d&view=2", gallery.Id)
	return gallery, nil
}
