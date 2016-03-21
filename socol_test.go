package socol

import (
  "testing"
  "os"
  "flag"
)

func TestMain(m *testing.M) {
  flag.Parse()
  os.Exit(m.Run())
}

func TestCollectStats(t *testing.T) {

}

func TestCanRunPlatform(t *testing.T) {

}
