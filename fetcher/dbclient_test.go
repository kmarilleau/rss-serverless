package fetcher

import (
	"fmt"
	"github.com/cskr/pubsub"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

const FirestoreEmulatorHost = "FIRESTORE_EMULATOR_HOST"
const LoopbackIP = "127.0.0.1"
const firestoreStdoutTopic = "firetore-logs"

func getFirestoreEmulatorCmd() *exec.Cmd {
	port := getFreeHostPort()
	hostPort := fmt.Sprintf("--host-port=%s:%d", LoopbackIP, port)
	cmd := exec.Command("gcloud", "beta", "emulators", "firestore", "start", "--quiet", hostPort)

	return cmd
}

func setHostEnvIfIsConfigured(stdoutLine string) {
	pos := strings.Index(stdoutLine, FirestoreEmulatorHost+"=")

	if pos > 0 {
		host := stdoutLine[pos+len(FirestoreEmulatorHost)+1 : len(stdoutLine)-1]
		os.Setenv(FirestoreEmulatorHost, host)
	}
}

func publishFirestoreLogs(firestoreStdout io.ReadCloser, firestorePubSub *pubsub.PubSub) {
	streamReadlinesIterator, err := getStreamReadlinesIterator(firestoreStdout)
	if err != nil {
		log.Fatal(err)
	}

	for line := range streamReadlinesIterator {
		firestorePubSub.Pub(line, firestoreStdoutTopic)
	}
}

func startFirestoreEmulator() (cmd *exec.Cmd, stdout io.ReadCloser) {
	cmd = getFirestoreEmulatorCmd()
	stdout = getBothStdoutStderrCombined(cmd)

	makeProcessKillable(cmd)
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	return
}

func firestoreEmulatoreIsReady(stdoutLine string) bool {
	return strings.Contains(stdoutLine, "Dev App Server is now running")
}

func waitForFirestoreToBeReady(ps *pubsub.PubSub) {
	channel := ps.Sub(firestoreStdoutTopic)
	for {
		if msg, ok := <-channel; ok {

			if firestoreEmulatoreIsReady(fmt.Sprintf("%s", msg)) {
				go ps.Unsub(channel, firestoreStdoutTopic)
			}

			setHostEnvIfIsConfigured(fmt.Sprintf("%s", msg))
		} else {
			break
		}
	}
}
