package wxchannels

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

const (
	WX_VIDEO_ADMIN_HOST                  = `https://channels.weixin.qq.com`
	WX_VIDEO_ADMIN_URI_REPLAY_LIST       = `/cgi-bin/mmfinderassistant-bin/live/get_live_replay_list_v2`
	WX_VIDEO_ADMIN_URI_REPLAY_FRAGMENT   = `/cgi-bin/mmfinderassistant-bin/live/get_live_replay_wonderful_fragment`
	WX_VIDEO_ADMIN_URI_LIVE_HISTORY      = `/cgi-bin/mmfinderassistant-bin/live/get_live_history`
	WX_VIDEO_ADMIN_REPLAY_CREATE_SUCCESS = 3 //回访创建成功
	WX_ADMIN_SCENE_DEFAULT               = 7
	WX_ADMIN_REPLAY_INFO_FRAGMENT_DOWN   = 0 //所有视频分片都已处理成功
)

// 腾讯视频号助手管理端相关接口
type WxChannelLiveAdmin struct {
	_wxBase
	account WxLiveAccount //当前管理账号
}

type WxLiveAccount struct {
	Uid             string // finderUsername,长字符串
	Nickname        string
	Avatar          string
	XAuthHeaderUin  string //登录人微信账号ID
	UniqId          string //视频号ID,后台可见
	FansCount       int64  //粉丝数
	AuthCompanyName string //认证公司名
}

type ReqPage struct {
	Index int32 `json:"pageIndex"` //当前位置，从1开始
	Size  int32 `json:"pageSize"`  //分页大小
}

type ReqTimeFilter struct {
	Begin string `json:"startTimeBegin"` //都是timestamp
	End   string `json:"startTimeEnd"`
}
type _unUsedParam struct {
	PluginSessionId *string `json:"pluginSessionId"`
	RawKeyBuff      *string `json:"rawKeyBuff"`
}
type ReqLiveReplayList struct {
	_unUsedParam
	TimeFilter        ReqTimeFilter `json:"reqFilter"` //默认是0
	PageFilter        ReqPage       `json:"reqPage"`
	ReqScene          int32         `json:"reqScene"` // reqScene、scene 默认都是7
	Scene             int32         `json:"scene"`
	TimestampUixMilli string        `json:"timestamp"`
	// LogFinderId       string        `json:"_log_finder_id"`
}

type ReqLivePlayInfo struct {
	_unUsedParam
	ObjectId          string `json:"objectId"`
	ReqScene          int32  `json:"reqScene"` // reqScene、scene 默认都是7
	Scene             int32  `json:"scene"`
	TimestampUixMilli string `json:"timestamp"`
}

type ReqLiveHistoryList struct {
	_unUsedParam
	ReqScene          int32  `json:"reqScene"` // reqScene、scene 默认都是7
	Scene             int32  `json:"scene"`
	ReqType           int32  `json:"reqType"` //默认2
	UxTimesStart      int64  `json:"filterStartTime"`
	UxTimesEnd        int64  `json:"filterEndTime"`
	PageIndex         int32  `json:"currentPage"` //从1开始
	PageSize          int32  `json:"pageSize"`
	TimestampUixMilli string `json:"timestamp"`
}

type RespLivePlayInfo struct {
	CreatetimeUnix  int64  `json:"createtime"`
	Description     string `json:"description"`
	DurationSeconds int64  `json:"duration"`
	Height          int64  `json:"height"`
	Width           int64  `json:"width"`
	HlsUri          string `json:"hlsUrl"`
	ReplayUri       string `json:"replayUrl"`
	ThumbUri        string `json:"thumbUrl"` //封面
	FragmentStatus  int64  `json:"wonderfulFragmentStatus"`
}

type RespLiveCommon struct {
	PluginSessionId *string     `json:"pluginSessionId"`
	RawKeyBuff      *string     `json:"rawKeyBuff"`
	Code            int32       `json:"errCode"`
	Msg             string      `json:"errMsg"`
	Data            interface{} `json:"data"`
}

type RespLiveObjectMedia struct {
	Description    string `json:"description"`
	OriginCoverUrl string `json:"originCoverUrl"`
	ThumbUrl       string `json:"thumbUrl"`
}

