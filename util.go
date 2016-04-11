package galaxy_fds_sdk_golang

import (
	// "crypto/md5"
	"encoding/json"
	"github.com/bitly/go-simplejson"
	// "io"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"
	"bytes"
	"cmd/go/testdata/src/vend/x/vendor/r"
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

type FDSClient struct {
	App_key    string
	App_secret string
}

type FDSAuth struct {
	Url          string
	Method       string
	Data         []byte
	Content_Md5  string
	Content_Type string
	Headers      map[string]string
}

func NEWFDSClient(App_key, App_secret string) *FDSClient {
	c := new(FDSClient)
	c.App_key = App_key
	c.App_secret = App_secret
	return c
}

func (c *FDSClient) Auth(auth FDSAuth) (*http.Response, error) {
	client := &http.Client{}
	var reader bytes.Reader = nil
	if auth.Data != nil {
		reader = bytes.NewReader(auth.Data)
	}

	req, _ := http.NewRequest(auth.Method, auth.Url, reader)
	date, signature := Signature(c.App_key,
		c.App_secret, req.Method, auth.Url,
		auth.Content_Md5, auth.Content_Type)
	for k, v := range auth.Headers {
		req.Header.Add(k, v)
	}
	req.Header.Add("authorization", signature)
	req.Header.Add("date", date)
	req.Header.Add("content-md5", auth.Content_Md5)
	req.Header.Add("content-type", auth.Content_Type)
	res, err := client.Do(req)
	return res, err
}

func (c *FDSClient) Is_Bucket_Exists(bucketname string) (bool, error) {
	url := DEFAULT_FDS_SERVICE_BASE_URI + bucketname
	auth := FDSAuth{
		Url:          url,
		Method:       "HEAD",
		Data:         nil,
		Content_Md5:  "",
		Content_Type: "",
		Headers:      map[string]string{},
	}
	res, err := c.Auth(auth)
	if err != nil {
		return false, err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return false, err
	}
	if res.StatusCode == 200 {
		return true, nil
	} else {
		return false, errors.New(string(body))
	}
}

func (c *FDSClient) List_Bucket() ([]string, error) {
	bucketlist := []string{}
	url := DEFAULT_FDS_SERVICE_BASE_URI
	auth := FDSAuth{
		Url:          url,
		Method:       "GET",
		Data:         nil,
		Content_Md5:  "",
		Content_Type: "",
		Headers:      map[string]string{},
	}
	res, err := c.Auth(auth)
	if err != nil {
		return bucketlist, err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return bucketlist, err
	}
	if res.StatusCode == 200 {
		sj, err := simplejson.NewJson(body)
		if err != nil {
			return bucketlist, err
		}
		buckets, _ := sj.Get("buckets").Array()
		for _, bucket := range buckets {
			// fmt.Printf("%#v\n", bucket.(map[string]interface{})["name"])
			bucket = bucket.(map[string]interface{})["name"]
			bucketlist = append(bucketlist, bucket.(string))
		}
		return bucketlist, nil
	} else {
		return bucketlist, errors.New(string(body))
	}
}

func (c *FDSClient) Create_Bucket(bucketname string) (bool, error) {
	url := DEFAULT_FDS_SERVICE_BASE_URI + bucketname
	auth := FDSAuth{
		Url:          url,
		Method:       "PUT",
		Data:         nil,
		Content_Md5:  "",
		Content_Type: "",
		Headers:      map[string]string{},
	}
	res, err := c.Auth(auth)
	if err != nil {
		return false, err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return false, err
	}
	if res.StatusCode == 200 {
		return true, nil
	} else {
		return false, errors.New(string(body))
	}
}

func (c *FDSClient) Delete_Bucket(bucketname string) (bool, error) {
	url := DEFAULT_FDS_SERVICE_BASE_URI + bucketname
	auth := FDSAuth{
		Url:          url,
		Method:       "DELETE",
		Data:         nil,
		Content_Md5:  "",
		Content_Type: "",
		Headers:      map[string]string{},
	}
	res, err := c.Auth(auth)
	if err != nil {
		return false, err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return false, err
	}
	if res.StatusCode == 200 {
		return true, nil
	} else {
		return false, errors.New(string(body))
	}
}

func (c *FDSClient) Is_Object_Exists(bucketname, objectname string) (bool, error) {
	url := DEFAULT_FDS_SERVICE_BASE_URI + bucketname + DELIMITER + objectname
	auth := FDSAuth{
		Url:          url,
		Method:       "HEAD",
		Data:         nil,
		Content_Md5:  "",
		Content_Type: "",
		Headers:      map[string]string{},
	}
	res, err := c.Auth(auth)
	if err != nil {
		return false, err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return false, err
	}
	if res.StatusCode == 200 {
		return true, nil
	} else {
		return false, errors.New(string(body))
	}
}

func (c *FDSClient) Get_Object(bucketname, objectname string, postion, size int) (string, error) {
	if postion < 0 {
		err := errors.New("Seek position should be no less than 0")
		return "", err
	}
	url := DEFAULT_CDN_SERVICE_URI + bucketname + DELIMITER + objectname
	headers := map[string]string{}
	if postion > 0 {
		headers["range"] = fmt.Sprintf("bytes=%d-", postion)
	}
	auth := FDSAuth{
		Url:          url,
		Method:       "GET",
		Data:         nil,
		Content_Md5:  "",
		Content_Type: "",
		Headers:      headers,
	}
	res, err := c.Auth(auth)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", err
	}
	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusPartialContent {
		content := string(body)
		if len(content) > size {
			content = content[0:size]
		}
		return content, nil
	} else {
		return "", errors.New(string(body))
	}
}

