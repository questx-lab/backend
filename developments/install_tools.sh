#!/bin/bash

install_tools() {
  go install github.com/spf13/cobra-cli@latest
  go install github.com/google/pprof@latest
  go install sigs.k8s.io/kind@v0.17.0
}

install_tools