type RespLiveObject struct {
	LiveId           string              `json:"liveId"`
	ChargeType       int32               `json:"chargeType"` //收费类型
	DurationSecs     int64               `json:"duration"`
	ObjectId         string              `json:"objectId"`
	Media            RespLiveObjectMedia `json:"media"`
	ReplayCreateTime string              `json:"replayCreateTime"` //回放创建时间
	ReplayStatus     int32               `json:"replayStatus"`     //回放状态
	StartTime        string              `json:"startTime"`        //开播时间
}

type RespLiveObjectStat struct { //直播统计数据
	LiveDurationInSeconds int64 `json:"liveDurationInSeconds"`
	TotalAudienceCount    int64 `json:"totalAudienceCount"`
}

type RespLiveHistoryItem struct {
	LiveObjectId   string             `json:"liveObjectId"`
	LiveData       RespLiveObjectStat `json:"liveStats"`
	UxCreateTime   int64              `json:"createTime"`
	MaxOnlineCount int64              `json:"maxOnlineCount"` //最大在线人数
}

type RespLiveReplayList struct {
	TotalCount *int32           `json:"totalCount"`
	LivePlays  []RespLiveObject `json:"replayObjects"`
}

type RespLiveHistoryList struct {
	TotalCount *int32                `json:"totalLiveCount"`
	Items      []RespLiveHistoryItem `json:"liveObjectList"`
}

func NewWxChannelLiveAdmin(logger log.Logger) *WxChannelLiveAdmin {
	_cookie, _ := cookiejar.New(nil)
	return &WxChannelLiveAdmin{
		_wxBase: _wxBase{
			AdminCookies: _cookie,
			AppLog:       log.NewHelper(logger),
		},
		account: WxLiveAccount{},
	}
}

func (d RespLiveCommon) UnMarshalData(out any) (err error) {
	_mResData, _dataJErr := json.Marshal(d.Data)
	if _dataJErr == nil {
		err = json.Unmarshal(_mResData, out)
	} else {
		err = _dataJErr
	}
	return
}

func (d WxLiveAccount) GetHeaderWechatUin() string {
	return d.XAuthHeaderUin
}

// 是否已准备好
func (d RespLivePlayInfo) IsReadyForDownload() bool {
	if len(d.ReplayUri) > 0 && d.FragmentStatus == WX_ADMIN_REPLAY_INFO_FRAGMENT_DOWN {
		return true
	}
	return false
}

func (d RespLivePlayInfo) GetQa() (qa string, err error) {
	qa = "none"
	if d.Width == 0 {
		err = fmt.Errorf("回放数据异常,宽为0")
	} else if d.Width > 1100 {
		qa = "uhd"
	} else if d.Width > 700 { // 1000的origin 去掉
		qa = "hd"
	} else if d.Width > 500 {
		qa = "sd"
	} else if d.Width > 400 {
		qa = "ld"
	} else if d.Width > 200 {
		qa = "md"
	} else {
		err = fmt.Errorf("回放数据异常，无法判断分辨率")
	}
	return
}

func NewReqLiveReplayList(pageIndex, pageSize int32) ReqLiveReplayList { //分页默认为6
	if pageSize <= 0 {
		pageSize = 6
	}
	if pageIndex <= 0 {
		pageIndex = 1
	}
	return ReqLiveReplayList{ReqScene: WX_ADMIN_SCENE_DEFAULT, Scene: WX_ADMIN_SCENE_DEFAULT, TimeFilter: ReqTimeFilter{"0", "0"}, PageFilter: ReqPage{Index: pageIndex, Size: pageSize}}
}

func NewReqLiveHistoryList(pageIndex, pageSize int32) ReqLiveHistoryList {
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageIndex <= 0 {
		pageIndex = 1
	}
	return ReqLiveHistoryList{ReqScene: WX_ADMIN_SCENE_DEFAULT, Scene: WX_ADMIN_SCENE_DEFAULT, ReqType: 2, PageIndex: pageIndex, PageSize: pageSize}
}

func NewReqLiveReplayInfo(id string) ReqLivePlayInfo {
	return ReqLivePlayInfo{ReqScene: WX_ADMIN_SCENE_DEFAULT, Scene: WX_ADMIN_SCENE_DEFAULT, ObjectId: id}
}

