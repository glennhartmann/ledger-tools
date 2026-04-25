#!/bin/bash
fd -t d . 'src/' -x go generate -mod=readonly github.com/glennhartmann/ledger-tools/{}
