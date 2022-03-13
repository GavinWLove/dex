
package wecom

import (
	"encoding/json"
	"fmt"
	"github.com/dexidp/dex/connector"
	"github.com/dexidp/dex/pkg/log"
	"io"
	"net/http"
	"net/url"
)


type Config struct {
	CorpId      string `json:"corpId"`
	CorpSecret  string `json:"corpSecret"`
	AgentId  string `json:"agentId"`
	RedirectURI   string `json:"redirectURI"`
	AccessTokenURI           string   `json:"accessTokenURI"`
	UserIdURI   string   `json:"userIdURI"`
	UserInfoURI        string   `json:"userInfoURI"`
	QrConnectURI        string   `json:"qrConnectURI"`
}


func (c *Config) Open(id string, logger log.Logger) (connector.Connector, error) {
	g := wecomConnector{
		logger:       logger,
		RedirectURI:   c.RedirectURI,
		AccessTokenURI:    c.AccessTokenURI,
		UserIdURI:   c.UserIdURI,
		UserInfoURI:        c.UserInfoURI,
		QrConnectURI:        c.QrConnectURI,
		CorpId:      c.CorpId,
		CorpSecret:  c.CorpSecret,
		AgentId:  c.AgentId,
	}
	return &g, nil
}

type connectorData struct {
	AccessToken string `json:"accessToken"`
}

var (
	_ connector.CallbackConnector = (*wecomConnector)(nil)
)

type wecomConnector struct {
	logger       log.Logger
	RedirectURI   string
	AccessTokenURI    string
	UserIdURI   string
	UserInfoURI        string
	QrConnectURI        string
	CorpId      string
	CorpSecret  string
	AgentId  string
}

func (c *wecomConnector) LoginURL(scopes connector.Scopes, callbackURL, state string) (string, error) {
	u, _ := url.Parse(c.QrConnectURI)
	q := u.Query()
	q.Set("appid", c.CorpId)
	q.Set("agentid", c.AgentId)
	q.Set("state", state)
	q.Set("redirect_uri", c.RedirectURI)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

type oauth2Error struct {
	error            string
	errorDescription string
}

func (e *oauth2Error) Error() string {
	if e.errorDescription == "" {
		return e.error
	}
	return e.error + ": " + e.errorDescription
}


func (c *wecomConnector) HandleCallback(s connector.Scopes, r *http.Request) (identity connector.Identity, err error) {
	q := r.URL.Query()
//扫描后跳转到：(获取code)
//http://cmdp.dev.jingkunsystem.com:19041/?code=Io1SvwImGcPPjdivyXOq-SveKkhI6umxi_Zkc2thbGM&state=weChat&appid=wx227c3136c137e769
	code := q.Get("code")
	accessResult,_ := c.accessToken()

	userId, _ := c.getUserId(accessResult.AccessToken,code)
	userInfo,_ :=c.getUser(accessResult.AccessToken,userId.UserId)

	if err != nil {
		return identity, fmt.Errorf("github: get user: %v", err)
	}

//TODO 添加 手机号
	identity = connector.Identity{
		UserID:            userInfo.Userid,
		Username:          userInfo.Name,
		PreferredUsername: userInfo.Name,
		Email:             userInfo.BizMail,
		EmailVerified:     true,
		PhoneNumber: userInfo.Mobile,
	}

	if s.OfflineAccess {
		data := connectorData{AccessToken: accessResult.AccessToken}
		connData, err := json.Marshal(data)
		if err != nil {
			return identity, fmt.Errorf("marshal connector data: %v", err)
		}
		identity.ConnectorData = connData
	}

	return identity, nil
}

type tokenResult struct {
	AccessToken string `json:"access_token"`
	ExpiresIn    int64    `json:"expires_in"`
	ErrCode    int    `json:"errcode"`
	ErrMsg string `json:"errmsg"`
}
//获取access_token:
//https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=wx227c3136c137e769&corpsecret=mImrlXelFEj9kwxzUdb4lsMiAfN4zJpp7PQEqenH6ng
//	_rvAVvneNbURn9dbVAMJ14z4gVj-A0_pdXY72v5JriXbZMWC9m67QIV4Ysj5O94-1yI9hPXO62jt_GjOk5hi4XEH4paadWOFBe9TtzNVljDy_bAyxZb_o39J7_sNdsjmIuwLfmUB1YQ6qEc20PQ6cyQyjo3wDQTpvHy5kxxvqBhfSthEbbtApeKZlPX0GH1qMZWm-ogDItzGusD4cZAZeQ
//	{"errcode":0,"errmsg":"ok",
//	"access_token":"_rvAVvneNbURn9dUB1YQ6qEc20PQ6cyQyjo3wDQTpvHy5kxxvqBhfSthEbbtApeKZlPX0GH1qMZWm-ogDItzGusD4cZAZeQ",
//	"expires_in":7200}
func (c *wecomConnector) accessToken() (tokenResult, error) {
	var u tokenResult
	resp, err := http.Get(c.AccessTokenURI+"?corpid="+c.CorpId+"&corpsecret="+c.CorpSecret)
	if err != nil {
		return u, fmt.Errorf("gitlab: get URL %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return u, fmt.Errorf("gitlab: read body: %v", err)
		}
		return u, fmt.Errorf("%s: %s", resp.Status, body)
	}
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return u, fmt.Errorf("failed to decode response: %v", err)
	}
	return u, nil
}
type userIdResult struct {
	UserId  string `json:"UserId"`
	DeviceId string `json:"DeviceId"`
	ErrCode    int    `json:"errcode"`
	ErrMsg string `json:"errmsg"`
}

