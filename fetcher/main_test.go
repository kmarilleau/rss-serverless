package fetcher

// var projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")

import (
	"github.com/cskr/pubsub"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	var testExitCode int
	firestorePubSub := pubsub.New(0)

	cmd, stdout := startFirestoreEmulator()
	defer stdout.Close()
	defer ensureProcessWillBeKilled(cmd)

	go publishFirestoreLogs(stdout, firestorePubSub)
	go logPubSubTopic(firestorePubSub, firestoreStdoutTopic)
	waitForFirestoreToBeReady(firestorePubSub)

	testExitCode = m.Run()

	firestorePubSub.Shutdown()
	os.Exit(testExitCode)
}
