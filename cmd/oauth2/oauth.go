package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	dropboxID := os.Getenv("DROPBOX_APP_ID")
	dropboxSecret := os.Getenv("DROPBOX_APP_SECRET")

	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID:     dropboxID,                                             // 第三方应用注册时使用唯一标识符ID
		ClientSecret: dropboxSecret,                                         // 与Client配对的密钥
		Scopes:       []string{"files.metadata.read", "files.content.read"}, // 类似请求的权限
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.dropbox.com/oauth2/authorize", // 用来获得认证，资源服务器的认证
			TokenURL: "https://api.dropboxapi.com/oauth2/token",  // 用于获得访问的token，令牌端点URL
		},
	}

	// use PKCE to protect against CSRF attacks
	// https://www.ietf.org/archive/id/draft-ietf-oauth-security-topics-22.html#name-countermeasures-6
	verifier := oauth2.GenerateVerifier()

	// Redirect user to consent page to ask for permission for the scopes specified above.
	// url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))  离线访问令牌的方法
	url := conf.AuthCodeURL("fake-state", oauth2.SetAuthURLParam("token_access_type", "offline"))
	fmt.Printf("Visit the URL for the auth dialog: %v\n", url)
	fmt.Printf("Once you have a code, paste it and press enter: ")

	// Use the authorization code that is pushed to the redirect URL. Exchange will do the handshake to retrieve the
	// initial access token. The HTTP Client returned by conf.Client will refresh the token as necessary.
	var code string
	if _, err := fmt.Scan(&code); err != nil { // 等待用户输入授权码，实际上是自动处理而不是通过标准输入
		log.Fatal(err)
	}
	tok, err := conf.Exchange(ctx, code, oauth2.VerifierOption(verifier)) // 通过Exchange方法，通过授权码和PKCE来获取访问令牌
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(tok)

	// 创建一个携带访问令牌的HTTP客户端，然后可以用这个客户端进行API调用
	client := conf.Client(ctx, tok)
	// client.Get("...") // 调用实例API
	resp, err := client.Post("https://api.dropboxapi.com/2/files/list_folder", "application/json", strings.NewReader(`{
		"path": ""
	}`))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)

}

// 在线访问令牌和离线访问令牌
// 在线访问令牌：不需要刷新令牌，如果访问令牌国企，用户需要在场重新验证
// 适用场景：用户主动发起操作的场景，如：用户想要将Dropbox的一个文件夹转化为画廊，或者应用程序希望用户使用OAuth进行登录，此时
// 用户会主动发起一个操作，并且该操作通常会在令牌过期之前完成

// 离线访问令牌：拥有刷新令牌，即使用户不在场，应用程序也可以刷新访问令牌。
// 适用场景：适用于用户不在场但是需要执行操作的场景。例如：应用程序需要定期检查文件夹中的新文件，并将它们同步到画廊。
// 这种操作可能需要每天或者每周进行一次，但用户在那个时间段不会登录。

// DropBox默认使用在线访问令牌，且对他们的参数使用的不同的key，因此，如果希望离线工作，需要稍有不同

// URL和URI之间的差别
// URI stands for uniform resource identifier. URL stands for uniform resource locator.
// URL 用于标识某一互联网资源。URI可以用URL(统一资源定位符)或URN(统一资源名称)组成。
// URL，即统一资源定位符，是用于指定互联网上资源位置的字符串。它告诉浏览器或其他Web客户端如何访问特定的Web页面、图片、视频或其他资源。
// URL通常包括协议（如http或https）、域名（或IP地址）、端口号（可选）、路径以及可能的其他参数。

// URN，即统一资源名称，与URL不同，它并不提供资源的网络位置信息，而是提供一种命名机制，用于唯一标识资源。
// URN主要用于那些不需要位置信息就能被访问的资源。
