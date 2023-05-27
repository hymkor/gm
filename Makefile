NAME=$(notdir $(abspath .))
EXE=$(shell go env EXE)
TARGET=$(NAME)$(EXE)

$(TARGET) : SKK-JISYO.L.bz2
	go fmt
	go build

SKK-JISYO.L.bz2:
	curl https://raw.githubusercontent.com/skk-dev/dict/master/SKK-JISYO.L | bzip2 > SKK-JISYO.L.bz2
