package proxy

import (
	"crypto/md5"
)

type ResponseTest struct {
	testURL      string
	statusCode   int
	responseHash [16]byte
}

func NewResponseTest(testURL string, httpStatusCode int, responseBody []byte) *ResponseTest {
	return &ResponseTest{testURL: testURL, statusCode: httpStatusCode, responseHash: md5.Sum(responseBody)}
}

func (r ResponseTest) Passed(httpStatusCode int, responseBody []byte) bool {
	return r.statusCode == httpStatusCode && r.responseHash == md5.Sum(responseBody)
}

func (r ResponseTest) GetTestURL() string {
	return r.testURL
}
