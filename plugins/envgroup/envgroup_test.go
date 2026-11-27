package envgroup

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/dokku/dokku/plugins/common"
	. "github.com/onsi/gomega"
)

// setupTestEnv points the dokku environment at temporary directories and
// creates the given apps, marked as not deployed so no scheduler is queried
func setupTestEnv(t *testing.T, apps ...string) (dokkuRoot string, libRoot string) {
	t.Helper()

	libRoot = t.TempDir()
	dokkuRoot = t.TempDir()
	t.Setenv("DOKKU_LIB_ROOT", libRoot)
	t.Setenv("DOKKU_ROOT", dokkuRoot)
	t.Setenv("PLUGIN_PATH", filepath.Join(libRoot, "plugins"))
	t.Setenv("PLUGIN_ENABLED_PATH", filepath.Join(libRoot, "plugins", "enabled"))
	t.Setenv("PLUGIN_CORE_AVAILABLE_PATH", filepath.Join(libRoot, "core-plugins", "available"))
	t.Setenv("DOKKU_QUIET_OUTPUT", "1")

	current, err := user.Current()
	Expect(err).To(Succeed())
	group, err := user.LookupGroupId(current.Gid)
	Expect(err).To(Succeed())
	t.Setenv("DOKKU_SYSTEM_USER", current.Username)
	t.Setenv("DOKKU_SYSTEM_GROUP", group.Name)

	Expect(TriggerInstall()).To(Succeed())
	for _, appName := range apps {
		Expect(os.MkdirAll(filepath.Join(dokkuRoot, appName), 0755)).To(Succeed())
		Expect(common.PropertyWrite("common", appName, "deployed", "false")).To(Succeed())
	}

	return dokkuRoot, libRoot
}

func TestValidateGroupName(t *testing.T) {
	RegisterTestingT(t)

	Expect(validateGroupName("dev")).To(Succeed())
	Expect(validateGroupName("dev-1.test")).To(Succeed())
	Expect(validateGroupName("")).To(MatchError("Please specify an env group name"))
	Expect(validateGroupName("Dev")).To(HaveOccurred())
	Expect(validateGroupName("-dev")).To(HaveOccurred())
	Expect(validateGroupName("dev/test")).To(HaveOccurred())
	Expect(validateGroupName("../dev")).To(HaveOccurred())
}

func TestInsertAtIndex(t *testing.T) {
	RegisterTestingT(t)

	Expect(insertAtIndex([]string{}, "a", 0)).To(Equal([]string{"a"}))
	Expect(insertAtIndex([]string{"a", "b"}, "c", 0)).To(Equal([]string{"a", "b", "c"}))
	Expect(insertAtIndex([]string{"a", "b"}, "c", 1)).To(Equal([]string{"c", "a", "b"}))
	Expect(insertAtIndex([]string{"a", "b"}, "c", 2)).To(Equal([]string{"a", "c", "b"}))
	Expect(insertAtIndex([]string{"a", "b"}, "c", 3)).To(Equal([]string{"a", "b", "c"}))
	Expect(insertAtIndex([]string{"a", "b"}, "c", 10)).To(Equal([]string{"a", "b", "c"}))
}

func TestCreateDestroy(t *testing.T) {
	RegisterTestingT(t)
	setupTestEnv(t)

	Expect(CommandExists("dev")).To(MatchError("Env group dev does not exist"))
	Expect(CommandCreate("dev")).To(Succeed())
	Expect(CommandExists("dev")).To(Succeed())
	Expect(CommandCreate("dev")).To(MatchError("Env group dev already exists"))
	Expect(CommandCreate("Invalid")).To(HaveOccurred())
	Expect(CommandCreate("staging")).To(Succeed())

	groups, err := ListGroups()
	Expect(err).To(Succeed())
	Expect(groups).To(Equal([]string{"dev", "staging"}))

	Expect(CommandDestroy("dev", false, false)).To(Succeed())
	Expect(GroupExists("dev")).To(BeFalse())
	Expect(CommandDestroy("dev", false, false)).To(MatchError("Env group dev does not exist"))
}

