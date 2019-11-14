package main

import (
	"log"
	"os"
	"path/filepath"
	"time"
	//"github.com/pkg/profile"
)

const envWatchPath string = "EXPIRE_WATCH_PATH"

const envModifiedExp string = "EXPIRE_MODIFIED_EXPIRATION"
const defModifiedExp string = "1h" // 1 hour
const envCheckFreq string = "EXPIRE_CHECK_FREQUENCY"
const defCheckFreq string = "1m" // 1 minute

type ExpireConfig struct {
	watchPath          string
	modifiedExpiration time.Duration
	checkFrequency     time.Duration
}

func NewExpireConfig() ExpireConfig {
	conf := ExpireConfig{}

	// get watchPath from environment
	if conf.watchPath = os.Getenv(envWatchPath); len(conf.watchPath) == 0 {
		log.Fatalf("%v must be set\n", envWatchPath)
	}

	// get modifiedExpiration from environment, or default
	if envValModifiedExp := os.Getenv(envModifiedExp); len(envValModifiedExp) > 0 {
		var modParseError error
		if conf.modifiedExpiration, modParseError = time.ParseDuration(envValModifiedExp); modParseError != nil {
			log.Fatalf("%v duration must be in the format expected by %q (%v)\n",
				envModifiedExp, "time.ParseDuration()", "https://golang.org/pkg/time/#ParseDuration")
		} else if conf.modifiedExpiration < 0 {
			log.Fatalf("%v duration must be positive\n", envModifiedExp)
		}
	} else {
		conf.modifiedExpiration, _ = time.ParseDuration(defModifiedExp)
	}

	// get checkFrequency from environment, or default
	if envValCheckFreq := os.Getenv(envCheckFreq); len(envValCheckFreq) > 0 {
		var frqParseError error
		if conf.checkFrequency, frqParseError = time.ParseDuration(envValCheckFreq); frqParseError != nil {
			log.Fatalf("%v duration must be in the format expected by %q (%v)\n",
				envCheckFreq, "time.ParseDuration()", "https://golang.org/pkg/time/#ParseDuration")
		} else if conf.checkFrequency < 0 {
			log.Fatalf("%v duration must be positive\n", envCheckFreq)
		}
	} else {
		conf.checkFrequency, _ = time.ParseDuration(defCheckFreq)
	}

	return conf
}

func main() {
	// PROFILING
	//defer profile.Start().Stop()
	//defer profile.Start(profile.MemProfile).Stop()

	config := NewExpireConfig()

	// do work immediately, and then every $frequency
	PruneOldFiles(config.watchPath, config.modifiedExpiration)
	PruneEmptyDirs(config.watchPath)

	for _ = range time.NewTicker(config.checkFrequency).C {
		PruneOldFiles(config.watchPath, config.modifiedExpiration)
		PruneEmptyDirs(config.watchPath)
	}
}

func PruneOldFiles(root string, lastModified time.Duration) {
	// walk path and check files for deletion criteria
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("failed to access %q: %v\n", path, err)
			return err
		}

		if !info.IsDir() && time.Now().Sub(info.ModTime()) > lastModified {
			if err := os.Remove(path); err != nil {
				log.Printf("failed to prune %q: %v\n", path, err)
			} else {
				log.Printf("pruned %q (older than %v)", path, lastModified)
			}
		}

		return nil
	})
}

func PruneEmptyDirs(root string) {
	//log.Printf("pruning %q for empty dirs\n", root)
	var dirs []string

	// walk path and save any child directories
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		// ignore path access errors
		if err != nil {
			log.Printf("failed to access %q: %v\n", path, err)
			return nil
		}

		// record non-duplicate directory paths
		if info.IsDir() && path != root && !SliceContainsString(dirs, path) {
			dirs = append(dirs, path)
		}

		return nil
	})

	// attempt to remove all directories collected, suppressing errors
	// done in reverse, to start with the most-nested
	for index := range dirs {
		if err := os.Remove(dirs[len(dirs)-1-index]); err == nil {
			log.Printf("pruned %q (empty dir)\n", dirs[len(dirs)-1-index])
		}
	}
}

func SliceContainsString(slice []string, element string) bool {
	for index := range slice {
		if slice[len(slice)-1-index] == element {
			return true
		}
	}
	return false
}
