//go:build mage
// +build mage

package main

import (
	"github.com/magefile/mage/sh"
)

func Bench() error {
	err := sh.Run("go", "test", "-bench=.", "-benchmem", "-trace=logs/trace.out", "-cpuprofile=logs/pprof.out")
	if err != nil {
		return err
	}

	return sh.Run("rm", "traq-wordcloud-bot.test")
}

func Trace() error {
	return sh.Run("go", "tool", "trace", "-http", "localhost:8080", "logs/trace.out")
}

func Pprof() error {
	return sh.Run("go", "tool", "pprof", "-http", "localhost:8080", "logs/pprof.out")
}