func TestGroupVars(t *testing.T) {
	RegisterTestingT(t)
	setupTestEnv(t)

	Expect(CommandSet("dev", []string{"A=1"}, false, false)).To(MatchError("Env group dev does not exist"))
	Expect(CommandCreate("dev")).To(Succeed())
	Expect(CommandSet("dev", []string{}, false, false)).To(MatchError("At least one env pair must be given"))
	Expect(CommandSet("dev", []string{"A=1", "B=2"}, false, false)).To(Succeed())
	Expect(CommandSet("dev", []string{"C=aGVsbG8="}, true, false)).To(Succeed())
	Expect(CommandSet("dev", []string{"1BAD=x"}, false, false)).To(HaveOccurred())

	env, err := LoadGroupEnv("dev")
	Expect(err).To(Succeed())
	Expect(env.Map()).To(Equal(map[string]string{"A": "1", "B": "2", "C": "hello"}))

	Expect(CommandGet("dev", []string{"MISSING"}, false)).To(MatchError(ErrKeyNotSet))
	Expect(CommandGet("dev", []string{}, false)).To(MatchError("Expected: key"))

	Expect(CommandUnset("dev", []string{}, false)).To(MatchError("At least one key must be given"))
	Expect(CommandUnset("dev", []string{"A"}, false)).To(Succeed())
	env, _ = LoadGroupEnv("dev")
	Expect(env.Keys()).To(Equal([]string{"B", "C"}))

	Expect(CommandExport("dev", "yaml")).To(MatchError("Unknown export format: yaml"))

	Expect(CommandClear("dev", false)).To(Succeed())
	env, _ = LoadGroupEnv("dev")
	Expect(env.Len()).To(Equal(0))
}

func TestGroupImport(t *testing.T) {
	RegisterTestingT(t)
	_, libRoot := setupTestEnv(t)
	Expect(CommandCreate("dev")).To(Succeed())
	Expect(CommandSet("dev", []string{"EXISTING=1"}, false, false)).To(Succeed())

	envfile := filepath.Join(libRoot, "import.env")
	Expect(os.WriteFile(envfile, []byte("A=1\nB=2\n"), 0600)).To(Succeed())
	Expect(CommandImport("dev", envfile, "envfile", false, false)).To(Succeed())
	env, _ := LoadGroupEnv("dev")
	Expect(env.Map()).To(Equal(map[string]string{"A": "1", "B": "2", "EXISTING": "1"}))

	jsonfile := filepath.Join(libRoot, "import.json")
	Expect(os.WriteFile(jsonfile, []byte(`{"C": "3"}`), 0600)).To(Succeed())
	Expect(CommandImport("dev", jsonfile, "json", true, false)).To(Succeed())
	env, _ = LoadGroupEnv("dev")
	Expect(env.Map()).To(Equal(map[string]string{"C": "3"}))

	emptyfile := filepath.Join(libRoot, "empty.env")
	Expect(os.WriteFile(emptyfile, []byte(""), 0600)).To(Succeed())
	Expect(CommandImport("dev", emptyfile, "envfile", false, false)).To(MatchError("No config vars to set"))
	Expect(CommandImport("dev", envfile, "yaml", false, false)).To(MatchError("Unknown format: yaml"))
}