func (r *WxChannelLiveAdmin) LoginAccount(cs []*http.Cookie) (err error) {
	//account用户信息直接通过接口获取
	err = r.SetAdminCookie(cs)
	// 获取当前人员登录信息 /cgi-bin/mmfinderassistant-bin/auth/auth_data
	_header := map[string]string{ // Set multiple headers at once
		"Referer":      "https://channels.weixin.qq.com/platform/live/liveReplayHistory",
		"X-Wechat-Uin": "0000000000",
	}

	type _reqAuthData struct {
		_unUsedParam
		ReqScene          int32  `json:"reqScene"` // reqScene、scene 默认都是7
		Scene             int32  `json:"scene"`
		TimestampUixMilli string `json:"timestamp"`
	}
	type _respAuthData struct {
		FindUser struct {
			NickName        string `json:"nickname"`        // 昵称
			Avatar          string `json:"headImgUrl"`      // 头像
			UidStr          string `json:"uniqId"`          // 视频号ID
			FinderUsername  string `json:"finderUserName"`  // 微信号
			FansCount       int32  `json:"fansCount"`       // 粉丝数
			FeedsCount      int32  `json:"feedsCount"`      // 动态数
			AuthCompanyName string `json:"authCompanyName"` // 公司名称
		} `json:"finderUser"`
	}
	reqAuth := _reqAuthData{TimestampUixMilli: fmt.Sprintf("%d", time.Now().UnixMilli()), ReqScene: WX_ADMIN_SCENE_DEFAULT, Scene: WX_ADMIN_SCENE_DEFAULT}
	var jAuth RespLiveCommon
	resp, respErr := r.GetReqClient(_header).R().SetBody(&reqAuth).
		SetSuccessResult(&jAuth).
		Post(fmt.Sprintf("%s%s", WX_VIDEO_ADMIN_HOST, `/cgi-bin/mmfinderassistant-bin/auth/auth_data`))
	if respErr != nil {
		err = fmt.Errorf("cookie请求authData异常,%s, 错误数据:%s ", respErr.Error(), getStrMax(resp.String(), 100))
		return
	}

	var authResp _respAuthData
	isRespOk := jAuth.UnMarshalData(&authResp)
	if isRespOk != nil {
		err = fmt.Errorf("类型转换失败,%s", isRespOk.Error())
		return
	} else if jAuth.Code != 0 {
		err = fmt.Errorf("获取授权数据失败,code:%d msg:%s", jAuth.Code, jAuth.Msg)
		return
	}
	r.account.Nickname = authResp.FindUser.NickName
	r.account.Avatar = authResp.FindUser.Avatar
	r.account.Uid = authResp.FindUser.FinderUsername
	r.account.UniqId = authResp.FindUser.UidStr
	r.account.FansCount = int64(authResp.FindUser.FansCount)
	r.account.AuthCompanyName = authResp.FindUser.AuthCompanyName

	// 获取uin /cgi-bin/mmfinderassistant-bin/helper/helper_upload_params
	reqAuth.TimestampUixMilli = fmt.Sprintf("%d", time.Now().UnixMilli())
	type _respHeaderParam struct {
		AppType int32  `json:"appType"`
		AuthKey string `json:"authKey"`
		Uin     int64  `json:"uin"`
	}

	var helperResp _respHeaderParam
	resp, respErr = r.GetReqClient(_header).R().SetBody(&reqAuth).
		SetSuccessResult(&jAuth).
		Post(fmt.Sprintf("%s%s", WX_VIDEO_ADMIN_HOST, `/cgi-bin/mmfinderassistant-bin/helper/helper_upload_params`))
	if respErr != nil {
		err = fmt.Errorf("cookie请求helper_upload_params异常,%s, 错误数据:%s ", respErr.Error(), getStrMax(resp.String(), 100))
		return
	} else if jAuth.Code != 0 {
		err = fmt.Errorf("获取视频号ID异常,code:%d msg:%s", jAuth.Code, jAuth.Msg)
		return
	}

	isRespOk = jAuth.UnMarshalData(&helperResp)
	if isRespOk != nil {
		err = fmt.Errorf("类型转换失败,%s", isRespOk.Error())
		return
	}
	r.account.XAuthHeaderUin = fmt.Sprintf("%d", helperResp.Uin) // 当前登录人的微信号标记，固定
	return
}

func (r *WxChannelLiveAdmin) GetWxChannelAccount() WxLiveAccount {
	return r.account
}

