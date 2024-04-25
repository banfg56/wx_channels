package wxchannels

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/imroc/req/v3"
)

//腾讯视频号相关

const (
	UA_MAC_FIREFOX = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36`
)

type _wxBase struct { //基类
	AdminCookies *cookiejar.Jar
	Ua           string
	AppLog       *log.Helper
}

func (b *_wxBase) GetUADefault() string {
	if len(b.Ua) > 0 {
		return b.Ua
	}
	return UA_MAC_FIREFOX
}

func (b *_wxBase) SetUa(ua string) {
	b.Ua = ua
}

func (b *_wxBase) GetReqClient(_headers map[string]string) *req.Client {
	client := req.C().
		SetTimeout(10 * time.Second).
		SetUserAgent(b.GetUADefault()).SetCookieJar(b.AdminCookies) //.EnableDumpAll()

	if len(_headers) > 0 {
		client = client.SetCommonHeaders(_headers)
	}
	return client
}

func (b *_wxBase) SetAdminCookie(cs []*http.Cookie) (err error) { //默认都是https
	if b.AdminCookies == nil {
		err = fmt.Errorf("CookieJar为空")
		return
	}
	// 根据不同cookie的域名来设置
	diffDomain := make(map[string][]*http.Cookie)
	for _, cv := range cs {
		if _, isSet := diffDomain[cv.Domain]; !isSet {
			diffDomain[cv.Domain] = make([]*http.Cookie, 0)
		}
		diffDomain[cv.Domain] = append(diffDomain[cv.Domain], cv)
	}

	for dName, dCookie := range diffDomain {
		domainUrl := ""
		if len(dName) > 2 && strings.HasPrefix(dName, `.`) {
			domainUrl = fmt.Sprintf("https://%s", strings.TrimPrefix(dName, `.`))
		} else if len(dName) > 2 && !strings.HasPrefix(dName, `.`) {
			domainUrl = fmt.Sprintf("https://%s", dName)
		}
		uDomain, udErr := url.Parse(domainUrl)
		if udErr == nil {
			b.AdminCookies.SetCookies(uDomain, dCookie)
		} else {
			log.Errorf("Errror set:%s  cookie异常:%s", dName, udErr.Error())
		}
	}
	return
}
