package nacos

import "testing"

func TestConfig_Validate(t *testing.T) {
	cfg := Config{
		Address: "127.0.0.1", Port: 8848, Username: "u", Password: "p",
		NamespaceID: "ns", DataID: "d", Group: "g",
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if cfg.PollTime != DefaultPollTime {
		t.Errorf("PollTime = %v", cfg.PollTime)
	}
}

func TestConfig_Validate_MissingFields(t *testing.T) {
	var cfg Config
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestShouldSkip(t *testing.T) {
	t.Setenv(EnvSkip, "")
	if ShouldSkip("Y") != true {
		t.Error("flag Y should skip")
	}
	t.Setenv(EnvSkip, "Y")
	if ShouldSkip("N") != true {
		t.Error("env Y should skip")
	}
}
