package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/bwesterb/go-xentop"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	cpuTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "xen_cpu_time",
		Help: "Total CPU time of vServer",
	}, []string{"dom"})
	cpuFract = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "xen_cpu_fract",
		Help: "Fraction of CPU used by vServer",
	}, []string{"dom"})
	networkTx = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "xen_net_tx",
		Help: "Total transmitted network data",
	}, []string{"dom"})
	networkRx = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "xen_net_rx",
		Help: "Total received network data",
	}, []string{"dom"})
	diskROps = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "xen_disk_rops",
		Help: "Total number of disk read operators",
	}, []string{"dom"})
	diskWOps = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "xen_disk_wops",
		Help: "Total number of disk write operators",
	}, []string{"dom"})
	diskBOps = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "xen_disk_bops",
		Help: "Total number of blocked disk operators",
	}, []string{"dom"})
	diskSWritten = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "xen_disk_sectw",
		Help: "Total number of disk sectors written",
	}, []string{"dom"})
	diskSRead = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "xen_disk_sectr",
		Help: "Total number of disk sectors read",
	}, []string{"dom"})
)

func init() {
	prometheus.MustRegister(cpuTime)
	prometheus.MustRegister(cpuFract)
	prometheus.MustRegister(networkTx)
	prometheus.MustRegister(networkRx)
	prometheus.MustRegister(diskROps)
	prometheus.MustRegister(diskWOps)
	prometheus.MustRegister(diskBOps)
	prometheus.MustRegister(diskSWritten)
	prometheus.MustRegister(diskSRead)
}

func updateMetrics(line xentop.Line) {
	cpuTime.With(prometheus.Labels{"dom": line.Name}).Set(float64(line.CpuTime))
	cpuFract.With(prometheus.Labels{"dom": line.Name}).Set(float64(line.CpuFraction))
	networkTx.With(prometheus.Labels{"dom": line.Name}).Set(float64(line.NetworkTx))
	networkRx.With(prometheus.Labels{"dom": line.Name}).Set(float64(line.NetworkRx))
	diskROps.With(prometheus.Labels{"dom": line.Name}).Set(float64(line.DiskReadOps))
	diskWOps.With(prometheus.Labels{"dom": line.Name}).Set(float64(line.DiskWriteOps))
	diskBOps.With(prometheus.Labels{"dom": line.Name}).Set(float64(line.DiskBlockedIO))
	diskSWritten.With(prometheus.Labels{"dom": line.Name}).Set(float64(line.DiskSectorsWritten))
	diskSRead.With(prometheus.Labels{"dom": line.Name}).Set(float64(line.DiskSectorsRead))
}

func main() {
	addr := flag.String("bind", ":8080", "The address to bind to")
	cmd := flag.String("xentop", "xentop", "Path to xentop")
	flag.Parse()

	http.Handle("/metrics", promhttp.Handler())

	lines := make(chan xentop.Line, 2)
	errs := make(chan error, 2)

	go xentop.XenTopCmd(lines, errs, *cmd)
	go func() {
		for {
			select {
			case err := <-errs:
				log.Printf("%s\n", err)
			case line := <-lines:
				updateMetrics(line)
			}
		}
	}()

	log.Fatal(http.ListenAndServe(*addr, nil))
}
