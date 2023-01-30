package nodeops

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	result NodeList
	// nodes  []int //= make([]string, 0)
)

type NodeList []struct {
	ID        int    `json:"id"`
	URL       string `json:"url"`
	Active    bool   `json:"active"`
	NodeTag   string `json:"nodeTag"`
	MetricUrl string `json:"metricUrl"`
}

func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

// Get nodes ID.
func GetNodes(standNumber int) (error, NodeList) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	var (
		n         string = strconv.Itoa(standNumber)
		urlString string = "https://node-manager-ift-" + n + ".apps.songd.sberdevices.ru/admin/node/list"
	)
	// nodes = make([]int, 0)
	resp, err := http.Get(urlString)
	if err != nil {
		return err, nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
	PrettyPrint(result)

	return nil, result
}

// func returnIDlist(standNuber int) []string{
// 	nodesId:=make([]string,0)
// 	_,result:= GetNodes(standNuber)
// 	for _,rec:=range result{
// 		nodesId = append(nodesId, string(rec.ID))
// 	}
// 	return nodesId
// }

// func returnNodeList(standNuber int) []string{
// 	nodeList:=make([]string,0)
// 	_,result:= GetNodes(standNuber)
// 	for _,rec:=range result{
// 		nodeList = append(nodeList, string(rec.ID))
// 	}
// 	return nodeList
// }

func ReturnNodeInfoMap(standNuber int, field string) map[int]string {
	nodeList := make(map[int]string, 0)
	_, result := GetNodes(standNuber)
	if field == "URL" {
		for _, rec := range result {
			nodeList[rec.ID] = rec.URL
		}
	} else if field == "ID" {
		for _, rec := range result {
			nodeList[rec.ID] = strconv.Itoa(rec.ID)
		}
	} else if field == "Active" {
		for _, rec := range result {
			if rec.Active {
				nodeList[rec.ID] = "true"
			} else {
				nodeList[rec.ID] = "false"
			}
		}
	} else if field == "NodeTag" {
		for _, rec := range result {
			nodeList[rec.ID] = rec.NodeTag
		}
	} else if field == "MetricUrl" {
		for _, rec := range result {
			nodeList[rec.ID] = rec.MetricUrl
		}
	} else {
		log.Fatal("No such field in NodeManager DB")
	}
	return nodeList
}

func metricUrl(standNumber string) string {
	var metricUrl string
	if standNumber == "1" {
		metricUrl = "http://d-jazz-bridge-sc-msk01.sberdevices.ru:8040/metrics"
	} else if standNumber == "2" {
		metricUrl = "http://d-jazz-bridge-sc-msk02.sberdevices.ru:8040/metrics"
	} else if standNumber == "3" {
		metricUrl = "http://d-jazz-bridge-sc-msk03.sberdevices.ru:8040/metrics"
	} else {
		log.Fatal("wrong stand Number")
	}
	return metricUrl
}

func addNode(standNumber, nodeName, nodeTag, nodeActive, metricUrl string) (error, string) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	var (
		url         string = "https://node-manager-ift-" + standNumber + ".apps.songd.sberdevices.ru/admin/node/add"
		jsonStrBody string = "{\"node\": \"" + nodeName + "\", \"active\": " + nodeActive + ", \"nodeTag\": \"" + nodeTag + "\", \"url\": \"" + nodeName + "\", \"metricUrl\": \"" + metricUrl + "\"}"
		jsonStr            = []byte(jsonStrBody)
	)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err, resp.Status
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	return nil, resp.Status
}

// Add node to stand.
func AddNodeToStand(stand string) error {
	nodeList := []string{"jazz-ift" + stand + "-bridge-s1.sberdevices.ru", "jazz-ift" + stand + "-bridge-s2.sberdevices.ru"}
	for _, node := range nodeList {
		fmt.Printf("Добавляем ноду %s к стенду IFT%s\n", node, stand)
		postAttempt := 0
		for {
			err, status := addNode(stand, node, "default", "1", metricUrl(stand))
			if err == nil && status == "200 OK" {
				break
			} else {
				postAttempt++
				duration := time.Duration(1) * time.Second
				time.Sleep(duration * time.Duration(postAttempt))
				if postAttempt > 5 {
					return err
				}
			}
		}
	}
	return nil
}
func main() {
	fmt.Println(ReturnNodeInfoMap(3, "ID"))
	fmt.Println(ReturnNodeInfoMap(3, "URL"))
	fmt.Println(ReturnNodeInfoMap(3, "Active"))
	fmt.Println(ReturnNodeInfoMap(3, "NodeTag"))
	fmt.Println(ReturnNodeInfoMap(3, "MetricUrl"))
}
