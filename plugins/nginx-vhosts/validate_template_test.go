package nginxvhosts

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gliderlabs/sigil"
	_ "github.com/gliderlabs/sigil/builtin"
)

func validateTemplatePath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "templates", "validate.conf.sigil")
}

func renderValidateTemplate(t *testing.T, nginxUser string) string {
	t.Helper()
	data, err := os.ReadFile(validateTemplatePath(t))
	if err != nil {
		t.Fatalf("read template: %v", err)
	}

	vars := map[string]interface{}{
		"NGINX_CONF": "/home/dokku/app/nginx.conf",
		"NGINX_USER": nginxUser,
	}
	buf, err := sigil.Execute(data, vars, "validate.conf.sigil")
	if err != nil {
		t.Fatalf("sigil.Execute: %v", err)
	}
	return buf.String()
}

func TestValidateTemplate_DeclaresNginxUser(t *testing.T) {
	out := renderValidateTemplate(t, "www-data")
	mustContain(t, out, "user www-data;")
	mustContain(t, out, "include /home/dokku/app/nginx.conf;")

	userIndex := strings.Index(out, "user www-data;")
	eventsIndex := strings.Index(out, "events {")
	if userIndex > eventsIndex {
		t.Errorf("expected the user directive before the events block\n--- output ---\n%s", out)
	}
}

func TestValidateTemplate_OmitsEmptyNginxUser(t *testing.T) {
	out := renderValidateTemplate(t, "")
	mustNotContain(t, out, "user ")
	mustContain(t, out, "include /home/dokku/app/nginx.conf;")
}
