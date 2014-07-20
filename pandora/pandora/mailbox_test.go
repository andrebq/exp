package pandora

import (
	"github.com/andrebq/exp/pandora"
	pandorahttp "github.com/andrebq/exp/pandora/http"
	"github.com/andrebq/exp/pandora/pgstore"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func mustCreateServer() *pandora.Server {
	bs, err := pgstore.OpenBlobStore("pandora", "pandora", "localhost", "pandora")
	if err != nil {
		panic(err)
	}

	ms, err := pgstore.OpenMessageStore("pandora", "pandora", "localhost", "pandora")
	if err != nil {
		panic(err)
	}
	if err := ms.DeleteMessages(); err != nil {
		panic(err)
	}
	return &pandora.Server{
		BlobStore:    bs,
		MessageStore: ms,
	}
}

func TestMailbox(t *testing.T) {
	server := mustCreateServer()
	handler := &pandorahttp.Handler{
		Server: server,
	}

	ts := httptest.NewServer(handler)
	defer ts.Close()

	// sending
	mb := Mailbox{
		Client:  http.DefaultClient,
		BaseUrl: ts.URL,
	}

	body := make(url.Values)
	body.Set("topic", "teste")
	mid, err := mb.Send("a@local", "b@remote", 0, body)
	if err != nil {
		t.Fatalf("error sending %v", err)
	}
	if len(mid) == 0 {
		t.Errorf("invalid mid")
	}

	fetched, err := mb.Fetch("b@remote", time.Minute*5)
	if err != nil {
		t.Fatalf("error fetching object")
	}
	if fetched.Get("mid") != mid {
		t.Errorf("expected mid %v got %v", fetched.Get("mid"), mid)
	}

	err = mb.Ack(fetched.Get("mid"), fetched.Get("lid"), Confirm)
	if err != nil {
		t.Errorf("error doing ACK. %v", err)
	}


	_, err = mb.Send("invalid@local", "b@remote", 0, body)
	if err != nil {
		t.Fatalf("error sending a message from invalid@local")
	}

	// now, I should fetch from invalid@local but
	_, err = mb.Fetch("invalid@local", time.Minute*5)
	if err != ErrNoData {
		t.Fatalf("should have received error %v but got %v", ErrNoData, err)
	}
}
