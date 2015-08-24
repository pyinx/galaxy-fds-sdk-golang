package galaxy_fds_sdk_golang

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"
)

func Signature(app_key, app_secret, method, u, content_md5, content_type string) (string, string) {
	var string_to_sign string
	var uri string
	date := time.Now().Format(time.RFC1123)
	string_to_sign += method + "\n"
	string_to_sign += content_md5 + "\n"
	string_to_sign += content_type + "\n"
	string_to_sign += date + "\n"
	url_str, _ := url.ParseRequestURI(u)
	if strings.Contains(url_str.RequestURI(), "?") {
		uri_list := strings.Split(url_str.RequestURI(), "?")
		if uri_list[1] != "acl" {
			uri = uri_list[0]
		} else {
			uri = url_str.RequestURI()
		}
	} else {
		uri = url_str.RequestURI()
	}
	string_to_sign += uri
	// fmt.Println(string_to_sign)
	h := hmac.New(sha1.New, []byte(app_secret))
	h.Write([]byte(string_to_sign))
	b := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return date, fmt.Sprintf("Galaxy-V2 %s:%s", app_key, b)
}