func TestMembership(t *testing.T) {
	RegisterTestingT(t)
	setupTestEnv(t, "app-a", "app-b")

	Expect(CommandAdd("dev", "app-a", 0, false)).To(MatchError("Env group dev does not exist"))
	Expect(CommandCreate("dev")).To(Succeed())
	Expect(CommandCreate("staging")).To(Succeed())
	Expect(CommandCreate("shared")).To(Succeed())

	Expect(CommandAdd("dev", "missing-app", 0, false)).To(HaveOccurred())
	Expect(CommandAdd("dev", "app-a", -1, false)).To(MatchError("The index must be a positive number"))
	Expect(CommandAdd("dev", "app-a", 0, false)).To(Succeed())
	Expect(CommandAdd("dev", "app-a", 0, false)).To(MatchError("App app-a is already a member of env group dev"))
	Expect(CommandAdd("staging", "app-a", 0, false)).To(Succeed())
	Expect(CommandAdd("shared", "app-a", 1, false)).To(Succeed())
	Expect(CommandAdd("dev", "app-b", 0, false)).To(Succeed())

	groups, err := AppGroups("app-a")
	Expect(err).To(Succeed())
	Expect(groups).To(Equal([]string{"shared", "dev", "staging"}))
	Expect(reportGroups("app-a")).To(Equal("shared dev staging"))

	members, err := GroupMembers("dev")
	Expect(err).To(Succeed())
	Expect(members).To(Equal([]string{"app-a", "app-b"}))

	Expect(CommandRemove("dev", "app-a", false)).To(Succeed())
	Expect(CommandRemove("dev", "app-a", false)).To(MatchError("App app-a is not a member of env group dev"))
	groups, _ = AppGroups("app-a")
	Expect(groups).To(Equal([]string{"shared", "staging"}))
}

func TestMergedAppGroupEnv(t *testing.T) {
	RegisterTestingT(t)
	setupTestEnv(t, "app-a", "app-b")

	Expect(CommandCreate("first")).To(Succeed())
	Expect(CommandCreate("second")).To(Succeed())
	Expect(CommandSet("first", []string{"SHARED=first", "FIRST=1"}, false, false)).To(Succeed())
	Expect(CommandSet("second", []string{"SHARED=second", "SECOND=2"}, false, false)).To(Succeed())

	env, err := MergedAppGroupEnv("app-a")
	Expect(err).To(Succeed())
	Expect(env.Len()).To(Equal(0))

	Expect(CommandAdd("first", "app-a", 0, false)).To(Succeed())
	Expect(CommandAdd("second", "app-a", 0, false)).To(Succeed())
	env, err = MergedAppGroupEnv("app-a")
	Expect(err).To(Succeed())
	Expect(env.Map()).To(Equal(map[string]string{"SHARED": "second", "FIRST": "1", "SECOND": "2"}))

	Expect(CommandAdd("second", "app-b", 0, false)).To(Succeed())
	Expect(CommandAdd("first", "app-b", 0, false)).To(Succeed())
	env, err = MergedAppGroupEnv("app-b")
	Expect(err).To(Succeed())
	Expect(env.Map()).To(Equal(map[string]string{"SHARED": "first", "FIRST": "1", "SECOND": "2"}))
}

func TestDestroyWithMembers(t *testing.T) {
	RegisterTestingT(t)
	setupTestEnv(t, "app-a", "app-b")

	Expect(CommandCreate("dev")).To(Succeed())
	Expect(CommandCreate("staging")).To(Succeed())
	Expect(CommandSet("dev", []string{"A=1"}, false, false)).To(Succeed())
	Expect(CommandAdd("dev", "app-a", 0, false)).To(Succeed())
	Expect(CommandAdd("staging", "app-a", 0, false)).To(Succeed())
	Expect(CommandAdd("dev", "app-b", 0, false)).To(Succeed())

	Expect(CommandDestroy("dev", false, false)).To(MatchError("Env group dev is in use by the following apps: app-a, app-b. Pass --force to remove it from them and destroy it"))
	Expect(GroupExists("dev")).To(BeTrue())

	Expect(CommandDestroy("dev", true, false)).To(Succeed())
	Expect(GroupExists("dev")).To(BeFalse())
	Expect(common.DirectoryExists(groupDirectory("dev"))).To(BeFalse())

	groups, _ := AppGroups("app-a")
	Expect(groups).To(Equal([]string{"staging"}))
	groups, _ = AppGroups("app-b")
	Expect(groups).To(BeEmpty())
}

