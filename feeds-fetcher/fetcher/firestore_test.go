package fetcher

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/cskr/pubsub"
)

const FirestoreEmulatorHost = "FIRESTORE_EMULATOR_HOST"
const LoopbackIP = "127.0.0.1"
const FirestoreStdoutTopic = "firetore-logs"

func getFirestoreEmulatorCmd(verbose bool, port uint16) *exec.Cmd {
	var cmdArgs = []string{"beta", "emulators", "firestore", "start"}

	if !verbose {
		cmdArgs = append(cmdArgs, "--quiet")
	}
	if port == 0 {
		port = getFreeHostPort()
	}

	hostPort := fmt.Sprintf("--host-port=%s:%d", LoopbackIP, port)
	cmdArgs = append(cmdArgs, hostPort)

	return exec.Command("gcloud", cmdArgs...)
}

func setHostEnvIfIsConfigured(stdoutLine string) {
	pos := strings.Index(stdoutLine, FirestoreEmulatorHost+"=")

	if pos > 0 {
		host := stdoutLine[pos+len(FirestoreEmulatorHost)+1:]
		os.Setenv(FirestoreEmulatorHost, host)
	}
}

func publishFirestoreLogs(firestoreStdout io.ReadCloser, firestorePubSub *pubsub.PubSub) {
	streamReadlinesIterator, err := getStreamReadlinesIterator(firestoreStdout)
	if err != nil {
		log.Fatal(err)
	}

	for line := range streamReadlinesIterator {
		firestorePubSub.Pub(line, FirestoreStdoutTopic)
	}
}

func startFirestoreEmulator(verbose bool, port uint16) (cmd *exec.Cmd, stdout io.ReadCloser) {
	cmd = getFirestoreEmulatorCmd(verbose, port)
	stdout = getBothStdoutStderrCombined(cmd)

	makeProcessKillable(cmd)
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	return
}

func firestoreEmulatorIsReady(stdoutLine string) bool {
	return strings.Contains(stdoutLine, "Dev App Server is now running")
}

func waitForFirestoreToBeReady(ps *pubsub.PubSub) {
	channel := ps.Sub(FirestoreStdoutTopic)
	for {
		if msg, ok := <-channel; ok {

			if firestoreEmulatorIsReady(fmt.Sprintf("%s", msg)) {
				go ps.Unsub(channel, FirestoreStdoutTopic)
			}

			setHostEnvIfIsConfigured(fmt.Sprintf("%s", msg))
		} else {
			break
		}
	}
}

type FirestoreEmulator struct {
	pubSub  *pubsub.PubSub
	stdout  io.ReadCloser
	cmd     *exec.Cmd
	Verbose bool
	Port    uint16
}

func (f *FirestoreEmulator) Start() {
	f.pubSub = pubsub.New(0)
	f.cmd, f.stdout = startFirestoreEmulator(f.Verbose, f.Port)

	go publishFirestoreLogs(f.stdout, f.pubSub)
	go logPubSubTopic(f.pubSub, FirestoreStdoutTopic)
	waitForFirestoreToBeReady(f.pubSub)
}

func (f *FirestoreEmulator) Shutdown() {
	f.stdout.Close()
	ensureProcessWillBeKilled(f.cmd)
	f.pubSub.Shutdown()
}
