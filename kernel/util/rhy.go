// SiYuan - Refactor your thinking
// Copyright (c) 2020-present, b3log.org
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package util

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/imroc/req/v3"
	"github.com/siyuan-note/httpclient"
	"github.com/siyuan-note/logging"
)


var cachedBazaarResult = map[string]interface{}{}
var bazaarResultCacheTime int64

//var cachedReleaseResult = map[string]interface{}{}
//var releaseResultCacheTime int64

var aggregateResultCacheTime int64
var rhyResultLock = sync.Mutex{}
var cachedRhyResult = map[string]interface{}{}

func NewGitHubApiRequest() *req.Request {
	return httpclient.NewCloudRequest30s().SetHeader("Accept", "application/vnd.github+json").SetHeader("X-GitHub-Api-Version", "2022-11-28")
}

// Query GitHub Releases for the Client
//func GetReleaseResult(force bool, cacheDuration int64) (map[string]interface{}, error) {
//
//	now := time.Now().Unix()
//	if cacheDuration >= now - releaseResultCacheTime && !force && 0 < len(cachedReleaseResult) {
//		return cachedRhyResult, nil
//	}
//
//	request := NewGitHubApiRequest()
//	resp, err := request.SetSuccessResult(&cachedReleaseResult).Get(GetGitHubApiEndpoint() + "/repo/" + GetGitHubSiyuanRepo() + "/releases/tags/v" + Ver)
//	if err != nil {
//		logging.LogErrorf("GitHub Release: query failed: %s", err)
//		return nil, err
//	}
//	if 200 != resp.StatusCode {
//		msg := fmt.Sprintf("GitHub Release: query failed: %d", resp.StatusCode)
//		logging.LogErrorf(msg)
//		return nil, errors.New(msg)
//	}
//	releaseResultCacheTime = now
//	return cachedReleaseResult, nil
//}

// Query the GitHub Bazaar Repo's Main Branch
func GetBazaarResult(force bool, cacheDuration int64) (map[string]interface{}, error) {

	now := time.Now().Unix()
	if cacheDuration >= now - bazaarResultCacheTime && !force && 0 < len(cachedBazaarResult) {
		return cachedBazaarResult, nil
	}

	uri := GetGitHubApiEndpoint() + "/repos/" + GetGitHubBazaarRepo() + "/branches/" + GetGitHubBazaarBranch()
	request := NewGitHubApiRequest()
	resp, err := request.SetSuccessResult(&cachedBazaarResult).Get(uri)
	if err != nil {
		logging.LogErrorf("GitHub Bazaar: query failed: %s", err)
		return nil, err
	}
	if 200 != resp.StatusCode {
		msg := fmt.Sprintf("GitHub Bazaar: query failed with code: %d", resp.StatusCode)
		logging.LogErrorf(msg)
		return nil, errors.New(msg)
	}
	bazaarResultCacheTime = now
	return cachedBazaarResult, nil
}

// Function Replaced to Directly Query Github.
func GetRhyResult(force bool) (map[string]interface{}, error) {

	rhyResultLock.Lock()
	defer rhyResultLock.Unlock()

	cacheDuration := int64(3600 * 6)
	if ContainerDocker == Container {
		cacheDuration = int64(3600 * 24)
	}
	now := time.Now().Unix()
	if cacheDuration >= now - aggregateResultCacheTime && !force && 0 < len(cachedBazaarResult) {
		return cachedRhyResult, nil
	}

	//respRelease, errRelease := GetReleaseResult(force, cacheDuration)
	//if errRelease != nil {
	//	nil, errRelease
	//}

	respBazaar, errBazaar := GetBazaarResult(force, cacheDuration)
	if errBazaar != nil {
		return nil, errBazaar
	}

	cachedRhyResult = make(map[string]interface{})
	cachedRhyResult["bazaar"] = respBazaar["commit"].(map[string]interface{})["sha"]

	//aggregateResultCacheTime = min(releaseResultCacheTime, bazaarResultCacheTime)
	aggregateResultCacheTime = bazaarResultCacheTime
	return cachedRhyResult, nil
}

func RefreshRhyResultJob() {
	_, err := GetRhyResult(true)
	if nil != err {
		// 系统唤醒后可能还没有网络连接，这里等待后再重试
		go func() {
			time.Sleep(7 * time.Second)
			GetRhyResult(true)
		}()
	}
}
