/* Copyright 2021 Victor Penso

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>. */

package main

import (
        "io/ioutil"
        "os/exec"
        "log"
        "strings"
        "strconv"
        "github.com/prometheus/client_golang/prometheus"
)

func FairShareData() []byte {
        cmd := exec.Command( "sshare", "-n", "-P", "-o", "account,fairshare" )
        stdout, err := cmd.StdoutPipe()
        if err != nil {
                log.Fatal(err)
        }
        if err := cmd.Start(); err != nil {
                log.Fatal(err)
        }
        out, _ := ioutil.ReadAll(stdout)
        if err := cmd.Wait(); err != nil {
                log.Fatal(err)
        }
        return out
}

func FairTreeData() []byte {
        cmd := exec.Command( "sshare", "-n", "-P", "-o", "account,levelfs" )
        stdout, err := cmd.StdoutPipe()
        if err != nil {
                log.Fatal(err)
        }
        if err := cmd.Start(); err != nil {
                log.Fatal(err)
        }
        out, _ := ioutil.ReadAll(stdout)
        if err := cmd.Wait(); err != nil {
                log.Fatal(err)
        }
        return out
}

type FairShareMetrics struct {
        fairshare float64
}

type FairTreeMetrics struct {
        fairtree float64
	depth string
	parent string
}

func ParseFairShareMetrics() map[string]*FairShareMetrics {
        accounts := make(map[string]*FairShareMetrics)
        lines := strings.Split(string(FairShareData()), "\n")
        for _, line := range lines {
                if ! strings.HasPrefix(line,"  ") {
                        if strings.Contains(line,"|") {
                                account := strings.Trim(strings.Split(line,"|")[0]," ")
                                _,key := accounts[account]
                                if !key {
                                        accounts[account] = &FairShareMetrics{0}
                                }
                                fairshare,_ := strconv.ParseFloat(strings.Split(line,"|")[1],64)
                                accounts[account].fairshare = fairshare
                        }
                }
        }
        return accounts
}

func countLeadingSpaces(line string) string {
	return strconv.Itoa(len(line) - len(strings.TrimLeft(line, " ")))
}

func ParseFairTreeMetrics() map[string]*FairTreeMetrics {
        accounts := make(map[string]*FairTreeMetrics)
        lines := strings.Split(string(FairTreeData()), "\n")
	previous_depth := 0
	parents := []string{"root"}
	for _, line := range lines[1:] {
		if strings.Contains(line,"|") {
			account := strings.Trim(strings.Split(line,"|")[0]," ")
			_,key := accounts[account]
			if !key {
				accounts[account] = &FairTreeMetrics{0, "", ""}
			}
			fairtree,_ := strconv.ParseFloat(strings.Split(line,"|")[1],64)
			accounts[account].fairtree = fairtree
			accounts[account].depth = countLeadingSpaces(line)
			current_depth, _ := strconv.Atoi(accounts[account].depth)
			if previous_depth < current_depth {
				parents = append(parents, account)
			} else if previous_depth == current_depth {
				parents = parents[:len(parents)-1]
				parents = append(parents, account)
			} else {
				parents = parents[:len(parents)-1]
				parents = append(parents, account)
			}
			accounts[account].parent = parents[len(parents)-2]
			previous_depth, _ = strconv.Atoi(accounts[account].depth)
		}
        }
        return accounts
}

type FairShareCollector struct {
        fairshare *prometheus.Desc
}

type FairTreeCollector struct {
        fairtree *prometheus.Desc
}

func NewFairShareCollector() *FairShareCollector {
        labels := []string{"account"}
        return &FairShareCollector{
                fairshare: prometheus.NewDesc("slurm_account_fairshare","FairShare for account" , labels,nil),
        }
}

func NewFairTreeCollector() *FairTreeCollector {
        labels := []string{"account", "account_depth", "account_parent"}
        return &FairTreeCollector{
                fairtree: prometheus.NewDesc("slurm_account_fairtree","FairTree for account" , labels,nil),
        }
}

func (fsc *FairShareCollector) Describe(ch chan<- *prometheus.Desc) {
        ch <- fsc.fairshare
}

func (ftc *FairTreeCollector) Describe(ch chan<- *prometheus.Desc) {
        ch <- ftc.fairtree
}

func (fsc *FairShareCollector) Collect(ch chan<- prometheus.Metric) {
        fsm := ParseFairShareMetrics()
        for f := range fsm {
                ch <- prometheus.MustNewConstMetric(fsc.fairshare, prometheus.GaugeValue, fsm[f].fairshare, f)
        }
}

func (ftc *FairTreeCollector) Collect(ch chan<- prometheus.Metric) {
        ftm := ParseFairTreeMetrics()
        for f := range ftm {
                ch <- prometheus.MustNewConstMetric(ftc.fairtree, prometheus.GaugeValue, ftm[f].fairtree, f, ftm[f].depth, ftm[f].parent)
        }
}