//获取用户userId:
//https://qyapi.weixin.qq.com/cgi-bin/user/getuserinfo?access_token=_rvAVvneNbURn9dbVAMJ14z4gVj-A0_pdXY72v5JriXbZMWC9m67QIV4Ysj5O94-1yI9hPXO62jt_GjOk5hi4XEH4paadWOFBe9TtzNVljDy_bAyxZb_o39J7_sNdsjmIuwLfmUB1YQ6qEc20PQ6cyQyjo3wDQTpvHy5kxxvqBhfSthEbbtApeKZlPX0GH1qMZWm-ogDItzGusD4cZAZeQ&code=Io1SvwImGcPPjdivyXOq-SveKkhI6umxi_Zkc2thbGM
//{"UserId":"wangwei","DeviceId":"","errcode":0,"errmsg":"ok"}
func (c *wecomConnector) getUserId(accessToken, code string) (userIdResult, error) {
	var u userIdResult
	resp, err := http.Get(c.UserIdURI+"?access_token="+accessToken+"&code="+code)
	if err != nil {
		return u, fmt.Errorf("gitlab: get URL %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return u, fmt.Errorf("gitlab: read body: %v", err)
		}
		return u, fmt.Errorf("%s: %s", resp.Status, body)
	}
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return u, fmt.Errorf("failed to decode response: %v", err)
	}
	return u, nil
}

type user struct {
	Userid  string `json:"userid"`
	Name  string `json:"name"`
	Department  string `json:"department"`
	Position  string `json:"position"`
	Mobile  string `json:"mobile"`
	Gender  string `json:"gender"`
	Email  string `json:"email"`
	Avatar  string `json:"avatar"`
	Status  string `json:"status"`
	Isleader  string `json:"isleader"`
	Extattr  string `json:"extattr"`
	EnglishName  string `json:"english_name"`
	Telephone  string `json:"telephone"`
	Enable  string `json:"enable"`
	MainDepartment  string `json:"main_department"`
	Alias  string `json:"alias"`
	BizMail  string `json:"biz_mail"`
	ErrCode    int    `json:"errcode"`
	ErrMsg string `json:"errmsg"`
}
//获取用户详细信息:
//https://qyapi.weixin.qq.com/cgi-bin//user/get?access_token=_rvAVvneNbURn9dbVAMJ14z4gVj-A0_pdXY72v5JriXbZMWC9m67QIV4Ysj5O94-1yI9hPXO62jt_GjOk5hi4XEH4paadWOFBe9TtzNVljDy_bAyxZb_o39J7_sNdsjmIuwLfmUB1YQ6qEc20PQ6cyQyjo3wDQTpvHy5kxxvqBhfSthEbbtApeKZlPX0GH1qMZWm-ogDItzGusD4cZAZeQ&userid=wangwei
//{"errcode":0,"errmsg":"ok","userid":"wangwei","name":"王威","department":[28],"position":"",
//"mobile":"15618388792","gender":"1","email":"",
//"avatar":"https://wework.qpic.cn/bizmail/iaczq0FMRCKLoDuCibz3edzTeP9YcrZ9m7D5QewPFgBOUuVyXjfoJPkg/0",
//"status":1,"isleader":0,"extattr":{"attrs":[]},
//"english_name":"","telephone":"","enable":1,"hide_mobile":0,"order":[0],"main_department":28,
//"qr_code":"https://open.work.weixin.qq.com/wwopen/userQRCode?vcode=vcb4681607d06129da",
//"alias":"","is_leader_in_dept":[0],"thumb_avatar":"https://wework.qpic.cn/bizmail/iaczq0FMRCKLoDuCibz3edzTeP9YcrZ9m7D5QewPFgBOUuVyXjfoJPkg/100",
//"direct_leader":[],"biz_mail":"wangwei@jk111.wecom.work"}
func (c *wecomConnector) getUser(accessToken, userid string) (user, error) {
	var u user
	resp, err := http.Get(c.UserInfoURI+"?access_token="+accessToken+"&userid="+userid)
	if err != nil {
		return u, fmt.Errorf("gitlab: get URL %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return u, fmt.Errorf("gitlab: read body: %v", err)
		}
		return u, fmt.Errorf("%s: %s", resp.Status, body)
	}
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return u, fmt.Errorf("failed to decode response: %v", err)
	}
	return u, nil
}
