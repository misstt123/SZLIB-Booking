package pkg

import (
	"errors"
	"github.com/rs/zerolog/log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-resty/resty/v2"
)

var (
	cli  = resty.New()
	once sync.Once
)

const baseURL = "https://www.szlib.org.cn"

type Option func(c *resty.Client)

func option(c *resty.Client, opts ...Option) {
	for _, o := range opts {
		o(c)
	}
}

func WithBaseURLMiddleware(baseURL string) Option {
	return func(c *resty.Client) {
		c.SetBaseURL(baseURL)
	}
}

func WithTimeoutMiddleware(timeout time.Duration) Option {
	return func(c *resty.Client) {
		c.SetTimeout(timeout)
		c.SetRetryCount(5)
	}
}

func WithCommonHeader() Option {
	return func(c *resty.Client) {
		c.OnBeforeRequest(func(client *resty.Client, req *resty.Request) error {
			req.SetHeader("Accept", "application/json, text/plain, */*").
				SetHeader("Accept-Language", "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7").
				SetHeader("Cache-Control", "no-cache").
				SetHeader("Connection", "keep-alive").
				SetHeader("Origin", "https://www.szlib.org.cn").
				SetHeader("Pragma", "no-cache").
				SetHeader("Referer", "https://www.szlib.org.cn/wxbooking/index").
				SetHeader("Sec-Fetch-Dest", "empty").
				SetHeader("Sec-Fetch-Mode", "cors").
				SetHeader("Sec-Fetch-Site", "same-origin").
				SetHeader("User-Agent", "Mozilla/5.0 (Linux; Android 14; V2183A Build/UP1A.231005.007; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/126.0.6478.122 Mobile Safari/537.36 XWEB/1260059 MMWEBSDK/20240404 MMWEBID/3376 MicroMessenger/8.0.49.2600(0x2800315A) WeChat/arm64 Weixin NetType/WIFI Language/zh_CN ABI/arm64").
				SetHeader("X-Requested-With", "com.tencent.mm").
				SetHeader("sec-ch-ua", "\"Not/A)Brand\";v=\"8\", \"Chromium\";v=\"126\", \"Android WebView\";v=\"126\"").
				SetHeader("sec-ch-ua-mobile", "?1").
				SetHeader("sec-ch-ua-platform", "\"Android\"")
			return nil
		})

	}
}

// OperateToken token响应体
type OperateToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	expireTime   time.Time
}

// isNeedRefresh 判断是否需要刷新
func (o *OperateToken) isNeedRefresh() bool {
	if o == nil || o.AccessToken == "" {
		return true
	}
	if time.Now().After(o.expireTime) {
		return true
	}
	return false

}

// oauthUrl 获取授权码api
const oauthUrl = "/operatorauthapi/auth/oauth/token"

var (
	// operatorToken 操作码
	operatorToken atomic.Value

	Wxuserinfo = `{"openid":"oNDDOwfWBv43OjFMWDYLvRiqjcc0","nickname":"用户已经注销","sex":0,"language":"","city":"","province":"","country":"","headimgurl":"https:\/\/thirdwx.qlogo.cn\/mmopen\/vi_32\/PiajxSqBRaELSFYibJFEiaxK4ykiaibiclu7Ey6vOw3D7VVCbs1U5dma6nbFjIPJsZiaamH7ta7bsL8DEG4EMo01y9T3PgY2BE5ibAQh3y1ujzXOGx4u7gaHb9KTpA\/132","privilege":[]}
`
)

// WithCookie 设置cookie
// 后续可以存在db中
func WithCookie() Option {
	return func(c *resty.Client) {
		encodedURL := url.QueryEscape(Wxuserinfo)
		c.OnBeforeRequest(func(client *resty.Client, req *resty.Request) error {
			req.SetCookies(
				[]*http.Cookie{
					&http.Cookie{
						Name:  "Wxuserinfo",
						Value: encodedURL,
					},
					&http.Cookie{
						Name:  "_pk_ses.8.8720",
						Value: "*",
					},
					&http.Cookie{
						Name:  "_pk_id.8.8720",
						Value: "a65a55c139cdb499.1721532344.2.1721534801.1721532432",
					},
				})

			if strings.Contains(req.URL, oauthUrl) {
				return nil
			}

			if token, ok := operatorToken.Load().(*OperateToken); ok && !token.isNeedRefresh() {
				req.SetCookie(&http.Cookie{
					Name:  "operatorToken",
					Value: token.AccessToken,
				},
				)
				return nil
			}

			res, err := GetClient().
				R().
				SetQueryParams(map[string]string{
					"grant_type":    "password",
					"scope":         "app",
					"client_id":     "t1",
					"client_secret": "szlib1114",
					"eventsite":     "SELF",
				}).
				SetFormData(map[string]string{
					"username": "stn1f99",
					"password": "755d20d85c31b568be24015e924d0dec",
				}).
				SetResult(&OperateToken{}).
				Post(oauthUrl)
			if err != nil {
				log.Err(err).Str("url", oauthUrl).Msg("invoke operateToken fail")
				return err
			}
			var (
				token *OperateToken
				ok    bool
			)
			if token, ok = res.Result().(*OperateToken); !ok || token == nil {
				log.Error().Str("url", oauthUrl).Msg("get response body fail")
				return errors.New("token marshal response fail")
			}

			token.expireTime = time.Now().Add(time.Second * time.Duration(token.ExpiresIn))
			operatorToken.Store(token)

			req.SetCookie(&http.Cookie{
				Name:  "operatorToken",
				Value: token.AccessToken,
			},
			)

			return nil
		})

	}
}

func GetClient() *resty.Client {
	once.Do(func() {
		option(
			cli,
			WithBaseURLMiddleware(baseURL),
			WithTimeoutMiddleware(45*time.Second),
			WithCommonHeader(),
			WithCookie(),
			//WithHeaderBodyMiddleware(),
			//WithTraceMiddleware(),
			//WithLogMiddleware(),
		)
	})
	return cli
}
