package wxchannels

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
)

type CookieFilterInteface interface {
	Apply(*http.Cookie) *http.Cookie
}

func getLiveAdmin(t *testing.T) *WxChannelLiveAdmin {
	// 从环境变量获取 logincookie
	loginData := make(map[string]string)
	loginData["wxuin"] = os.Getenv("wxuin")
	loginData["sessionid"] = os.Getenv("sessionid")

	if loginData["wxuin"] == "" || loginData["sessionid"] == "" {
		t.Skip("wxuin 与sessionid  环境变量未设置，跳过测试")
		return nil
	}

	wx_channelsCookieLifeLeft := int64(20 * 60) // 20个小时

	// 解析 cookie 字符串
	cookies := parseJosnCookie(loginData, "channels.weixin.qq.com", &wx_channelsCookieLifeLeft, nil)
	if len(cookies) == 0 {
		t.Fatal("无法解析 cookie 字符串")
		return nil
	}
	// 创建 logger
	logger := log.NewStdLogger(os.Stdout)

	// 创建 WxChannelLiveAdmin 实例
	admin := NewWxChannelLiveAdmin(logger)

	// 设置登录信息
	err := admin.LoginAccount(cookies)
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}

	return admin
}

func TestGetLiveReplayPostList(t *testing.T) {
	// 创建 WxChannelLiveAdmin 实例
	admin := getLiveAdmin(t)

	// 创建请求参数
	req := NewReqLiveReplayPostList(1, 70)

	// 调用被测试函数
	posts, count, err := admin.GetLiveReplayPostList(req)

	// 验证结果
	assert.NoError(t, err, "GetLiveReplayPostList 应该不返回错误")
	assert.GreaterOrEqual(t, int(count), 0, "总数应该大于或等于0")

	// 如果有数据，验证第一条数据的格式
	if len(posts) > 0 {
		post := posts[0]
		assert.NotEmpty(t, post.ObjectId, "ObjectId 不应为空")
		assert.NotEmpty(t, post.LiveObjectId, "LiveObjectId 不应为空")
		assert.Greater(t, post.CreatetimeUnix, int64(0), "CreatetimeUnix 应大于0")
	}

	// 输出测试结果
	t.Logf("成功获取直播回放->直播回放视频列表，总数: %d, 当前页数量: %d", count, len(posts))
	for i, post := range posts {
		if i < 3 { // 只打印前3条记录
			t.Logf("回放视频 #%d: ObjectId=%s, LiveObjectId=%s, 创建时间=%s",
				i+1,
				post.ObjectId,
				post.LiveObjectId,
				time.Unix(post.CreatetimeUnix, 0).Format("2006-01-02 15:04:05"))
		}
	}
}

func TestGetLiveReplayList(t *testing.T) {
	// 创建 WxChannelLiveAdmin 实例
	admin := getLiveAdmin(t)

	// 创建请求参数
	req := NewReqLiveReplayList(1, 50)
	tmNow := time.Now()
	req.TimeFilter.Begin = fmt.Sprintf("%d", tmNow.Add(-time.Hour*24*30).Unix())
	req.TimeFilter.End = fmt.Sprintf("%d", tmNow.Unix())

	// 调用被测试函数
	posts, count, err := admin.GetLiveReplayList(req)

	// 验证结果
	assert.NoError(t, err, "GetLiveReplayList 应该不返回错误")
	assert.GreaterOrEqual(t, int(count), 0, "总数应该大于或等于0")

	// 如果有数据，验证第一条数据的格式
	if len(posts) > 0 {
		post := posts[0]
		assert.NotEmpty(t, post.ObjectId, "ObjectId 不应为空")
		assert.NotEmpty(t, post.LiveId, "LiveId 不应为空")
		assert.NotEmpty(t, post.StartTime, "直播开始时间不应为空")
	}

	// 输出测试结果
	t.Logf("成功获取直播回放->直播记录列表，总数: %d, 当前页数量: %d", count, len(posts))
	for i, post := range posts {
		if i < 3 { //
			tmFromStr, _ := strconv.ParseInt(post.StartTime, 10, 64)
			t.Logf("回放视频 #%d: ObjectId=%s, liveId=%s, 直播开始时间=%s replayStatus=%d",
				i+1,
				post.ObjectId,
				post.LiveId,
				time.Unix(tmFromStr, 0).Format("2006-01-02 15:04:05"), post.ReplayStatus)
		}
	}
}

