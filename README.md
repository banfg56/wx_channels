# 微信视频号后端接口模拟

视频号后端接口模拟，用于获取直播数据与回放地址下载


## 补充测试用例

> 需要设置环境变量 wxuin 与 sessionid 用于真实环境登录

### 1. 直播回放视频列表
go test -v -run TestGetLiveReplayPostList 

### 2. 直播记录
go test -v -run TestGetLiveReplayList

### 3. 直播单场数据
go test -v -run TestGetLiveHistory

## 4. 直播回放生成
objId=""  liveId="" go test -v -run TestLiveCreateReplay

## 5. 设置直播回放可见范围
objId=""   go test -v -run TestLiveUpdateVisible
