/*
 * Copyright (C) 2020-2022, IrineSistiana
 *
 * This file is part of mosdns.
 *
 * mosdns is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * mosdns is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package tools

import (
	"fmt"
	"github.com/IrineSistiana/mosdns/v4/mlog"
	"github.com/IrineSistiana/mosdns/v4/pkg/matcher/domain"
	"github.com/IrineSistiana/mosdns/v4/pkg/matcher/netlist"
	"github.com/IrineSistiana/mosdns/v4/pkg/matcher/v2data"
	"github.com/spf13/cobra"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func newUnpackDomainCmd() *cobra.Command {
	var ourDir string
	c := &cobra.Command{
		Use:   "unpack-domain [-o output_dir] geosite.dat",
		Args:  cobra.ExactArgs(1),
		Short: "Unpack v2ray domain data file to text files.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := UnpackDomainDAT(args[0], ourDir); err != nil {
				mlog.S().Fatal(err)
			}
		},
	}
	c.PersistentFlags().StringVarP(&ourDir, "out", "o", "", "output dir")
	return c
}

func newUnpackIPCmd() *cobra.Command {
	var ourDir string
	c := &cobra.Command{
		Use:   "unpack-ip [-o output_dir] geoip.dat",
		Args:  cobra.ExactArgs(1),
		Short: "Unpack v2ray ip data file to text files.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := UnpackIPDAT(args[0], ourDir); err != nil {
				mlog.S().Fatal(err)
			}
		},
	}
	c.PersistentFlags().StringVarP(&ourDir, "out", "o", "", "output dir")
	return c
}

func UnpackDomainDAT(in, outDir string) error {
	b, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	geoSiteList, err := domain.LoadGeoSiteList(b)
	if err != nil {
		return err
	}

	for _, geoSite := range geoSiteList.GetEntry() {
		tag := strings.ToLower(geoSite.GetCountryCode())
		file := fmt.Sprintf("%s_%s.txt", fileName(in), tag)
		if len(outDir) > 0 {
			file = filepath.Join(outDir, file)
		}
		mlog.S().Infof("saving %s domain to %s", tag, file)
		err := convertV2DomainToTextFile(geoSite.GetDomain(), file)
		if err != nil {
			return err
		}
	}
	return nil
}

func fileName(f string) string {
	f = filepath.Base(f)
	if i := strings.LastIndexByte(f, '.'); i == -1 {
		return f
	} else {
		return f[:i]
	}
}

func convertV2DomainToTextFile(domain []*v2data.Domain, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	return convertV2DomainToText(domain, f)
}

func convertV2DomainToText(domain []*v2data.Domain, w io.Writer) error {
	for _, r := range domain {
		var prefix string
		switch r.Type {
		case v2data.Domain_Plain:
			prefix = "keyword:"
		case v2data.Domain_Regex:
			prefix = "regexp:"
		case v2data.Domain_Domain:
			prefix = ""
		case v2data.Domain_Full:
			prefix = "full:"
		default:
			return fmt.Errorf("invalid domain type %d", r.Type)
		}
		_, err := w.Write([]byte(prefix + r.Value + "\n"))
		if err != nil {
			return err
		}
	}
	return nil
}

func UnpackIPDAT(in, ourDir string) error {
	b, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	geoIPList, err := netlist.LoadGeoIPListFromDAT(b)
	if err != nil {
		return err
	}

	for _, ipList := range geoIPList.GetEntry() {
		tag := strings.ToLower(ipList.GetCountryCode())
		file := fmt.Sprintf("%s_%s.txt", fileName(in), tag)
		if len(ourDir) > 0 {
			file = filepath.Join(ourDir, file)
		}
		mlog.S().Infof("saving %s ip to %s", tag, file)
		err := convertV2CidrToTextFile(ipList.GetCidr(), file)
		if err != nil {
			return err
		}
	}

	return nil
}

func convertV2CidrToTextFile(cidr []*v2data.CIDR, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	return convertV2CidrToText(cidr, f)
}

func convertV2CidrToText(cidr []*v2data.CIDR, w io.Writer) error {
	for _, record := range cidr {
		n := net.IPNet{
			IP: record.Ip,
		}
		switch len(record.Ip) {
		case 4:
			n.Mask = net.CIDRMask(int(record.Prefix), 32)
		case 16:
			n.Mask = net.CIDRMask(int(record.Prefix), 128)
		}
		_, err := w.Write([]byte(n.String() + "\n"))
		if err != nil {
			return err
		}
	}
	return nil
}