func TestGetLiveHistory(t *testing.T) {
	// 创建 WxChannelLiveAdmin 实例
	admin := getLiveAdmin(t)

	// 创建请求参数
	req := NewReqLiveHistoryList(1, 10)

	tmNow := time.Date(2025, 03, 01, 0, 0, 0, 0, time.Local)
	req.UxTimesStart = tmNow.Unix()
	req.UxTimesEnd = tmNow.Add(time.Hour * 24 * 2).Unix()
	t.Logf("筛选 %s  至  %s 的直播数据", tmNow.Format("2006-01-02 15:04:05"), tmNow.Add(time.Hour*24*2).Format("2006-01-02 15:04:05"))

	// 调用被测试函数
	posts, count, err := admin.GetLiveHistory(req)

	// 验证结果
	assert.NoError(t, err, "GetLiveHistory 应该不返回错误")
	assert.GreaterOrEqual(t, int(count), 0, "总数应该大于或等于0")

	// 如果有数据，验证第一条数据的格式
	if len(posts) > 0 {
		post := posts[0]
		assert.NotEmpty(t, post.LiveObjectId, "ObjectId 不应为空")
		assert.NotNil(t, post.AuthorReplayStatus, "authorReplayStatus 不应为空")
		assert.Greater(t, post.UxCreateTime, int64(0), "UxCreateTime 应大于0")
	}

	// 输出测试结果
	t.Logf("成功获取直播数据->单场数据列表，总数: %d, 当前页数量: %d", count, len(posts))
	for i, post := range posts {
		if i < 3 { //
			t.Logf("回放视频 #%d: ObjectId=%s, authorReplayStatus=%d, 直播开始时间=%s  ",
				i+1,
				post.LiveObjectId,
				post.AuthorReplayStatus,
				time.Unix(post.UxCreateTime, 0).Format("2006-01-02 15:04:05"))
		}
	}
}

func TestLiveCreateReplay(t *testing.T) {
	// 创建 WxChannelLiveAdmin 实例
	admin := getLiveAdmin(t)
	objId := os.Getenv("objId")
	liveId := os.Getenv("liveId")
	if objId == "" {
		t.Fatalf("objId  要转换录播ID不能为空")
	}
	t.Logf("开始生成直播回放: ObjId:%s, liveId: %s", objId, liveId)
	// 调用被测试函数
	req := NewReqLiveSetReplay(liveId, objId)
	err := admin.LiveCreateReplay(req)
	assert.NoError(t, err, "LiveCreateReplay 应该不返回错误")
	t.Logf("成功生成直播回放:  ObjId:%s, liveId: %s", objId, liveId)
}

func TestLiveUpdateVisible(t *testing.T) {
	// 创建 WxChannelLiveAdmin 实例
	admin := getLiveAdmin(t)
	objId := os.Getenv("objId")
	if objId == "" {
		t.Fatalf("objId  导出任务ID不能为空")
	}
	t.Logf("开始设置回放可见范围: 导出任务ID:%s", objId)
	// 调用被测试函数
	req := NewReqLiveReplayPostUpdateVisible(objId)
	err := admin.LiveUpdateVisible(req)
	assert.NoError(t, err, "LiveUpdateVisible 应该不返回错误")
	t.Logf("成功设置回放可见范围:  ObjId:%s ", objId)
}

func parseJosnCookie(data map[string]string, domainName string, timeOutMinuts *int64, fi interface{}) (res []*http.Cookie) {
	for cK, cV := range data {
		vSaveCookie := &http.Cookie{Name: cK, Value: cV, Path: "/", Domain: domainName, Secure: true, HttpOnly: false, SameSite: http.SameSiteDefaultMode}
		vSaveCookie.Expires = time.Now().Add(time.Hour * 7 * 30)
		if timeOutMinuts != nil && *timeOutMinuts > 0 {
			vSaveCookie.Expires = time.Now().Add(time.Minute * time.Duration(*timeOutMinuts))
		}
		if fi != nil {
			switch fi.(type) {
			case CookieFilterInteface:
				vSaveCookie = fi.(CookieFilterInteface).Apply(vSaveCookie)
			}

		}
		res = append(res, vSaveCookie)
	} //for
	return
}
