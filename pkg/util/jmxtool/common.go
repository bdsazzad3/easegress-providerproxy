package jmxtool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func JsonToKVMap(jsonStr string) (map[string]string, error) {

	m := map[string]interface{}{}
	err := json.Unmarshal([]byte(jsonStr), &m)
	if err != nil {
		return nil, err
	}
	kvs := extractKVs("", m)
	resultMap := map[string]string{}
	length := len(kvs)
	for i := 0; i < length; i++ {
		tmp := kvs[i]
		for k, v := range tmp {
			resultMap[k] = v
		}

	}
	return resultMap, nil
}

func extractKVs(prefix string, obj interface{}) []map[string]string {
	var rst []map[string]string

	switch obj.(type) {
	case map[string]interface{}:
		for k, v := range obj.(map[string]interface{}) {
			current := k
			rst = append(rst, extractKVs(join(prefix, current), v)...)
		}

	case []interface{}:
		o := obj.([]interface{})
		length := len(o)
		for i := 0; i < length; i++ {
			rst = append(rst, extractKVs(join(prefix, strconv.Itoa(i)), o[i])...)
		}
	case bool:
		rst = append(rst, map[string]string{prefix: strconv.FormatBool(obj.(bool))})
	case int:
		rst = append(rst, map[string]string{prefix: strconv.Itoa(obj.(int))})
	case float64:
		rst = append(rst, map[string]string{prefix: strconv.FormatFloat(obj.(float64), 'f', 0, 64)})
	default:
		if obj == nil {
			rst = append(rst, map[string]string{prefix: ""})
		} else {
			rst = append(rst, map[string]string{prefix: obj.(string)})
		}
	}
	return rst
}

func join(prefix string, current string) string {
	if prefix == "" {
		return current
	} else {
		return strings.Join([]string{prefix, current}, ".")
	}
}

type APIErr struct {
	Code    int    `yaml:"code"`
	Message string `yaml:"message"`
}

func handleRequest(httpMethod string, url string, reqBody []byte) (*http.Response, error) {
	req, err := http.NewRequest(httpMethod, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, err
	}

	if !successfulStatusCode(resp.StatusCode) {
		msg := string(body)
		apiErr := &APIErr{}
		err = yaml.Unmarshal(body, apiErr)
		if err == nil {
			msg = apiErr.Message
		}

		return resp, fmt.Errorf("Request failed: Code: %d, Msg: %s ", resp.StatusCode, msg)
	}

	return resp, nil
}

func successfulStatusCode(code int) bool {
	return code >= 200 && code < 300
}