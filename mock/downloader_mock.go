// Automatically generated by MockGen. DO NOT EDIT!
// Source: ../downloader.go

package mock

import (
	io "io"

	gomock "github.com/golang/mock/gomock"
)

// Mock of Downloader interface
type MockDownloader struct {
	ctrl     *gomock.Controller
	recorder *_MockDownloaderRecorder
}

// Recorder for MockDownloader (not exported)
type _MockDownloaderRecorder struct {
	mock *MockDownloader
}

func NewMockDownloader(ctrl *gomock.Controller) *MockDownloader {
	mock := &MockDownloader{ctrl: ctrl}
	mock.recorder = &_MockDownloaderRecorder{mock}
	return mock
}

func (_m *MockDownloader) EXPECT() *_MockDownloaderRecorder {
	return _m.recorder
}

func (_m *MockDownloader) DownloadFile(URL string) (io.ReadCloser, error) {
	ret := _m.ctrl.Call(_m, "DownloadFile", URL)
	ret0, _ := ret[0].(io.ReadCloser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockDownloaderRecorder) DownloadFile(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DownloadFile", arg0)
}

func (_m *MockDownloader) StoreFileToTemp(URL string) (string, error) {
	ret := _m.ctrl.Call(_m, "StoreFileToTemp", URL)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockDownloaderRecorder) StoreFileToTemp(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "StoreFileToTemp", arg0)
}
