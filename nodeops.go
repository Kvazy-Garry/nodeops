package nodeops

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	result NodeList
	nodes  []int //= make([]string, 0)
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
func GetNodesId(standNumber int) (error, []int) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	var (
		n         string = strconv.Itoa(standNumber)
		urlString string = "https://node-manager-ift-" + n + ".apps.songd.sberdevices.ru/admin/node/list"
	)
	nodes = make([]int, 0)
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
	for _, rec := range result {
		nodes = append(nodes, rec.ID)
	}
	return nil, nodes
}

func checkIfnodesNeedtoBeAdded(stand int) error {
	postAttempt := 0
	for {
		err, _ := GetNodesId(stand)
		if err == nil {
			if _, lenNodeList := GetNodesId(stand); len(lenNodeList) > 0 {
				fmt.Println("Ноды уже существуют")
				os.Exit(0)
			}
			break
		} else {
			postAttempt++
			duration := time.Duration(5) * time.Second
			time.Sleep(duration * time.Duration(postAttempt))
		}
		if postAttempt > 5 {
			return err
		}
	}
	return nil
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
	standPtr := flag.Int("standNumber", 100, "Stand number")
	flag.Parse()
	if err := checkIfnodesNeedtoBeAdded(*standPtr); err != nil {
		log.Fatal(err)
	}

	if err := AddNodeToStand(strconv.Itoa(*standPtr)); err != nil {
		log.Fatal(err)
	} else {
		os.Exit(0)
	}
}
