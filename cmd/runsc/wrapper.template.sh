#!/usr/bin/env bash

export TTRPC_ADDRESS=/run/containerd/containerd.sock.ttrpc
/usr/local/bin/containerd-shim-runsc-v1 $@
