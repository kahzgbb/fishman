package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gookit/color"
)

const (
	checkWindow = 50 * time.Minute
	prefetchDir = `C:\Windows\Prefetch`
)

func main() {
	clearScreen()
	color.Red.Println("}<)))*>  <---\n")
	progress(0)

	var wg sync.WaitGroup
	var executed sync.Map
	now := time.Now()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for k, v := range getExecutionsFromPrefetch(now) {
			if isExe(k) {
				executed.Store(k, v)
			}
		}
		progress(10)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for k, v := range getShimCacheExecutions(now) {
			if isExe(k) {
				if _, exists := executed.Load(k); !exists {
					executed.Store(k, v)
				}
			}
		}
		progress(30)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for k, v := range getAmcacheExecutions(now) {
			if isExe(k) {
				if _, exists := executed.Load(k); !exists {
					executed.Store(k, v)
				}
			}
		}
		progress(60)
	}()

	wg.Wait()
	progress(85)

	now = time.Now()
	executed.Range(func(key, value any) bool {
		exe, execTime := key.(string), value.(time.Time)
		if !executableExists(exe) {
			minAgo := int(now.Sub(execTime).Minutes())
			color.Red.Printf("%s (Executado em %s) Deleted (%dm ago)\n",
				exe, execTime.Format("2006-01-02 15:04:05"), minAgo)
		}
		return true
	})

	progress(100)
	fmt.Println("\nExit")
	fmt.Scanln()
}

func getExecutionsFromPrefetch(now time.Time) map[string]time.Time {
	results := make(map[string]time.Time)
	files, err := os.ReadDir(prefetchDir)
	if err != nil {
		return results
	}
	for _, file := range files {
		if strings.HasSuffix(strings.ToLower(file.Name()), ".pf") {
			info, _ := file.Info()
			if now.Sub(info.ModTime()) <= checkWindow {
				name := extractExeName(file.Name())
				results[name] = info.ModTime()
			}
		}
	}
	return results
}

func extractExeName(pfName string) string {
	if parts := strings.Split(pfName, "-"); len(parts) > 0 {
		return parts[0]
	}
	return pfName
}

func getShimCacheExecutions(now time.Time) map[string]time.Time {
	results := make(map[string]time.Time)
	cmd := exec.Command("powershell", "-Command", `
		$reg = 'HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\AppCompatCache'
		$data = Get-ItemProperty -Path $reg -Name AppCompatCache -ErrorAction SilentlyContinue
		if ($data) {
			$data | Format-List
		}
	`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return results
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(strings.ToLower(line), ".exe") {
			exe := filepath.Base(strings.TrimSpace(line))
			results[exe] = now.Add(-40 * time.Minute)
		}
	}
	return results
}

func getAmcacheExecutions(now time.Time) map[string]time.Time {
	results := make(map[string]time.Time)
	cmd := exec.Command("powershell", "-Command", `
		$hive = 'C:\Windows\AppCompat\Programs\Amcache.hve'
		if (Test-Path $hive) {
			reg load HKLM\Amcache $hive | Out-Null
			$entries = Get-ChildItem 'HKLM:\Amcache\Root\File' -Recurse -ErrorAction SilentlyContinue | ForEach-Object {
				(Get-ItemProperty $_.PsPath).Name
			}
			reg unload HKLM\Amcache | Out-Null
			$entries
		}
	`)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return results
	}
	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(strings.ToLower(line), ".exe") {
			exe := filepath.Base(line)
			results[exe] = now.Add(-30 * time.Minute)
		}
	}
	return results
}

func executableExists(exeName string) bool {
	if !isExe(exeName) {
		return true
	}
	paths := []string{
		`C:\Program Files`,
		`C:\Program Files (x86)`,
		`C:\Windows\System32`,
		`C:\Users`,
	}
	var found bool
	var wg sync.WaitGroup
	ch := make(chan bool, len(paths))

	for _, base := range paths {
		wg.Add(1)
		go func(base string) {
			defer wg.Done()
			filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
				if err == nil && !d.IsDir() && strings.EqualFold(d.Name(), exeName) {
					ch <- true
					return filepath.SkipAll
				}
				return nil
			})
		}(base)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for v := range ch {
		if v {
			found = true
			break
		}
	}

	return found
}

func isExe(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	return strings.HasSuffix(name, ".exe") && len(name) > 4
}

func clearScreen() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

func progress(p int) {
	fmt.Printf("%d\n", p)
}
