package galaxy_fds_sdk_golang

import (
	// "crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	// "io"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	DEFAULT_FDS_SERVICE_BASE_URI       = "http://files.fds.api.xiaomi.com/"
	DEFAULT_FDS_SERVICE_BASE_URI_HTTPS = "https://files.fds.api.xiaomi.com/"
	DEFAULT_CDN_SERVICE_URI            = "http://cdn.fds.api.xiaomi.com/"
	USER_DEFINED_METADATA_PREFIX       = "x-xiaomi-meta-"
	DELIMITER                          = "/"
)

// permission
const (
	PERMISSION_READ         = "READ"
	PERMISSION_WRITE        = "WRITE"
	PERMISSION_FULL_CONTROL = "FULL_CONTROL"
	PERMISSION_USER         = "USER"
	PERMISSION_GROUP        = "GROUP"
)

var ALL_USERS = map[string]string{"id": "ALL_USERS"}
var AUTHENTICATED_USERS = map[string]string{"id": "AUTHENTICATED_USERS"}

var PRE_DEFINED_METADATA = []string{"cache-control",
	"content-encoding",
	"content-length",
	"content-md5",
	"content-type",
}

type Client struct {
	App_key    string
	App_secret string
}

func NewClient(App_key, App_secret string) *Client {
	c := new(Client)
	c.App_key = App_key
	c.App_secret = App_secret
	return c
}

func (c *Client) Is_Bucket_Exit(bucketname string) bool {
	client := &http.Client{}
	url := DEFAULT_FDS_SERVICE_BASE_URI + bucketname
	req, _ := http.NewRequest("HEAD", url, nil)
	date, signature := Signature(c.App_key, c.App_secret, req.Method, url)
	req.Header.Add("authorization", signature)
	req.Header.Add("date", date)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if res.StatusCode == 200 {
		return true
	} else {
		return false
	}
}

func (c *Client) List_Bucket() {
	client := &http.Client{}
	url := DEFAULT_FDS_SERVICE_BASE_URI
	req, _ := http.NewRequest("GET", url, nil)
	date, signature := Signature(c.App_key, c.App_secret, req.Method, url)
	req.Header.Add("authorization", signature)
	req.Header.Add("date", date)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		panic(err)
	}
	sj, err := simplejson.NewJson(body)
	if err != nil {
		panic(err)
	}
	buckets, _ := sj.Get("buckets").Array()
	for _, bucket := range buckets {
		// for _, key := range bucket.(map[string]interface{}) {
		// 	fmt.Println(key)
		// }
		fmt.Printf("%v\n", bucket.(map[string]interface{})["name"])
	}
}

func (c *Client) Create_Bucket(bucketname string) bool {
	client := &http.Client{}
	url := DEFAULT_FDS_SERVICE_BASE_URI + bucketname
	req, _ := http.NewRequest("PUT", url, nil)
	date, signature := Signature(c.App_key, c.App_secret, req.Method, url)
	req.Header.Add("authorization", signature)
	req.Header.Add("date", date)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if res.StatusCode == 200 {
		return true
	} else {
		return false
	}
}

func (c *Client) Delete_Bucket(bucketname string) bool {
	client := &http.Client{}
	url := DEFAULT_FDS_SERVICE_BASE_URI + bucketname
	req, _ := http.NewRequest("DELETE", url, nil)
	date, signature := Signature(c.App_key, c.App_secret, req.Method, url)
	req.Header.Add("authorization", signature)
	req.Header.Add("date", date)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if res.StatusCode == 200 {
		return true
	} else {
		return false
	}
}

func (c *Client) Is_Object_Exit(bucketname, objectname string) bool {
	client := &http.Client{}
	url := DEFAULT_FDS_SERVICE_BASE_URI + bucketname + DELIMITER + objectname
	req, _ := http.NewRequest("HEAD", url, nil)
	date, signature := Signature(c.App_key, c.App_secret, req.Method, url)
	req.Header.Add("authorization", signature)
	req.Header.Add("date", date)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if res.StatusCode == 200 {
		return true
	} else {
		return false
	}
}

