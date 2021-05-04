module github.com/lmas/dumblog

go 1.16

require (
	github.com/alecthomas/chroma v0.9.1
	github.com/dvyukov/go-fuzz v0.0.0-20210429054444-fca39067bc72 // fuzzer complains without this one
	github.com/yuin/goldmark v1.3.5
	github.com/yuin/goldmark-highlighting v0.0.0-20210428103930-3a9678dbb86c
	gopkg.in/yaml.v2 v2.4.0
)
