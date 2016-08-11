package gogoanimeAssistant

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"animedom.com/common"
)

type streamMoeUploader struct {
	username, password string
	loggedInCookies    string
	csaKey1, csaKey2   string
}

func (v *streamMoeUploader) login() {
	form := url.Values{}
	form.Add("username", "kryptodev")
	form.Add("password", "eDt2avW8OFRvKvbjMeg2")

	req, err := http.NewRequest("POST", "https://stream.moe/ajax/_account_login.ajax.php", strings.NewReader(form.Encode()))
	common.CheckErrorAndPanic(err)

	req.Header.Set("origin", "https://stream.moe")
	req.Header.Set("referer", "https://stream.moe/login.html")
	req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/52.0.2743.116 Safari/537.36")
	req.Header.Set("x-requested-with", "XMLHttpRequest")
	req.Header.Set("content-type", "application/x-www-form-urlencoded; charset=UTF-8")

	client := http.Client{}
	resp, err := client.Do(req)
	common.CheckErrorAndPanic(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	common.CheckErrorAndPanic(err)

	//reply := string(body)

	var bodyResponse map[string]interface{}

	err = json.Unmarshal(body, &bodyResponse)
	common.CheckErrorAndPanic(err)

	interfaceLoginStatus := bodyResponse["login_status"].(string)
	if interfaceLoginStatus != "success" {
		panic(string(body))
	}

	var loggedInCookies string
	for _, v := range resp.Header["Set-Cookie"] {
		loggedInCookies += v + " "
	}
	loggedInCookies = strings.Replace(loggedInCookies, "HttpOnly", "", -1)
	v.loggedInCookies = strings.TrimSpace(loggedInCookies)
}

func (v *streamMoeUploader) getCSAKeys() {
	req, err := http.NewRequest("GET", "https://stream.moe/account_home.html", nil)
	common.CheckErrorAndPanic(err)

	req.Header.Set("referer", "https://stream.moe/login.html")
	req.Header.Set("upgrade-insecure-requests", "1")
	req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/52.0.2743.116 Safari/537.36")
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("dnt", "1")
	req.Header.Set("accept-language", "en-GB,en-US;q=0.8,en;q=0.6")

	req.Header.Set("cookie", v.loggedInCookies)

	client := http.Client{}
	resp, err := client.Do(req)
	common.CheckErrorAndPanic(err)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	common.CheckErrorAndPanic(err)

	reply := string(body)

	reply = reply[strings.Index(reply, "csaKey=")+10:]

	csaKey1Index := strings.Index(reply, "csaKey1=")
	var key1 string
	for i := csaKey1Index; i < len(reply); i++ {
		if reply[i] == '&' {
			key1 = reply[csaKey1Index+len("csaKey1=") : i]
			break
		}
	}

	csaKey2Index := strings.Index(reply, "csaKey2=")
	var key2 string
	for i := csaKey2Index; i < len(reply); i++ {
		if reply[i] == '\'' {
			key2 = reply[csaKey2Index+len("csaKey2=") : i]
			break
		}
	}
	v.csaKey1 = key1
	v.csaKey2 = key2
}

func (v *streamMoeUploader) fileList() {
	form := url.Values{}
	form.Add("nodeId", "-1")
	form.Add("filterText", "")
	form.Add("filterUploadedDateRange", "")
	form.Add("filterOrderBy", "order_by_filename_asc")
	form.Add("pageStart", "0")
	form.Add("perPage", "100")

	req, err := http.NewRequest("POST", "https://stream.moe/ajax/_account_home_v2_file_listing.ajax.php", strings.NewReader(form.Encode()))
	common.CheckErrorAndPanic(err)

	req.Header.Set("accept", "text/html, */*; q=0.01")
	req.Header.Set("accept-language", "en-GB,en-US;q=0.8,en;q=0.6")
	req.Header.Set("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("dnt", "1")
	req.Header.Set("origin", "https://stream.moe")
	req.Header.Set("referer", "https://stream.moe/account_home.html")
	req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/52.0.2743.116 Safari/537.36")
	req.Header.Set("x-requested-with", "XMLHttpRequest")

	req.Header.Set("cookie", v.loggedInCookies)

	client := http.Client{}
	resp, err := client.Do(req)
	common.CheckErrorAndPanic(err)

	defer resp.Body.Close()

	//body, err := ioutil.ReadAll(resp.Body)
	//common.CheckErrorAndPanic(err)

	//	reply := string(body)

	//fmt.Println("BODY", reply)
	//fmt.Println("STATUS CODE", resp.StatusCode)
	//fmt.Println("COOKIES", resp.Cookies())
	//fmt.Println("OURCOOKIES", v.loggedInCookies)
	//fmt.Println("HeaderCookies", resp.Header["Set-Cookie"])
	//fmt.Println("HEADER", resp.Header)

}

func (v *streamMoeUploader) upload(remoteFile string) (shortUrl, fileID string) {
	req, err := http.NewRequest("GET", "https://sillysergal.stream.moe/core/page/ajax/url_upload_handler.ajax.php", nil)
	common.CheckErrorAndPanic(err)

	form := req.URL.Query()
	form.Add("csaKey1", v.csaKey1)
	form.Add("csaKey2", v.csaKey2)
	form.Add("rowId", "0")
	form.Add("url", remoteFile)
	form.Add("folderId", "null")

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-GB,en-US;q=0.8,en;q=0.6")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", v.loggedInCookies)
	req.Header.Set("DNT", "1")
	req.Header.Set("Host", "sillysergal.stream.moe")
	req.Header.Set("Referer", "https://stream.moe/account_home.html")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/52.0.2743.116 Safari/537.36")

	req.URL.RawQuery = form.Encode()

	client := http.Client{}

	resp, err := client.Do(req)
	common.CheckErrorAndPanic(err)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	common.CheckErrorAndPanic(err)

	shortUrl = string(body)
	{
		idIndex := strings.Index(shortUrl, "short_url\":\"") + len("short_url\":\"")

		for i := idIndex; i < len(shortUrl); i++ {
			if shortUrl[i] == '"' {
				shortUrl = shortUrl[idIndex:i]
				break
			}
		}
	}

	fileID = string(body)
	{
		idIndex := strings.Index(fileID, "file_id\":\"") + len("file_id\":\"")

		for i := idIndex; i < len(fileID); i++ {
			if fileID[i] == '"' {
				fileID = fileID[idIndex:i]
				break
			}
		}
	}

	return shortUrl, fileID
}

func (v *streamMoeUploader) renameFile(fileID, newName string) {
	form := url.Values{}
	form.Add("filename", newName)
	form.Add("folder", "")
	form.Add("password", "")
	form.Add("reset_stats", "0")
	form.Add("isPublic", "1")
	form.Add("submitme", "1")
	form.Add("fileId", fileID)

	req, err := http.NewRequest("POST", "https://stream.moe/ajax/_account_edit_file.process.ajax.php", strings.NewReader(form.Encode()))
	common.CheckErrorAndPanic(err)

	req.Header.Set("accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("accept-language", "en-GB,en-US;q=0.8,en;q=0.6")
	req.Header.Set("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("dnt", "1")
	req.Header.Set("origin", "https://stream.moe")
	req.Header.Set("referer", "https://stream.moe/account_home.html")
	req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/52.0.2743.116 Safari/537.36")
	req.Header.Set("x-requested-with", "XMLHttpRequest")

	req.Header.Set("cookie", v.loggedInCookies)

	client := http.Client{}
	resp, err := client.Do(req)
	common.CheckErrorAndPanic(err)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	common.CheckErrorAndPanic(err)

	var bodyResponse map[string]interface{}

	err = json.Unmarshal(body, &bodyResponse)
	common.CheckErrorAndPanic(err)

	interfaceLoginStatus := bodyResponse["msg"].(string)
	if interfaceLoginStatus != "File updated." {
		panic(string(body))
	}
}
