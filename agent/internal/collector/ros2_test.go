package collector

import (
	"strings"
	"testing"
)

func TestROSCommandExportsEnvironment(t *testing.T) {
	command, err := rosCommand(nil, map[string]string{"ROS_LOCALHOST_ONLY": "1"}, "", "baize-ros2-subscriber --topic '/motor/state' --read-timeout '5s'")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(command, "export ROS_LOCALHOST_ONLY='1'") {
		t.Fatalf("ROS environment was not exported: %s", command)
	}
}

func TestROSCommandDropsRootToProfileUser(t *testing.T) {
	command, err := wrapROSCommand("exec baize-ros2-subscriber --topic '/motor/state'", "ubuntu", 0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(command, "/usr/bin/setpriv --reset-env") || !strings.Contains(command, "--reuid='ubuntu'") {
		t.Fatalf("ROS user transition was not configured: %s", command)
	}
}

func TestRequireROSUser(t *testing.T) {
	if err := requireROSUser("definitely-not-a-real-user-xyz", 0); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("missing ROS user must be rejected, got: %v", err)
	}
	if err := requireROSUser("root", 0); err != nil {
		t.Fatalf("existing ROS user must be accepted, got: %v", err)
	}
	if err := requireROSUser("definitely-not-a-real-user-xyz", 1000); err != nil {
		t.Fatalf("ROS user is irrelevant when not running as root, got: %v", err)
	}
	if err := requireROSUser("", 0); err != nil {
		t.Fatalf("empty ROS user must be accepted, got: %v", err)
	}
	if err := requireROSUser("Bad User!", 0); err == nil || !strings.Contains(err.Error(), "invalid ROS user") {
		t.Fatalf("malformed ROS user must be rejected, got: %v", err)
	}
}