// prefix需要改进
func (c *FDSClient) List_Object(bucketname string) ([]string, error) {
	listobject := []string{}
	url := DEFAULT_FDS_SERVICE_BASE_URI + bucketname + "?prefix=&delimiter=" + DELIMITER
	auth := FDSAuth{
		Url:          url,
		Method:       "GET",
		Data:         nil,
		Content_Md5:  "",
		Content_Type: "",
		Headers:      map[string]string{},
	}
	res, err := c.Auth(auth)
	if err != nil {
		return listobject, err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return listobject, err
	}
	if res.StatusCode == 200 {
		sj, err := simplejson.NewJson(body)
		if err != nil {
			return listobject, err
		}
		objects, _ := sj.Get("objects").Array()
		for _, object := range objects {
			// fmt.Printf("%v\n", object.(map[string]interface{})["name"])
			object = object.(map[string]interface{})["name"]
			listobject = append(listobject, object.(string))
		}
		return listobject, nil
	} else {
		return listobject, errors.New(string(body))
	}
}

// v1类型
func (c *FDSClient) Post_Object(bucketname, data []byte, filetype string) (string, error) {
	url := DEFAULT_FDS_SERVICE_BASE_URI_HTTPS + bucketname + DELIMITER
	if !strings.HasPrefix(filetype, ".") {
		filetype = "." + filetype
	}
	content_type := mime.TypeByExtension(filetype)
	if content_type == "" {
		content_type = "application/octet-stream"
	}
	auth := FDSAuth{
		Url:          url,
		Method:       "POST",
		Data:         data,
		Content_Md5:  "",
		Content_Type: content_type,
		Headers:      map[string]string{},
	}
	res, err := c.Auth(auth)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", err
	}
	if res.StatusCode == 200 {
		sj, err := simplejson.NewJson(body)
		if err != nil {
			return "", err
		}
		objectname, _ := sj.Get("objectName").String()
		return objectname, nil
	} else {
		return "", errors.New(string(body))
	}
}

// v2类型  自定义文件名 如果object已存在，将会覆盖
func (c *FDSClient) Put_Object(bucketname, objectname, data []byte, filetype string) (bool, error) {
	url := DEFAULT_FDS_SERVICE_BASE_URI_HTTPS + bucketname + DELIMITER + objectname
	if !strings.HasPrefix(filetype, ".") {
		filetype = "." + filetype
	}
	content_type := mime.TypeByExtension(filetype)
	if content_type == "" {
		content_type = "application/octet-stream"
	}
	auth := FDSAuth{
		Url:          url,
		Method:       "PUT",
		Data:         data,
		Content_Md5:  "",
		Content_Type: content_type,
		Headers:      map[string]string{},
	}
	res, err := c.Auth(auth)
	if err != nil {
		return false, err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return false, err
	}
	if res.StatusCode == 200 {
		return true, nil
	} else {
		return false, errors.New(string(body))
	}
}

