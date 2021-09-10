package tendermintlogger

import tmlog "github.com/interchained/genesismint/libs/log"

var _ tmlog.Logger = (*DiscardLogger)(nil)
