package route_middleware

import (
	"Gin-IPs/src/configure"
	"Gin-IPs/src/dao"
	"Gin-IPs/src/route/response"
	"Gin-IPs/src/utils/uuid"
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

var SnowWorker, _ = uuid.NewSnowWorker(100) // 随机生成一个uuid，100是节点的值（随便给一个）

/*
接口认证模块，根据申请到的 accessKey（公钥）和secretKey(私钥）加密验证
只解析post/put json 和 get/delete url query 的请求，不解析post form的请求
如果时间 expires 字段和当前的时间相差超过5s，则返回认证过期
// 使用 params 参数尽量使用 POST请求，Get请求可能会导致参数过长截断URL

签名算法：Signature = HMAC-SHA1('SecretKey', UTF-8-Encoding-Of( StringToSign ) ) );
StringToSign = method + "\n" +
               URL + "\n" +
               Sort-UrlParams + "\n" +
               Content-MD5 + "\n" +  // md5(params)
               Expires + "\n" +
               AccessKey;
*/
func Validate() gin.HandlerFunc {
	return func(c *gin.Context) {
		response := route_response.Response{}
		response.Data.List = []interface{}{} // 初始化为空切片，而不是空引用
		traceId := SnowWorker.GetId()
		c.Writer.Header().Set("X-Request-Trace-Id", traceId)

		uri := c.Request.URL.Path
		// remoteAddr := c.ClientIP()  // 也可以对客户端IP进行限制
		contentType := c.Request.Header.Get("Content-Type")
		if contentType != "application/json" {
			c.Abort()
			response.Code, response.Message = configure.RequestParameterTypeError, "Content-Type 类型只支持 application/json"
			c.JSON(http.StatusUnauthorized, response)
			return
		}
		accessKey := c.DefaultQuery("accesskey", "")
		expires := c.DefaultQuery("expires", "")
		signature := c.DefaultQuery("signature", "")
		if accessKey == "" {
			c.Abort()
			response.Code, response.Message = configure.RequestParameterMiss, "Token缺失"
			c.JSON(http.StatusUnauthorized, response)
			return
		}
		secret, err := dao.FetchSecret(accessKey)
		if err != nil || "valid" != secret.State {
			c.Abort()
			response.Code, response.Message = configure.RequestKeyNotFound, "无效的Token"
			c.JSON(http.StatusUnauthorized, response)
			return
		}
		c.Writer.Header().Set("X-Request-User", secret.User)

		if expires == "" {
			c.Abort()
			response.Code, response.Message = configure.RequestParameterMiss, "有效期参数缺失"
			c.JSON(http.StatusUnauthorized, response)
			return
		}
		if signature == "" {
			c.Abort()
			response.Code, response.Message = configure.RequestParameterMiss, "签名缺失"
			c.JSON(http.StatusUnauthorized, response)
			return
		}

		secretKey := secret.SecretKey
		if nowTs, err := strconv.ParseInt(expires, 10, 64); err != nil {
			c.Abort()
			response.Code, response.Message = configure.RequestParameterTypeError, "有效期参数类型错误"
			c.JSON(http.StatusUnauthorized, response)
			return
		} else {
			passTime := time.Now().Unix() - nowTs
			if passTime < 0 || passTime >= configure.GinConfigValue.Expires {
				// 容错时间越大越容易被攻击，10秒左右是为了解决reload重启导致的时间差，同时服务器时间必须准
				c.Abort()
				response.Code, response.Message = configure.RequestExpired, "请求已过期"
				c.JSON(http.StatusUnauthorized, response)
				return
			}
		}

		// method is GET or DELETE , params build  the url_params
		method := strings.ToUpper(c.Request.Method)
		// method is POST or PUT, params build the bodyContent
		// json 格式的参数，如果获取到json则以json为准，不会再解析 query string 格式的
		var urlParams, params string
		if "POST" == method || "PUT" == method {
			body, _ := ioutil.ReadAll(c.Request.Body)
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body)) // 重设body
			params = string(body)
		} else if "GET" == method || "DELETE" == method {
			queryParams := c.Request.URL.Query()
			allParams := make(map[string]string)
			for k, v := range queryParams {
				if k != "accesskey" && k != "expires" && k != "signature" {
					allParams[k] = v[0] // 如果某个key传入了2个，只用第一个的值
				}
			}
			keys := getMapKeysSorted(allParams)
			for _, k := range keys {
				urlParams += k + allParams[k]
			}
		}
		if signatureString, err := genSignature(accessKey, secretKey, uri, method, urlParams, params, expires); err != nil {
			c.Abort()
			response.Code, response.Message = configure.ApiGenSignatureError, "API内部错误"
			c.JSON(http.StatusUnauthorized, response)
			return
		} else {
			if signature != signatureString {
				c.Abort()
				response.Code, response.Message = configure.RequestAuthorizedFailed, "API认证失败"
				c.JSON(http.StatusUnauthorized, response)
				return
			}
		}
		//fmt.Println(c.Request.RequestURI)
		//fmt.Println(c.Request.RemoteAddr)

		//response.Code, response.Message = configure.RequestSuccess, secretKey
		//c.JSON(http.StatusOK, response)

		c.Next() //该句可以省略，写出来只是表明可以进行验证下一步中间件，不写，也是内置会继续访问下一个中间件的
	}
}

// 签名算法如下
/*
Signature = HMAC-SHA1('SecretKey', UTF-8-Encoding-Of( StringToSign ) ) );
StringToSign = method + "\n" +
               URL + "\n" +
               Sort-UrlParams + "\n" +
               Content-MD5 + "\n" +  // md5(params)
               Expires + "\n" +
               AccessKey;
*/
func genSignature(accessKey, secretKey, uri, method, urlParams, params, nowTS string) (string, error) {
	if params != "" {
		md5Ctx := md5.New()
		_, _ = io.WriteString(md5Ctx, params)
		params = fmt.Sprintf("%x", md5Ctx.Sum(nil))
	}

	strSign := method + "\n" + uri + "\n" + urlParams  + "\n" + params + "\n" + nowTS + "\n" + accessKey
	sign := hmacSHA1Encrypt(strSign, secretKey)
	return sign, nil
}

// hmacSHA1Encrypt encrypt the encryptText use encryptKey
func hmacSHA1Encrypt(encryptText, encryptKey string) string {
	key := []byte(encryptKey)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(encryptText))
	var str = hex.EncodeToString(mac.Sum(nil))
	return str
}

func getMapKeysSorted(originMap map[string]string) []string {
	keys := make([]string, len(originMap))
	i := 0
	for k := range originMap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}
