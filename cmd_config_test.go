package notes

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfigCmd(t *testing.T) {
	cfg := &Config{
		HomePath:  "/path/to/home",
		GitPath:   "/path/to/git",
		EditorCmd: "vim",
	}
	for _, tc := range []struct {
		name string
		want string
	}{
		{
			name: "",
			want: "HOME=/path/to/home\nGIT=/path/to/git\nEDITOR=vim\n",
		},
		{
			name: "home",
			want: "/path/to/home\n",
		},
		{
			name: "git",
			want: "/path/to/git\n",
		},
		{
			name: "editor",
			want: "vim\n",
		},
		{
			name: "HOME",
			want: "/path/to/home\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			c := ConfigCmd{
				Config: cfg,
				Name:   tc.name,
				Out:    &buf,
			}
			if err := c.Do(); err != nil {
				t.Fatal(err)
			}
			have := buf.String()
			if have != tc.want {
				t.Fatalf("want '%#v' but have '%#v'", tc.want, have)
			}
		})
	}
}

func TestConfigCmdError(t *testing.T) {
	cfg := &Config{
		HomePath:  "/path/to/home",
		GitPath:   "/path/to/git",
		EditorCmd: "vim",
	}
	c := ConfigCmd{
		Config: cfg,
		Name:   "unknown name",
	}
	err := c.Do()
	if err == nil {
		t.Fatal("Error did not occur")
	}
	if !strings.Contains(err.Error(), "Unknown config name") {
		t.Fatal("Unexpected error:", err)
	}
}
