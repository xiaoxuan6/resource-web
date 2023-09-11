package main

import (
	"context"
	"embed"
	"encoding/json"
	"github.com/google/go-github/v48/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

type Menu struct {
	Values []string `json:"values"`
}

type data struct {
	Title string `json:"title,omitempty"`
	Items []item `json:"items,omitempty"`
}

type item struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

var (
	c     context.Context
	dbMap map[string]data
	menu  Menu
	lock  sync.RWMutex
	wg    sync.WaitGroup
	//go:embed static
	dirStatic embed.FS
	//go:embed index.html
	fileIndex embed.FS
)

func init() {
	dbMap = make(map[string]data)
	c = context.Background()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}

	menu = fetchMenus()
	for _, val := range menu.Values {
		wg.Add(1)
		go fetchContent(val, &wg)
	}
	wg.Wait()

	http.HandleFunc("/", tplHandler)
	http.HandleFunc("/refresh", refreshHandler)

	//加载静态文件
	fs := http.FileServer(http.FS(dirStatic))
	http.Handle("/static/", fs)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func client() (client *github.Client) {
	oauth := oauth2.NewClient(c, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")}))
	client = github.NewClient(oauth)
	return
}

func fetchMenus() (menu Menu) {
	_, directoryContent, _, _ := client().Repositories.GetContents(c, os.Getenv("GITHUB_OWNER"), os.Getenv("GITHUB_REPO"), "", &github.RepositoryContentGetOptions{})

	var values []string
	for _, val := range directoryContent {
		if strings.HasSuffix(val.GetName(), ".md") {
			values = append(values, val.GetName())
		}
	}

	menu.Values = values
	return menu
}

func fetchContent(filename string, wg *sync.WaitGroup) {
	defer wg.Done()
	repositoryContent, _, _, err := client().Repositories.GetContents(c, os.Getenv("GITHUB_OWNER"), os.Getenv("GITHUB_REPO"), filename, &github.RepositoryContentGetOptions{})
	if err != nil {
		return
	}

	content, err2 := repositoryContent.GetContent()
	if err2 != nil {
		return
	}

	contents := strings.Split(content, "\n")

	// 数据相同直接返回
	if cache, ok := dbMap[filename]; ok {
		if len(contents) == len(cache.Items) {
			return
		}
	}

	var items []item
	for _, val := range contents {
		url := regexpUrl(val)
		if url == "" {
			continue
		}

		title := regexpTitle(val)
		items = append(items, item{
			Title: title,
			Url:   url,
		})
	}

	lock.Lock()
	dbMap[filename] = data{
		Title: filename,
		Items: items,
	}
	lock.Unlock()
}

func regexpTitle(str string) string {
	re := regexp.MustCompile(`\[(.*?)\]`)
	matches := re.FindStringSubmatch(str)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func regexpUrl(str string) string {
	re := regexp.MustCompile(`\((.*?)\)`)
	matches := re.FindStringSubmatch(str)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func tplHandler(w http.ResponseWriter, r *http.Request) {
	// 创建一个新的模板，并设置自定义分隔符为<< >>，避免与Vue的语法冲突
	tmplInstance := template.New("index.html").Delims("<<", ">>")
	//添加加法函数计数
	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
	}
	// 加载模板文件
	tmpl, err := tmplInstance.Funcs(funcMap).ParseFS(fileIndex, "index.html")
	if err != nil {
		log.Println("模板加载错误:", err)
		return
	}

	// 定义一个数据对象
	dates := struct {
		RssDataList []data
	}{
		RssDataList: fetchDatas(),
	}

	// 渲染模板并将结果写入响应
	err = tmpl.Execute(w, dates)
	if err != nil {
		log.Println("模板渲染错误:", err)
	}
}

func fetchDatas() (datas []data) {
	for _, val := range menu.Values {
		lock.RLock()
		cache, ok := dbMap[val]
		lock.RUnlock()
		if !ok {
			log.Printf("Error getting data from db is null %v", val)
			continue
		}

		datas = append(datas, cache)
	}
	return
}

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	for _, val := range menu.Values {
		wg.Add(1)
		go fetchContent(val, &wg)
	}
	wg.Wait()

	w.Header().Set("content-type", "text/json")

	response := struct {
		Status int    `json:"status"`
		Msg    string `json:"msg"`
	}{
		Status: 200,
		Msg:    "ok",
	}

	msg, _ := json.Marshal(response)
	_, _ = w.Write(msg)
}
