package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
)

type profilingRequest struct {
	cpuProfilePath string
	memProfilePath string
}

func (r profilingRequest) enabled() bool {
	return strings.TrimSpace(r.cpuProfilePath) != "" || strings.TrimSpace(r.memProfilePath) != ""
}

func buildProfilingRequest(analyzeRoot, cpuPath, memPath string, watch bool) (profilingRequest, error) {
	request := profilingRequest{}
	if watch && (strings.TrimSpace(cpuPath) != "" || strings.TrimSpace(memPath) != "") {
		return request, NewCLIError(
			ErrorInvalidArgument,
			"Profiling cannot be enabled in watch mode",
			"Use analyze without -watch when collecting profiles",
			nil,
		)
	}

	var err error
	request.cpuProfilePath, err = sanitizeProfilePath(analyzeRoot, cpuPath)
	if err != nil {
		return profilingRequest{}, err
	}

	request.memProfilePath, err = sanitizeProfilePath(analyzeRoot, memPath)
	if err != nil {
		return profilingRequest{}, err
	}

	return request, nil
}

func sanitizeProfilePath(analyzeRoot, profilePath string) (string, error) {
	trimmed := strings.TrimSpace(profilePath)
	if trimmed == "" {
		return "", nil
	}

	root := filepath.Clean(analyzeRoot)
	if resolvedRoot, err := filepath.EvalSymlinks(root); err == nil {
		root = filepath.Clean(resolvedRoot)
	}

	target := filepath.Clean(trimmed)
	if !filepath.IsAbs(target) {
		target = filepath.Join(root, target)
	}

	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", HandleInvalidPathError(profilePath, err)
	}
	absTarget = filepath.Clean(absTarget)
	if resolvedTarget, err := filepath.EvalSymlinks(absTarget); err == nil {
		absTarget = filepath.Clean(resolvedTarget)
	}

	rel, err := filepath.Rel(root, absTarget)
	if err != nil {
		return "", NewCLIError(
			ErrorInvalidArgument,
			fmt.Sprintf("Invalid profile path: %s", profilePath),
			"Use a path under the analyze root directory",
			err,
		)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", NewCLIError(
			ErrorInvalidArgument,
			fmt.Sprintf("Profile path escapes analyze root: %s", profilePath),
			"Profile artifacts must stay under the analyze root directory",
			nil,
		)
	}

	return absTarget, nil
}

type profileSession struct {
	cpuFile *os.File
	memPath string
}

func startProfiling(request profilingRequest) (*profileSession, error) {
	if !request.enabled() {
		return nil, nil
	}

	session := &profileSession{memPath: request.memProfilePath}

	if request.cpuProfilePath != "" {
		if err := ensureProfileParentDir(request.cpuProfilePath); err != nil {
			return nil, err
		}
		cpuFile, err := os.Create(request.cpuProfilePath)
		if err != nil {
			return nil, err
		}
		if err := pprof.StartCPUProfile(cpuFile); err != nil {
			_ = cpuFile.Close()
			return nil, err
		}
		session.cpuFile = cpuFile
	}

	return session, nil
}

func (s *profileSession) Stop() error {
	if s == nil {
		return nil
	}

	if s.cpuFile != nil {
		pprof.StopCPUProfile()
		if err := s.cpuFile.Close(); err != nil {
			return err
		}
	}

	if s.memPath != "" {
		if err := ensureProfileParentDir(s.memPath); err != nil {
			return err
		}
		memFile, err := os.Create(s.memPath)
		if err != nil {
			return err
		}
		runtime.GC()
		if err := pprof.WriteHeapProfile(memFile); err != nil {
			_ = memFile.Close()
			return err
		}
		if err := memFile.Close(); err != nil {
			return err
		}
	}

	return nil
}

func ensureProfileParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "" || dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}
