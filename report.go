package main

import (
	"encoding/json"
	"fmt"
	"github.com/xionghengheng/ff_plib/comm"
	"github.com/xionghengheng/ff_plib/db/dao"
	"github.com/xionghengheng/ff_plib/db/model"
	"net/http"
	"time"
)

type ReportReq struct {
	Events []Event `json:"events"`
}

type ReportRsp struct {
	Code     int    `json:"code"`
	ErrorMsg string `json:"errorMsg,omitempty"`
}

// 定义结构体
type Event struct {
	ActionID   int    `json:"actionId"` // 动作id，比如101代表曝光、102代表点击
	IsCoach    bool   `json:"isCoach"`
	SessionID  string `json:"sessionId"` //访问id，uid + 用户进入小程序的时间戳，确保唯一性
	ItemID     string `json:"itemId"`    //按钮id，每个点击对象分配一个，比如buy_vip
	ModuleID   string `json:"moduleId"`  //模块id，每个模块分配一个，比如course_info
	PageID     string `json:"pageId"`    //页面id，每个页面分配一个，比如home
	Model      string `json:"model"`
	AppID      string `json:"appId"`    // 应用id，全局唯一，比如funcoach
	Duration   int    `json:"duration"` //当前在小程序的停留时长，以用户进入小程序为基点计算
	BusiInfo   string `json:"busiInfo"` //额外信息，json字符串，比如一些非通用的附带状态等信息可以放到这里
	Brand      string `json:"brand"`
	EnvVersion string `json:"envVersion"`
	Platform   string `json:"platform"`
	System     string `json:"system"`
	Version    string `json:"version"`
	StrExt1    string `json:"str_ext1"`
	StrExt2    string `json:"str_ext2"`
	StrExt3    string `json:"str_ext3"`
	Ext1       int    `json:"ext1"`
	Ext2       int    `json:"ext2"`
	Ext3       int    `json:"ext3"`
}

func getReportReq(r *http.Request) (ReportReq, error) {
	req := ReportReq{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, err
	}
	defer r.Body.Close()
	return req, nil
}

// UpdateUserInfoHandler 统一上报接口
func Report(w http.ResponseWriter, r *http.Request) {
	strOpenId := r.Header.Get("X-WX-OPENID")
	req, err := getReportReq(r)
	rsp := &ReportRsp{}
	Printf("req start, strOpenId:%s req:%+v\n", strOpenId, req)
	defer func() {
		msg, err := json.Marshal(rsp)
		if err != nil {
			fmt.Fprint(w, "内部错误")
			return
		}
		w.Header().Set("content-type", "application/json")
		w.Write(msg)
	}()
	if err != nil || len(strOpenId) == 0 {
		rsp.Code = -999
		rsp.ErrorMsg = err.Error()
		return
	}

	if len(req.Events) == 0 || len(req.Events) >= 50 {
		rsp.Code = -10003
		rsp.ErrorMsg = "参数错误"
		return
	}
	for _,v := range req.Events{
		if v.ActionID >= 103 || v.ActionID <= 100{
			rsp.Code = -10003
			rsp.ErrorMsg = "ActionID 参数错误"
			return
		}
	}

	stUserInfoModel, err := comm.GetUserInfoByOpenId(strOpenId)
	Printf("getUserInfoByOpenId succ, strOpenId:%s stUserInfoModel:%+v err:%+v\n", strOpenId, stUserInfoModel, err)
	if err != nil {
		rsp.Code = -1
		rsp.ErrorMsg = err.Error()
		return
	}

	if stUserInfoModel.UserID == 0 {
		rsp.Code = -1
		rsp.ErrorMsg = "user not exist"
		return
	}


	for _,v := range req.Events{
		go func(stEvent Event, uid int64) {
			err = dao.ImpReport.DoReport(tranReportItem2DbItem(stEvent, uid))
			if err != nil {
				rsp.Code = -33
				rsp.ErrorMsg = err.Error()
				Printf("DoReport err, strOpenId:%s stUserInfoModel:%+v err:%+v\n", uid, stUserInfoModel, err)
				return
			}
		}(v, stUserInfoModel.UserID)
	}
	return
}

func tranReportItem2DbItem(event Event, uid int64)model.ReportModel{
	return model.ReportModel{
		UID:        uid,
		ReportTime: time.Now().Unix(),
		IsCoach:    event.IsCoach,
		SessionID:  event.SessionID,
		ItemID:     event.ItemID,
		ModuleID:   event.ModuleID,
		PageID:     event.PageID,
		Model:      event.Model,
		ActionID:   event.ActionID,
		AppID:      event.AppID,
		Duration:   event.Duration,
		BusiInfo:   event.BusiInfo,
		Brand:      event.Brand,
		EnvVersion: event.EnvVersion,
		Platform:   event.Platform,
		System:     event.System,
		Version:    event.Version,
		StrExt1:    event.StrExt1,
		StrExt2:    event.StrExt2,
		StrExt3:    event.StrExt3,
		Ext1:       event.Ext1,
		Ext2:       event.Ext2,
		Ext3:       event.Ext3,
	}
}