func (c *Client) Get_Object(bucketname, objectname string, postion, size int) interface{} {
	if postion < 0 {
		panic("Seek position should be no less than 0")
	}
	client := &http.Client{}
	url := DEFAULT_FDS_SERVICE_BASE_URI + bucketname + DELIMITER + objectname
	req, _ := http.NewRequest("GET", url, nil)
	if postion > 0 {
		req.Header.Add("range", fmt.Sprintf("bytes=%d-", size))
	}
	date, signature := Signature(c.App_key, c.App_secret, req.Method, url)
	req.Header.Add("authorization", signature)
	req.Header.Add("date", date)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusPartialContent {
		obj := make(map[string]map[string]string)
		summary := make(map[string]string)
		metadata := make(map[string]string)
		summary["bucket_name"] = bucketname
		summary["object_name"] = objectname
		summary["size"] = string(res.ContentLength)
		// obj["stream"]
		obj["summary"] = summary
		for _, key := range PRE_DEFINED_METADATA {
			metadata[key] = res.Header.Get(key)
		}
		obj["metadata"] = metadata
		return obj
		// obj = FDSObject()
		// obj.stream = response.iter_content(chunk_size=size)
		// summary = FDSObjectSummary()
		// summary.bucket_name = bucket_name
		// summary.object_name = object_name
		// summary.size = int(response.headers['content-length'])
		// obj.summary = summary
		// obj.metadata = self._parse_object_metadata_from_headers(response.headers)
		// return obj
	} else {
		reason, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			panic(err)
		}
		panic(fmt.Sprintf("Get object failed, status=%d, reason=%s",
			res.StatusCode, string(reason)))
	}
}

func (c *Client) List_Object(bucketname, prefix string) {
	client := &http.Client{}
	url := DEFAULT_FDS_SERVICE_BASE_URI + bucketname + "?prefix=" + prefix + "&delimiter=" + DELIMITER
	req, _ := http.NewRequest("GET", url, nil)
	date, signature := Signature(c.App_key, c.App_secret, req.Method, url)
	req.Header.Add("authorization", signature)
	req.Header.Add("date", date)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		panic(err)
	}
	if res.StatusCode == 200 {
		sj, err := simplejson.NewJson(body)
		if err != nil {
			panic(err)
		}
		objects, _ := sj.Get("objects").Array()
		for _, object := range objects {
			fmt.Printf("%v\n", object.(map[string]interface{})["name"])
		}
	} else {
		fmt.Println("[]")
	}
}

// v1类型
func (c *Client) Post_Object(bucketname, data string) {
	// h := md5.New()
	// io.WriteString(h, data)
	client := &http.Client{}
	url := DEFAULT_FDS_SERVICE_BASE_URI_HTTPS + bucketname + DELIMITER
	req, _ := http.NewRequest("POST", url, strings.NewReader(data))
	date, signature := Signature(c.App_key, c.App_secret, req.Method, url)
	req.Header.Add("authorization", signature)
	req.Header.Add("date", date)
	// req.Header.Add("content-md5", string(h.Sum(nil)))
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		panic(err)
	}
	if res.StatusCode == 200 {
		fmt.Println(string(body))
	} else {
		fmt.Println(string(body))
		fmt.Println(res.StatusCode)
	}
}

// v2类型  自定义文件名 如果object已存在，将会覆盖
func (c *Client) Put_Object(bucketname, objectname, data string) {
	client := &http.Client{}
	url := DEFAULT_FDS_SERVICE_BASE_URI_HTTPS + bucketname + DELIMITER + objectname
	req, _ := http.NewRequest("PUT", url, strings.NewReader(data))
	date, signature := Signature(c.App_key, c.App_secret, req.Method, url)
	req.Header.Add("authorization", signature)
	req.Header.Add("date", date)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		panic(err)
	}
	if res.StatusCode == 200 {
		fmt.Println(string(body))
	} else {
		fmt.Println(string(body))
		fmt.Println(res.StatusCode)
	}
}

