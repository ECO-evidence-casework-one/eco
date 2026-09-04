module github.com/ECO-evidence-casework-one/eco

go 1.23

require github.com/shirou/gopsutil/v4 v4.25.3

require (
	github.com/ebitengine/purego v0.8.2 // indirect
	github.com/emersion/go-mbox v0.0.0-20250604181414-1345da99f125
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/ledongthuc/pdf v0.0.0-20260903153007-b3c860c23753
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	golang.org/x/sys v0.28.0 // indirect
)

replace github.com/ledongthuc/pdf => ./third_party/ledongthuc_pdf