// 获取直播回放列表
func (r *WxChannelLiveAdmin) GetLiveReplayList(req ReqLiveReplayList) (livePlays []RespLiveObject, allCount int32, err error) {
	livePlays = make([]RespLiveObject, 0)
	_header := map[string]string{ // Set multiple headers at once
		"Referer":      "https://channels.weixin.qq.com/platform/live/liveReplayHistory",
		"X-Wechat-Uin": r.account.GetHeaderWechatUin(),
	}
	c := r.GetReqClient(_header)
	req.TimestampUixMilli = fmt.Sprintf("%d", time.Now().UnixMilli())
	var jres RespLiveCommon
	resp, respErr := c.R().SetBody(&req).
		SetSuccessResult(&jres).
		Post(fmt.Sprintf("%s%s", WX_VIDEO_ADMIN_HOST, WX_VIDEO_ADMIN_URI_REPLAY_LIST))

	if respErr != nil {
		err = fmt.Errorf("请求异常,%s, 错误数据:%s ", respErr, getStrMax(resp.String(), 100))
	} else {
		if jres.Code == 0 {
			var playResp RespLiveReplayList
			isRespOk := jres.UnMarshalData(&playResp)
			if isRespOk != nil {
				err = fmt.Errorf("类型转换失败,%s", isRespOk.Error())
			} else {
				allCount = 0
				if playResp.TotalCount != nil {
					allCount = *playResp.TotalCount
				}
				livePlays = playResp.LivePlays
			}
		} else {
			err = fmt.Errorf("请求异常,code:%d msg:%s", jres.Code, jres.Msg)
		}
	}

	return
}

// 获取回放下载地址
func (r *WxChannelLiveAdmin) GetLiveReplayInfo(req ReqLivePlayInfo) (info RespLivePlayInfo, err error) {
	_header := map[string]string{ // Set multiple headers at once
		"Referer":      "https://channels.weixin.qq.com/platform/live/replayDetail?liveObjectId=" + req.ObjectId,
		"X-Wechat-Uin": r.account.GetHeaderWechatUin(),
	}
	c := r.GetReqClient(_header)
	req.TimestampUixMilli = fmt.Sprintf("%d", time.Now().UnixMilli())
	var jres RespLiveCommon
	resp, respErr := c.R().SetBody(&req).
		SetSuccessResult(&jres).
		Post(fmt.Sprintf("%s%s", WX_VIDEO_ADMIN_HOST, WX_VIDEO_ADMIN_URI_REPLAY_FRAGMENT))

	if respErr != nil {
		err = fmt.Errorf("请求异常,%s, 错误数据:%s ", respErr, getStrMax(resp.String(), 100))

	} else {
		if jres.Code == 0 {
			var playResp RespLivePlayInfo
			isRespOk := jres.UnMarshalData(&playResp)
			if isRespOk != nil {
				err = fmt.Errorf("类型转换失败,%s", isRespOk.Error())
			} else {
				info = playResp
				//增加校验
				if len(info.ReplayUri) == 0 && len(info.HlsUri) == 0 {
					err = fmt.Errorf("回放片段为空")
				}
			}
		} else {
			err = fmt.Errorf("请求异常,code:%d msg:%s", jres.Code, jres.Msg)
		}
	}
	return
}

func (r *WxChannelLiveAdmin) GetLiveHistory(req ReqLiveHistoryList) (res []RespLiveHistoryItem, allCount int32, err error) {
	allCount = 0
	res = make([]RespLiveHistoryItem, 0)
	_header := map[string]string{ // Set multiple headers at once
		"Referer":      "https://channels.weixin.qq.com/platform/statistic/live?mode=history",
		"X-Wechat-Uin": r.account.GetHeaderWechatUin(),
	}
	c := r.GetReqClient(_header)
	req.TimestampUixMilli = fmt.Sprintf("%d", time.Now().UnixMilli())
	var jres RespLiveCommon
	resp, respErr := c.R().SetBody(&req).
		SetSuccessResult(&jres).
		Post(fmt.Sprintf("%s%s", WX_VIDEO_ADMIN_HOST, WX_VIDEO_ADMIN_URI_LIVE_HISTORY))

	if respErr != nil {
		err = fmt.Errorf("请求异常,%s, 错误数据:%s ", respErr, getStrMax(resp.String(), 100))

	} else {
		if jres.Code == 0 {
			var historyList RespLiveHistoryList
			isRespOk := jres.UnMarshalData(&historyList)
			if isRespOk != nil {
				err = fmt.Errorf("类型转换失败,%s", isRespOk.Error())
			} else {
				allCount = 0
				if historyList.TotalCount != nil {
					allCount = *historyList.TotalCount
				}
				res = historyList.Items
			}
		} else {
			err = fmt.Errorf("请求异常,code:%d msg:%s", jres.Code, jres.Msg)
		}
	}
	return
}
