TAG=0.5.9

build: extract
	go build

go-check:
	go vet
	staticcheck
	gosec .

fetch:
	curl -C - -LO https://github.com/logseq/logseq/releases/download/${TAG}/Logseq-linux-x64-${TAG}.zip
	curl -C - -LO https://github.com/logseq/logseq/releases/download/${TAG}/SHA256SUMS.txt

verify: fetch
	grep $$(sha256 Logseq-linux-x64-${TAG}.zip | awk '{print $$NF}') SHA256SUMS.txt

extract: verify
	unzip Logseq-linux-x64-${TAG}.zip

clean:
	rm -f SHA256SUMS.txt
	rm -f logseq-go
	rm -f *.zip
	rm -rf Logseq-linux-x64
