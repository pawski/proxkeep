package proxy

import (
	"crypto/md5"
)

type ResponseTest struct {
	statusCode   int
	responseHash [16]byte
}

func Prepare(httpStatusCode int, responseBody []byte) *ResponseTest {
	return &ResponseTest{statusCode: httpStatusCode, responseHash: md5.Sum(responseBody)}
}

func (r ResponseTest) Passed(httpStatusCode int, responseBody []byte) bool {
	return r.statusCode == httpStatusCode && r.responseHash == md5.Sum(responseBody)
}