func (c *Client) Delete_Object(bucketname, objectname string) bool {
	client := &http.Client{}
	url := DEFAULT_FDS_SERVICE_BASE_URI + bucketname + DELIMITER + objectname
	req, _ := http.NewRequest("DELETE", url, nil)
	date, signature := Signature(c.App_key, c.App_secret, req.Method, url)
	req.Header.Add("authorization", signature)
	req.Header.Add("date", date)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if res.StatusCode == 200 {
		return true
	} else {
		return false
	}
}

func (c *Client) Rename_Object(bucketname, src_objectname, dst_objectname string) bool {
	client := &http.Client{}
	url := DEFAULT_FDS_SERVICE_BASE_URI + bucketname + DELIMITER + src_objectname +
		"?renameTo=" + dst_objectname
	req, _ := http.NewRequest("PUT", url, nil)
	date, signature := Signature(c.App_key, c.App_secret, req.Method, url)
	req.Header.Add("authorization", signature)
	req.Header.Add("date", date)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if res.StatusCode == 200 {
		return true
	} else {
		return false
	}
}

func (c *Client) Prefetch_Object(bucketname, objectname string) bool {
	client := &http.Client{}
	url := DEFAULT_FDS_SERVICE_BASE_URI + bucketname + DELIMITER + objectname + "?prefetch"
	req, _ := http.NewRequest("PUT", url, nil)
	date, signature := Signature(c.App_key, c.App_secret, req.Method, url)
	req.Header.Add("authorization", signature)
	req.Header.Add("date", date)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if res.StatusCode == 200 {
		return true
	} else {
		// 打印错误信息
		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		fmt.Println(string(body))
		return false
	}
}

func (c *Client) Refresh_Object(bucketname, objectname string) bool {
	client := &http.Client{}
	url := DEFAULT_FDS_SERVICE_BASE_URI + bucketname + DELIMITER + objectname + "?refresh"
	req, _ := http.NewRequest("PUT", url, nil)
	date, signature := Signature(c.App_key, c.App_secret, req.Method, url)
	req.Header.Add("authorization", signature)
	req.Header.Add("date", date)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if res.StatusCode == 200 {
		return true
	} else {
		// 打印错误信息
		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		fmt.Println(string(body))
		return false
	}
}

func (c *Client) Set_Object_Acl(bucketname, objectname string, acl map[string]interface{}) bool {
	acp := make(map[string]interface{})
	acp["owner"] = map[string]string{"id": c.App_key}
	acp["accessControlList"] = []interface{}{acl}
	jsonString, _ := json.Marshal(acp)
	fmt.Println(string(jsonString))
	client := &http.Client{}
	url := DEFAULT_FDS_SERVICE_BASE_URI + bucketname + DELIMITER + objectname + "?acl"
	req, _ := http.NewRequest("PUT", url, strings.NewReader(string(jsonString)))
	date, signature := Signature(c.App_key, c.App_secret, req.Method, url)
	req.Header.Add("authorization", signature)
	req.Header.Add("date", date)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if res.StatusCode == 200 {
		return true
	} else {
		// 打印错误信息
		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		fmt.Println(string(body))
		return false
	}
}

func (c *Client) Set_Public(bucketname, objectname string, is_prefetch bool) (bool, bool) {
	var setpublic_result bool
	var prefetch_result bool
	grant := map[string]interface{}{
		"grantee":    ALL_USERS,
		"type":       PERMISSION_GROUP,
		"permission": string(PERMISSION_READ),
	}
	// acl := make(map[string]interface{})
	// key := ALL_USERS["id"] + ":" + PERMISSION_GROUP
	// acl[key] = grant
	// result := Set_Object_Acl(bucketname, objectname, acl)
	setpublic_result = c.Set_Object_Acl(bucketname, objectname, grant)
	if !is_prefetch {
		prefetch_result = c.Prefetch_Object(bucketname, objectname)
	} else {
		prefetch_result = true //默认设为true
	}
	return setpublic_result, prefetch_result
}

// list_object_next
// set_bucket_acl
// get_bucket_acl
// get_object_acl
// get_object_metadata
// generate_presigned_uri
// generate_download_object_uri