func (c *FDSClient) Delete_Object(bucketname, objectname string) (bool, error) {
	url := DEFAULT_FDS_SERVICE_BASE_URI + bucketname + DELIMITER + objectname
	auth := FDSAuth{
		Url:          url,
		Method:       "DELETE",
		Data:         nil,
		Content_Md5:  "",
		Content_Type: "",
		Headers:      map[string]string{},
	}
	res, err := c.Auth(auth)
	if err != nil {
		return false, err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return false, err
	}
	if res.StatusCode == 200 {
		return true, nil
	} else {
		return false, errors.New(string(body))
	}
}

func (c *FDSClient) Rename_Object(bucketname, src_objectname, dst_objectname string) (bool, error) {
	url := DEFAULT_FDS_SERVICE_BASE_URI + bucketname + DELIMITER + src_objectname +
		"?renameTo=" + dst_objectname
	auth := FDSAuth{
		Url:          url,
		Method:       "PUT",
		Data:         nil,
		Content_Md5:  "",
		Content_Type: "",
		Headers:      map[string]string{},
	}
	res, err := c.Auth(auth)
	if err != nil {
		return false, err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return false, err
	}
	if res.StatusCode == 200 {
		return true, nil
	} else {
		return false, errors.New(string(body))
	}
}

func (c *FDSClient) Prefetch_Object(bucketname, objectname string) (bool, error) {
	url := DEFAULT_FDS_SERVICE_BASE_URI + bucketname + DELIMITER + objectname + "?prefetch"
	auth := FDSAuth{
		Url:          url,
		Method:       "PUT",
		Data:         nil,
		Content_Md5:  "",
		Content_Type: "",
		Headers:      map[string]string{},
	}
	res, err := c.Auth(auth)
	if err != nil {
		return false, err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return false, err
	}
	if res.StatusCode == 200 {
		return true, nil
	} else {
		return false, errors.New(string(body))
	}
}

func (c *FDSClient) Refresh_Object(bucketname, objectname string) (bool, error) {
	url := DEFAULT_FDS_SERVICE_BASE_URI + bucketname + DELIMITER + objectname + "?refresh"
	auth := FDSAuth{
		Url:          url,
		Method:       "PUT",
		Data:         nil,
		Content_Md5:  "",
		Content_Type: "",
		Headers:      map[string]string{},
	}
	res, err := c.Auth(auth)
	if err != nil {
		return false, err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return false, err
	}
	if res.StatusCode == 200 {
		return true, nil
	} else {
		return false, errors.New(string(body))
	}
}

func (c *FDSClient) Set_Object_Acl(bucketname, objectname string, acl map[string]interface{}) (bool, error) {
	acp := make(map[string]interface{})
	acp["owner"] = map[string]string{"id": c.App_key}
	acp["accessControlList"] = []interface{}{acl}
	jsonString, _ := json.Marshal(acp)
	url := DEFAULT_FDS_SERVICE_BASE_URI + bucketname + DELIMITER + objectname + "?acl"
	auth := FDSAuth{
		Url:          url,
		Method:       "PUT",
		Data:         jsonString,
		Content_Md5:  "",
		Content_Type: "",
		Headers:      map[string]string{},
	}
	res, err := c.Auth(auth)
	if err != nil {
		return false, err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return false, err
	}
	if res.StatusCode == 200 {
		return true, nil
	} else {
		return false, errors.New(string(body))
	}
}

func (c *FDSClient) Set_Public(bucketname, objectname string, disable_prefetch bool) (bool, error) {
	grant := map[string]interface{}{
		"grantee":    ALL_USERS,
		"type":       PERMISSION_GROUP,
		"permission": string(PERMISSION_READ),
	}
	// acl := make(map[string]interface{})
	// key := ALL_USERS["id"] + ":" + PERMISSION_GROUP
	// acl[key] = grant
	// result := Set_Object_Acl(bucketname, objectname, acl)
	_, err := c.Set_Object_Acl(bucketname, objectname, grant)
	if err != nil {
		return false, err
	}
	if !disable_prefetch {
		_, err := c.Prefetch_Object(bucketname, objectname)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

// list_object_next
// set_bucket_acl
// get_bucket_acl
// get_object_acl
// get_object_metadata
// generate_presigned_uri
// generate_download_object_uri