func TestAppLifecycle(t *testing.T) {
	RegisterTestingT(t)
	dokkuRoot, _ := setupTestEnv(t, "app-a")

	Expect(CommandCreate("dev")).To(Succeed())
	Expect(CommandAdd("dev", "app-a", 0, false)).To(Succeed())

	Expect(TriggerPostAppCloneSetup("app-a", "app-clone")).To(Succeed())
	groups, _ := AppGroups("app-clone")
	Expect(groups).To(Equal([]string{"dev"}))

	Expect(TriggerPostAppRenameSetup("app-a", "app-renamed")).To(Succeed())
	groups, _ = AppGroups("app-renamed")
	Expect(groups).To(Equal([]string{"dev"}))
	groups, _ = AppGroups("app-a")
	Expect(groups).To(BeEmpty())

	Expect(os.MkdirAll(filepath.Join(dokkuRoot, "app-renamed"), 0755)).To(Succeed())
	Expect(TriggerPostDelete("app-renamed")).To(Succeed())
	groups, _ = AppGroups("app-renamed")
	Expect(groups).To(BeEmpty())
}

func TestBackupRoundTrip(t *testing.T) {
	RegisterTestingT(t)
	_, libRoot := setupTestEnv(t, "app-a")

	Expect(CommandCreate("dev")).To(Succeed())
	Expect(CommandCreate("empty")).To(Succeed())
	Expect(CommandCreate("staging")).To(Succeed())
	Expect(CommandSet("dev", []string{"A=1", "B=two words"}, false, false)).To(Succeed())
	Expect(CommandAdd("staging", "app-a", 0, false)).To(Succeed())
	Expect(CommandAdd("dev", "app-a", 0, false)).To(Succeed())

	globalScope := filepath.Join(libRoot, "backup", "global")
	appScope := filepath.Join(libRoot, "backup", "apps", "app-a")
	Expect(TriggerBackupGlobalExport(globalScope)).To(Succeed())
	Expect(TriggerBackupAppExport("app-a", appScope)).To(Succeed())

	Expect(CommandDestroy("dev", true, false)).To(Succeed())
	Expect(CommandDestroy("empty", true, false)).To(Succeed())
	Expect(CommandDestroy("staging", true, false)).To(Succeed())
	Expect(CommandCreate("dev")).To(Succeed())
	Expect(CommandSet("dev", []string{"STALE=1"}, false, false)).To(Succeed())

	Expect(TriggerBackupGlobalImport(globalScope)).To(Succeed())
	Expect(TriggerBackupAppImport("app-a", appScope)).To(Succeed())

	groups, err := ListGroups()
	Expect(err).To(Succeed())
	Expect(groups).To(Equal([]string{"dev", "empty", "staging"}))

	env, _ := LoadGroupEnv("dev")
	Expect(env.Map()).To(Equal(map[string]string{"A": "1", "B": "two words"}))

	appGroups, _ := AppGroups("app-a")
	Expect(appGroups).To(Equal([]string{"staging", "dev"}))

	Expect(CommandDestroy("staging", true, false)).To(Succeed())
	Expect(TriggerBackupAppImport("app-a", appScope)).To(Succeed())
	appGroups, _ = AppGroups("app-a")
	Expect(appGroups).To(Equal([]string{"dev"}))
}

func TestBackupWithoutGroups(t *testing.T) {
	RegisterTestingT(t)
	_, libRoot := setupTestEnv(t, "app-a")

	globalScope := filepath.Join(libRoot, "backup", "global")
	appScope := filepath.Join(libRoot, "backup", "apps", "app-a")
	Expect(TriggerBackupGlobalExport(globalScope)).To(Succeed())
	Expect(TriggerBackupAppExport("app-a", appScope)).To(Succeed())
	Expect(common.FileExists(common.BackupRecipePath(pluginName, globalScope))).To(BeFalse())
	Expect(common.FileExists(common.BackupRecipePath(pluginName, appScope))).To(BeFalse())
	Expect(TriggerBackupGlobalImport(globalScope)).To(Succeed())
	Expect(TriggerBackupAppImport("app-a", appScope)).To(Succeed())
}
