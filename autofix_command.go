package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func handleFixCommand(args []string) error {
	request, err := parseFixFlags(args)
	if err != nil {
		return err
	}

	result, runErr := runAutoFix(request)
	if runErr != nil {
		return runErr
	}

	printAutoFixResult(result)
	return nil
}

func parseFixFlags(args []string) (AutoFixRequest, error) {
	fixCmd := flag.NewFlagSet("fix", flag.ContinueOnError)
	fixCmd.SetOutput(os.Stderr)

	path := fixCmd.String("path", ".", "Path to apply safe fixes")
	apply := fixCmd.Bool("apply", false, "Apply safe changes to files (default: dry-run)")
	packs := fixCmd.String("packs", "size,imports,error-wrap", "Safe fix packs: size,imports,error-wrap")

	if err := fixCmd.Parse(args); err != nil {
		return AutoFixRequest{}, NewCLIError(
			ErrorCLIUsage,
			fmt.Sprintf("Invalid fix arguments: %v", err),
			"Usage: repodoctor fix -path . [--apply] [--packs size,imports,error-wrap]",
			err,
		)
	}

	request := AutoFixRequest{RepositoryPath: *path, Apply: *apply}
	request.Packs = parseFixPacks(*packs)
	if len(request.Packs) == 0 {
		return AutoFixRequest{}, NewCLIError(
			ErrorInvalidArgument,
			"No valid safe packs provided",
			"Use --packs size,imports,error-wrap",
			nil,
		)
	}

	return request, nil
}

func parseFixPacks(raw string) []AutoFixPack {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(raw)), ",")
	result := make([]AutoFixPack, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, AutoFixPack(trimmed))
	}
	return resolveAutoFixPacks(result)
}
