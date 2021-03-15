module github.com/raidancampbell/audio-adblock

go 1.15

require (
	github.com/go-fingerprint/fingerprint v0.0.0-20140803133125-29397256b7ff
	github.com/go-fingerprint/gochroma v0.0.0-20171219195534-a19f54ab0cc7
	github.com/hajimehoshi/oto v0.7.1
	github.com/raidancampbell/gochroma v0.0.0-20210314173728-9bf21426b48e
	github.com/stretchr/testify v1.7.0
	github.com/tosone/minimp3 v0.0.0-20210225113649-a34145f13bef
	github.com/viert/go-lame v0.0.0-20201108052322-bb552596b11d
)

replace github.com/go-fingerprint/gochroma => github.com/raidancampbell/gochroma v0.0.0-20210315030441-506aba83182e
