package artifactstore_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/dikaeinstein/artifactstore/artifactstore"
	"github.com/dikaeinstein/artifactstore/pkg/fsys"
	"github.com/dikaeinstein/artifactstore/pkg/store"
)

// The RoundTripFunc type is an adapter to allow the use of
// ordinary functions as  net/http.RoundTripper. If f is a function
// with the appropriate signature, RoundTripFunc(f) is a
// RoundTripper that calls f.
type roundTripFunc func(req *http.Request) *http.Response

// RoundTrip executes a single HTTP transaction, returning
// a Response for the provided Request.
func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func newTestClient(fn http.RoundTripper) *http.Client {
	return &http.Client{Transport: fn}
}

func TestDowloadArtifact(t *testing.T) {
	fakeRoundTripper := func(req *http.Request) *http.Response {
		testData := bytes.NewBufferString("This is test data")

		return &http.Response{
			StatusCode:    http.StatusOK,
			Body:          ioutil.NopCloser(testData),
			ContentLength: int64(len(testData.Bytes())),
		}
	}
	failingTestClient := newTestClient(roundTripFunc(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Status:     "404 Not Found",
			Body:       ioutil.NopCloser(bytes.NewBufferString("")),
		}
	}))
	testClient := newTestClient(roundTripFunc(fakeRoundTripper))

	artifactName := "mkl.zip"
	url := "intel.com/" + artifactName
	initializedMap := make(map[string]*artifactstore.Artifact)
	initializedMap[url] = &artifactstore.Artifact{
		Name:               artifactName,
		Url:                url,
		FileType:           "zip",
		Type:               "3rdparty",
		RetrievedFromCache: true,
	}

	testCases := []struct {
		desc      string
		client    *http.Client
		store     artifactstore.Store
		errMsg    string
		usedCache bool
		prefix    string
	}{
		{
			desc:      "Can Retrieve artifact from remote",
			client:    testClient,
			errMsg:    "",
			store:     store.NewInMemStore(make(map[string]*artifactstore.Artifact)),
			usedCache: false,
			prefix:    "3rdparty",
		},
		{
			desc:      "Can Retrieve artifact from cache",
			client:    testClient,
			errMsg:    "",
			store:     store.NewInMemStore(initializedMap),
			usedCache: true,
			prefix:    "3rdparty",
		},
		{
			desc:      "Returns an error when downloading artifact fails",
			client:    failingTestClient,
			errMsg:    "failed to download artifact[mkl.zip]: 404 Not Found",
			store:     store.NewInMemStore(make(map[string]*artifactstore.Artifact)),
			usedCache: false,
			prefix:    "3rdparty",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			afStore := artifactstore.New(tC.client, ".", fsys.NewInMemFS("."), tC.store)
			t.Cleanup(func() {
				if err := fsys.Cleanup("mkl.zip*"); err != nil {
					t.Log("failed to cleanup tmp Files")
				}
			})

			artifact, err := afStore.Get(context.Background(), "3rdparty", url)
			if err != nil && err.Error() != tC.errMsg {
				t.Fatal(err)
			}

			if err == nil && tC.usedCache != artifact.RetrievedFromCache {
				t.Errorf("RetrievedFromCache: want: %t, got: %t",
					tC.usedCache, artifact.RetrievedFromCache)
			}

			if err == nil && artifact.Name != artifactName {
				t.Errorf("artifact names not matching; want: %s, got: %s",
					artifactName, artifact.Name)
			}

			if err == nil && artifact.Type != tC.prefix {
				t.Errorf("artifact type not matching; want: %s, got: %s",
					tC.prefix, artifact.Type)
			}
		})
	}
}
