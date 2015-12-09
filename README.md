# surv-client-go

## Description

Surveillance (more or less ADS-B data but not raw) client.

We can consider this as a first approach at writing a WS-N (web services notifications -- see OASIS) push client.  It uses the [WS-N Go](https://github.com/keltia/wsn-go/) library.

Connects to the WS-N broker, subscribe to one or more topics and get the
notifications through the builtin http server.

# Build status

*Stable*
[![Build Status](https://secure.travis-ci.org/keltia/surv-client-go.png)](http://travis-ci.org/keltia/surv-client-go)

