package fittencode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gogf/gf/v2/encoding/gjson"
)

const (
	LoginUrl  string = "https://fc.fittenlab.cn/codeuser/login"
	FicoUrl   string = "https://fc.fittenlab.cn/codeuser/get_ft_token"
	ServerUrl string = "https://fc.fittenlab.cn/codeapi/completion/generate_one_stage/"
)

type FittenCode struct {
	Token     string
	FcioToken string
}

func New() *FittenCode {
	return &FittenCode{}
}

func (fc *FittenCode) Login() error {
	data := []byte(`{"username": "", "password": ""}`)

	// 创建一个 POST 请求
	resp, err := http.Post(LoginUrl, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// 读取响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	gj, err := gjson.DecodeToJson(body)
	if err != nil {
		return err
	}
	fc.Token = gj.Get("data.token").String()
	return nil
}

func (fc *FittenCode) GetFcioToken() error {
	fc.Login()
	if fc.Token == "" {
		return fmt.Errorf("token is empty")
	}
	// 创建一个 GET 请求对象
	req, err := http.NewRequest("GET", FicoUrl, nil)
	if err != nil {
		return err
	}

	// 设置请求头
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", fc.Token))

	// 使用 http.Client 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	gj, err := gjson.DecodeToJson(body)
	if err != nil {
		return err
	}
	fc.FcioToken = gj.Get("data.fico_token").String()
	return nil
}

type CompletionMetaDatas struct {
	FileName string `json:"filename"`
}

type CompletionData struct {
	Inputs    string              `json:"inputs"`
	MetaDatas CompletionMetaDatas `json:"meta_datas"`
}

func (fc *FittenCode) Complete() error {
	fc.GetFcioToken()

	prefix := `import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

func main() {
	// 定义请求的 URL
	url := "https://example.com/api"

	// 定义请求的参数
	data := url.Values{}
	data.Set("key1", "value1")
	data.Set("key2", "value2")

	// 创建一个 POST 请求
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}

	// 设置 Content-Type 为 application/x-www-form-urlencoded
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
`
	suffix := `}`

	prompt := "!FCPREFIX!%s!FCSUFFIX!%s!FCMIDDLE!"
	prompt = fmt.Sprintf(prompt, strconv.Quote(prefix), strconv.Quote(suffix))

	data := CompletionData{
		Inputs:    prompt,
		MetaDatas: CompletionMetaDatas{FileName: "main.go"},
	}

	jsonData, _ := json.Marshal(data)

	sUrl := fmt.Sprintf("%s%s?ide=vim&v=0.2.1", ServerUrl, fc.FcioToken)
	req, err := http.NewRequest("POST", sUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	// 设置请求头 application/json
	req.Header.Set("Content-Type", "application/json")

	// 使用 http.Client 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(body))
	return nil
}

func TestFitten() {
	f := New()
	f.Complete()
}
