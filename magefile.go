// +build mage

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	goVersion   = "1.11"
	packageName = "github.com/brettpechiney/workout-service"
)

var (
	// Default target to run when none is passed specified.
	Default    = StartRoachContainer
	gocmd      = mg.GoCmd()
	releaseTag = regexp.MustCompile(`^v+[0-9]+\.[0-9]+\.[0-9]+$`)
)

// StartRoachContainer starts the CockroachDB container and kicks off it's
// migration scripts.
func StartRoachContainer() error {
	const MsgPrefix = "in StartRoachContainer"
	stopped := make(chan struct{})
	errchan := make(chan error)
	go func() {
		defer close(stopped)
		if err := sh.Run("cmd", "/C", "start", "docker-compose", "up", "cockroach"); err != nil {
			errchan <- fmt.Errorf("%s: %v", MsgPrefix, err)
		}
	}()
	for {
		select {
		case err := <-errchan:
			return err
		case <-stopped:
			return sh.Run("mage", "-d", "./migrations", "Migrate")
		}
	}
}

// Install runs go install and generates version information in the binary.
func Install() error {
	pkgslice := strings.Split(packageName, "/")
	name := pkgslice[len(pkgslice)-1]
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	gopath, err := sh.Output(gocmd, "env", "GOPATH")
	if err != nil {
		return fmt.Errorf("unable to detect GOPATH: %v", err)
	}
	paths := strings.Split(gopath, string([]rune{os.PathListSeparator}))
	bin := filepath.Join(paths[0], "bin")
	if err := os.Mkdir(bin, 0700); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create %q: %v", bin, err)
	}
	path := filepath.Join(bin, name)
	return sh.Run(gocmd, "build", "-o", path, "-ldflags="+flags(), packageName)
}

// Test runs all tests with the verbose flag set.
func Test() error {
	if !isGoLatest() {
		return fmt.Errorf("go %s.x must be installed", goVersion)
	}
	s, err := sh.Output(gocmd, "test", "./...")
	if err != nil {
		return err
	}
	fmt.Println(s)
	return nil
}

// TestRace runs all tests with both the race and verbose flags set.
func TestRace() error {
	s, err := sh.Output(gocmd, "test", "-v", "-race", "./...")
	if err != nil {
		return err
	}
	fmt.Println(s)
	return nil
}

// TestCover generates a test coverage report.
//noinspection GoUnusedExportedFunction
func TestCover() error {
	const (
		coverAll = "coverage-all.out"
		cover    = "coverage.out"
	)

	f, err := os.Create(coverAll)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write([]byte("mode: count")); err != nil {
		return err
	}
	pkgs, err := packages()
	if err != nil {
		return err
	}
	for _, pkg := range pkgs {
		if err := sh.Run(gocmd, "test", "-coverprofile="+cover, "-covermode=count", pkg); err != nil {
			return err
		}
		b, err := ioutil.ReadFile(cover)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		idx := bytes.Index(b, []byte{'\n'})
		b = b[idx+1:]
		if _, err := f.Write(b); err != nil {
			return err
		}
	}
	if err := f.Close(); err != nil {
		return err
	}
	return sh.Run(gocmd, "tool", "cover", "-html="+coverAll)
}

// Fmt runs the gofmt tool on all packages in the project.
func Fmt() error {
	if err := sh.Run(gocmd, "fmt", "./..."); err != nil {
		return fmt.Errorf("error running go fmt: %v", err)
	}
	return nil
}

// Lint runs the golint tool on all packages in the project.
//noinspection GoUnusedExportedFunction
func Lint() error {
	if err := sh.Run("golint", "./..."); err != nil {
		return fmt.Errorf("error running golint: %v", err)
	}
	return nil
}

// Vet runs the vet tool with all flags set.
func Vet() error {
	if err := sh.Run(gocmd, "vet", "-all", "./..."); err != nil {
		return fmt.Errorf("error running go vet: %v", err)
	}
	return nil
}

// Check makes sure the correct Golang version is being used, runs Fmt and Vet
// in parallel, and then runs TestRace.
func Check() error {
	if !isGoLatest() {
		return fmt.Errorf("go %s.x must be installed", goVersion)
	}
	mg.Deps(Fmt, Vet)
	mg.Deps(TestRace)
	return nil
}

// Release creates a new version of the service. It expects the TAG environment
// variable to be set because that is what it will use to create a new git tag.
// This target pushes changes to the Github repo.
func Release() (err error) {
	tag := strings.TrimSpace(os.Getenv("TAG"))
	if !releaseTag.MatchString(tag) {
		return fmt.Errorf("TAG should have format vx.x.x but is %s", tag)
	}

	trimmedTag := strings.TrimPrefix(tag, "v")
	msg := fmt.Sprintf("Release %s", trimmedTag)
	if err := sh.Run("git", "tag", "-a", tag, "-m", msg); err != nil {
		return fmt.Errorf("error adding git tag: %v", err)
	}
	if err := sh.Run("git", "push", "origin", tag); err != nil {
		return fmt.Errorf("error pushing to origin: %v", err)
	}
	defer func() {
		if err != nil {
			_ = sh.Run("git", "tag", "--delete", "$TAG")
			_ = sh.Run("git", "push", "--delete", "origin", "$TAG")
		}
	}()
	return nil
}

func isGoLatest() bool {
	return strings.Contains(runtime.Version(), goVersion)
}

func flags() string {
	timestamp := time.Now().Format(time.RFC3339)
	hash := hash()
	tag := tag()
	if tag == "" {
		tag = "dev"
	}
	ts := fmt.Sprintf("%s.timestamp=%s", packageName, timestamp)
	ghash := fmt.Sprintf("%s.commitHash=%s", packageName, hash)
	rtag := fmt.Sprintf("%s.gitTag=%s", packageName, tag)
	return fmt.Sprintf(`-X %s -X %s -X %s `, ts, ghash, rtag)
}

func tag() string {
	s, _ := sh.Output("git", "describe", "--tags")
	return s
}

func hash() string {
	hash, _ := sh.Output("git", "rev-parse", "--short", "HEAD")
	return hash
}

var (
	pkgPrefixLen = len(packageName)
	pkgs         []string
	pkgsInit     sync.Once
)

func packages() ([]string, error) {
	var err error
	pkgsInit.Do(func() {
		var s string
		s, err = sh.Output(gocmd, "list", "./...")
		if err != nil {
			return
		}
		pkgs = strings.Split(s, "\n")
		for i := range pkgs {
			pkgs[i] = "." + pkgs[i][pkgPrefixLen:]
		}
	})
	return pkgs, err
}
