# galaxy-fds-sdk-golang
Golang SDK for Xiaomi File Data Storage.
# install
```
go get github.com/pyinx/galaxy-fds-sdk-golang
```
#example
```
package main

import (
	"fmt"
	"github.com/pyinx/galaxy-fds-sdk-golang"
)

func main() {
	c := galaxy_fds_sdk_golang.NewFDSClient("YOUR_APP_KEY", "YOUR_APP_SECRET")
	fmt.Println(c.Create_Bucket("test-testaaaaaaaaaa"))
	fmt.Println(c.Delete_Bucket("test-testaaaaaaaaaa"))
	c.List_Bucket()
	fmt.Println(c.Is_Bucket_Exists("test-testaaaaaa"))
	fmt.Println(c.Is_Object_Exists("test-testaaaaaa", "a.jpg"))
	content := c.Get_Object("test-testaaaaaa", "2.txt", 0, 100)
	fmt.Println(content)
	c.List_Object("test-testaaaaaa")
	c.Post_Object("test-testaaaaaa", "abcdefgssss")
	c.Put_Object("test-testaaaaaa", "1.txt", "abcdefg")
	fmt.Println(c.Delete_Object("test-testaaaaaa", "2.txt"))
	fmt.Println(c.Rename_Object("test-testaaaaaa", "1.txt", "2.txt"))
	// Set_Public最后一个参数表示是否需要关闭CDN预取，如无特殊需要建议设成true
	// 放回两个值: 第一个是setpublic是否成功,第二个是cdn预取是否成功，默认为true
	fmt.Println(c.Set_Public("test-testaaaaaa", "2.txt", true))
	fmt.Println(c.Refresh_Object("test-testaaaaaa", "2.txt"))
}

```
